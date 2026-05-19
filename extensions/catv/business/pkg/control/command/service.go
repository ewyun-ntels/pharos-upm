package command

import (
	"context"
	"log/slog"
	"time"

	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/pools/workers"
	control_common "ntels.com/pharos/extensions/catv/business/pkg/control/common"
	"ntels.com/pharos/extensions/catv/business/pkg/control/common/tables"
)

func NewService(config common.Config, k8sJobName string) (*Service, error) {
	stbControlScheduleResultTable := tables.NewStbControlScheduleResultTable(config)

	stbControlScheduleResultRaws, err := stbControlScheduleResultTable.SelectWhereK8sJobName(k8sJobName)
	if err != nil {
		return nil, err
	}

	ctx, cancel := context.WithCancel(context.Background())

	return &Service{
		config:                        config,
		k8sJobName:                    k8sJobName,
		stbControlScheduleResultRaws:  stbControlScheduleResultRaws,
		ctx:                           ctx,
		cancel:                        cancel,
		stbControlScheduleResultTable: stbControlScheduleResultTable,
	}, nil
}

type Service struct {
	config                        common.Config
	k8sJobName                    string
	stbControlScheduleResultRaws  []tables.StbControlScheduleResultRaw
	ctx                           context.Context
	cancel                        context.CancelFunc
	stbControlScheduleResultTable tables.StbControlScheduleResultTable

	workerPool     *workers.WorkerPool
	batchSubmitter *workers.BatchSubmitter
	rateLimiter    *workers.RateLimiter
	resultWriter   *AsyncBatchWriter
	resultChannel  chan<- *BatchWriteRequest
}

func (s *Service) Start() error {
	startTime := time.Now()
	slog.Info("Starting schedule", "k8s_job_name", s.k8sJobName, "schedule_result_count", len(s.stbControlScheduleResultRaws), "start_time", startTime.Format(time.RFC3339))

	if err := s.load(); err != nil {
		return err
	}

	if len(s.stbControlScheduleResultRaws) == 0 {
		slog.Warn("No schedule to process", "k8s_job_name", s.k8sJobName)
		return nil
	}

	for index := range s.stbControlScheduleResultRaws {
		select {
		case <-s.ctx.Done():
			slog.Warn("Context cancelled, stopping schedule submission", "k8s_job_name", s.k8sJobName)
			return s.ctx.Err()
		default:
		}

		if err := s.rateLimiter.Wait(s.ctx); err != nil {
			slog.Warn("Rate limiter wait failed, cancelling schedule submission", "error", err)
			return err
		}

		workerJob := workers.Job{
			ID:      s.stbControlScheduleResultRaws[index].ID,
			Type:    JobTypeStbControl,
			Payload: &s.stbControlScheduleResultRaws[index],
		}

		if err := s.batchSubmitter.Add(workerJob); err != nil {
			slog.Error("Failed to submit schedule to worker pool", "job_id", workerJob.ID, "error", err)
			s.stbControlScheduleResultRaws[index].ResultCode = control_common.ResponseCodeError
			s.stbControlScheduleResultRaws[index].ResultMessage = err.Error()
			s.writeResult(&s.stbControlScheduleResultRaws[index])

			continue
		}
	}

	if err := s.batchSubmitter.Flush(); err != nil {
		slog.Error("Failed to flush batch submitter", "k8s_job_name", s.k8sJobName, "error", err)
		return err
	}

	slog.Info("All controller schedules submitted", "k8s_job_name", s.k8sJobName, "duration_seconds", time.Since(startTime).Seconds())

	return nil
}

func (s *Service) Stop() error {
	slog.Info("Stopping control service", "k8s_job_name", s.k8sJobName)

	// Step 1: Cancel context to signal all operations to stop
	if s.cancel != nil {
		s.cancel()
	}

	// Step 2: Stop accepting new jobs
	if s.rateLimiter != nil {
		s.rateLimiter.Close()
	}

	// Step 3: Stop worker pool and wait for all jobs to complete
	// This ensures no more data will be sent to resultChannel
	if s.workerPool != nil {
		s.workerPool.Stop()
	}

	// Step 4: Stop AsyncBatchWriter (safe now that workers are stopped)
	// This will flush all remaining data in the channel
	if s.resultWriter != nil {
		s.resultWriter.Stop()
	}

	slog.Info("Control service stopped", "k8s_job_name", s.k8sJobName)

	return nil
}

func (s *Service) load() error {
	// Rate Limiter: Prevent port exhaustion based on TIME_WAIT (60s)
	// Formula: max_ports(20,000) ÷ TIME_WAIT(60s) = 333/s (theoretical max)
	// Safety margin: 300/s (prevents reaching system limits)
	// NOTE: This limits connection START rate, not completion rate
	//
	// Example with 5s processing time:
	//   - Second 0: 300 jobs start (300 active, 0 in TIME_WAIT)
	//   - Second 1: 300 jobs start (600 active, 0 in TIME_WAIT)
	//   - Second 5: 300 jobs start, first 300 complete (1,500 active, 300 in TIME_WAIT)
	//   - Steady state: 1,500 active + 18,000 TIME_WAIT = 19,500 ports (safe)
	//
	// WARNING: Each Service instance creates its own rate limiter.
	// Multiple concurrent services will multiply the actual connection rate.
	// Total rate = maxConnectionsPerSecond × number of active services
	//
	// Connection pooling NOT used: Each device has unique IP (0% reuse rate)
	// Direct net.DialTimeout eliminates pool overhead (~50-100μs per connection)
	maxConnectionsPerSecond := s.config.Catv.Control.MaxConnectionsPerSecond
	if maxConnectionsPerSecond <= 0 {
		maxConnectionsPerSecond = 300 // Default: safe value for port exhaustion prevention
	}
	s.rateLimiter = workers.NewRateLimiter(maxConnectionsPerSecond)

	channelSize := max(s.config.Catv.Control.Batch.MaxCount*2, 1000)
	batchSize := s.config.Catv.Control.Batch.MaxCount
	flushInterval, err := time.ParseDuration(s.config.Catv.Control.Batch.FlushInterval)
	if err != nil {
		slog.Error("invalid flush interval duration, using default 5s", "input", s.config.Catv.Control.Batch.FlushInterval, "error", err)
		flushInterval = 5 * time.Second
	}
	s.resultWriter = NewAsyncBatchWriter(s.config, channelSize, batchSize, flushInterval)
	s.resultChannel = s.resultWriter.GetChannel()
	s.resultWriter.Start()

	jobFail := func(job workers.Job) {
		raw, ok := job.Payload.(*tables.StbControlScheduleResultRaw)
		if !ok {
			slog.Error("failed to assert job payload to *tables.StbControlScheduleResultRaw")
			return
		}

		raw.ResultCode = control_common.ResponseCodeError
		s.writeResult(raw)
	}

	s.workerPool = workers.NewWorkerPool(workers.WorkerPoolOptions{
		WorkerCount:   s.config.Pool.Workers.WorkerCount,
		JobQueueSize:  s.config.Pool.Workers.JobQueueSize,
		EnableMetrics: s.config.Pool.Workers.EnableMetrics,
		OnError: func(job workers.Job, err error) {
			slog.Error("worker pool handler error", "error", err)

			jobFail(job)
		},
		OnPanic: func(job workers.Job, recovered any) {
			slog.Error("worker pool handler panic recovered", "recovered", recovered)

			jobFail(job)
		},
	})

	if err := s.workerPool.RegisterHandler(NewStbControlHandler(s.config, s)); err != nil {
		slog.Error("failed to register worker pool handler", "error", err)
		return err
	}
	s.batchSubmitter = workers.NewBatchSubmitter(s.workerPool, s.config.Pool.Workers.BatchSize)

	s.workerPool.Start()

	return nil
}

func (s *Service) writeResult(raw *tables.StbControlScheduleResultRaw) {
	select {
	case s.resultChannel <- &BatchWriteRequest{StbControlScheduleResultRaw: raw}:
	case <-time.After(5 * time.Second):
		slog.Error("Result channel blocked, falling back to direct insert", "result_id", raw.ID, "progress_status", raw.ProgressStatus)
		if err := s.stbControlScheduleResultTable.Inserts([]tables.StbControlScheduleResultRaw{*raw}); err != nil {
			slog.Error("Direct insert failed", "result_id", raw.ID, "error", err)
		}
	}
}
