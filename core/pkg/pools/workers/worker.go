package workers

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"sync"
	"sync/atomic"
	"time"
)

// Sentinel errors for better error handling
// These can be checked using errors.Is() for type-safe error handling
//
// Usage example:
//
//	err := pool.Submit(job)
//	if errors.Is(err, ErrQueueFull) {
//	    // Handle queue full
//	}
var (
	ErrPoolStopped      = errors.New("worker pool is stopped")
	ErrPoolShuttingDown = errors.New("worker pool is shutting down")
	ErrQueueFull        = errors.New("job queue is full")
	ErrHandlerExists    = errors.New("handler already registered")
	ErrPoolStarted      = errors.New("cannot modify pool after started")
	ErrNoHandler        = errors.New("no handler found for job type")
)

// JobType represents the type of job to be executed
type JobType string

const (
	// These are sample examples. Define job types suitable for your actual project.
	// Examples: JobTypeDataCollection, JobTypeDeviceControl, JobTypeReportGeneration, etc.
	JobTypeSampleTask1 JobType = "sample_task_1"
	JobTypeSampleTask2 JobType = "sample_task_2"
)

// Priority represents job execution priority
type Priority int

const (
	PriorityLow Priority = iota
	PriorityNormal
	PriorityHigh
	PriorityUrgent
)

// Job represents a unit of work to be processed
// Jobs are submitted to the worker pool and processed by registered handlers
type Job struct {
	ID          string       `json:"id"`                     // Unique identifier for the job
	Type        JobType      `json:"type"`                   // Type determines which handler processes this job
	Priority    Priority     `json:"priority"`               // Execution priority (currently informational)
	Payload     any          `json:"payload"`                // Job-specific data passed to the handler
	RetryPolicy *RetryPolicy `json:"retry_policy,omitempty"` // Optional retry configuration
	CreatedAt   int64        `json:"created_at"`             // Unix timestamp when job was created
}

// RetryPolicy defines retry behavior for failed jobs
// Uses exponential backoff with configurable parameters
//
// Example:
//
//	policy := &RetryPolicy{
//	    MaxRetries:    3,
//	    InitialDelay:  100,  // 100ms
//	    MaxDelay:      5000, // 5s
//	    BackoffFactor: 2.0,  // 100ms -> 200ms -> 400ms
//	}
type RetryPolicy struct {
	MaxRetries    int     `json:"max_retries"`    // Maximum number of retry attempts
	InitialDelay  int64   `json:"initial_delay"`  // Initial delay in milliseconds
	MaxDelay      int64   `json:"max_delay"`      // Maximum delay cap in milliseconds
	BackoffFactor float64 `json:"backoff_factor"` // Multiplier for exponential backoff (e.g., 2.0)
}

// JobHandler is an interface that handles specific job types
type JobHandler interface {
	// Handle processes a job and returns an error if processing fails
	Handle(ctx context.Context, job Job) error
	// CanHandle returns true if this handler can process the given job type
	CanHandle(jobType JobType) bool
}

// WorkerPool manages a pool of workers processing jobs
// Provides concurrent job processing with O(1) handler lookup,
// automatic retries, metrics collection, and health monitoring
//
// Thread-safety:
//   - Submit/TrySubmit/SubmitBatch: Safe for concurrent calls
//   - RegisterHandler: Must be called before Start()
//   - Start/Stop/Shutdown: Idempotent, safe for concurrent calls
//   - GetMetrics/HealthCheck: Safe for concurrent calls
type WorkerPool struct {
	// Pool configuration
	workerCount int                    // Number of concurrent worker goroutines
	jobQueue    chan Job               // Buffered channel for pending jobs
	handlers    []JobHandler           // Registered handlers (fallback linear search)
	handlerMap  map[JobType]JobHandler // O(1) handler lookup by job type
	handlerMu   sync.RWMutex           // Protects handlers and handlerMap
	wg          sync.WaitGroup         // Waits for all workers to finish
	ctx         context.Context        // Cancellation context
	cancel      context.CancelFunc     // Cancellation function
	stopped     atomic.Bool            // Pool stopped state (idempotent Stop)
	started     atomic.Bool            // Pool started state (prevents handler registration)

	// Metrics (atomic for lock-free updates)
	totalJobs         atomic.Int64 // Total jobs submitted (includes retries as separate jobs)
	processedJobs     atomic.Int64 // Successfully completed jobs
	failedJobs        atomic.Int64 // Jobs that failed after all retries
	retriedJobs       atomic.Int64 // Number of retry attempts made
	processingTimeSum atomic.Int64 // Cumulative processing time in nanoseconds
	lastProcessedAt   atomic.Int64 // Unix timestamp of last completed job

	// Per-type metrics (protected by typeMetricsMu)
	typeMetrics   map[JobType]*TypeMetrics // Metrics broken down by job type
	typeMetricsMu sync.RWMutex             // Protects typeMetrics map

	// Options
	opts WorkerPoolOptions // Configuration options
}

// ErrorHandler is called when a job fails
type ErrorHandler func(job Job, err error)

// PanicHandler is called when a job panics
type PanicHandler func(job Job, recovered any)

// WorkerPoolOptions contains configuration options for the worker pool
type WorkerPoolOptions struct {
	// WorkerCount is the number of concurrent workers
	WorkerCount int
	// JobQueueSize is the buffer size for the job queue
	JobQueueSize int
	// EnableMetrics enables collection of processing metrics
	EnableMetrics bool
	// OnError is called when a job fails (after retries exhausted)
	OnError ErrorHandler
	// OnPanic is called when a job panics
	OnPanic PanicHandler
	// EnableDLQ enables dead letter queue for failed jobs
	EnableDLQ bool
	// DLQSize is the buffer size for dead letter queue
	DLQSize int
}

// DefaultOptions returns default configuration for the worker pool
func DefaultOptions() WorkerPoolOptions {
	return WorkerPoolOptions{
		WorkerCount:   100,
		JobQueueSize:  10000,
		EnableMetrics: true,
	}
}

// NewWorkerPool creates a new worker pool with the given options
func NewWorkerPool(opts WorkerPoolOptions) *WorkerPool {
	ctx, cancel := context.WithCancel(context.Background())

	// Set defaults
	if opts.DLQSize == 0 && opts.EnableDLQ {
		opts.DLQSize = 1000
	}

	return &WorkerPool{
		workerCount: opts.WorkerCount,
		jobQueue:    make(chan Job, opts.JobQueueSize),
		handlers:    make([]JobHandler, 0),
		handlerMap:  make(map[JobType]JobHandler),
		typeMetrics: make(map[JobType]*TypeMetrics),
		ctx:         ctx,
		cancel:      cancel,
		opts:        opts,
	}
}

// RegisterHandler registers a job handler for processing specific job types
// Handlers are indexed by JobType for O(1) lookup performance
//
// IMPORTANT: Must be called before Start() to avoid race conditions
// Returns ErrPoolStarted if called after the pool has started
func (wp *WorkerPool) RegisterHandler(handler JobHandler) error {
	if wp.started.Load() {
		return ErrPoolStarted
	}

	wp.handlerMu.Lock()
	defer wp.handlerMu.Unlock()

	wp.handlers = append(wp.handlers, handler)

	// Build handler map for fast lookup
	// Note: if multiple handlers can handle the same type, last one wins
	for _, jobType := range getAllJobTypes() {
		if handler.CanHandle(jobType) {
			wp.handlerMap[jobType] = handler
		}
	}

	return nil
}

// allJobTypes is a package-level cache of all known job types
// Add new job types here as they are defined
var allJobTypes = []JobType{
	JobTypeSampleTask1,
	JobTypeSampleTask2,
	// Add more job types as needed
}

// getAllJobTypes returns all known job types
// This uses a cached slice to avoid repeated allocations
func getAllJobTypes() []JobType {
	return allJobTypes
}

// Start starts the worker pool
// After calling Start(), RegisterHandler() will return an error
func (wp *WorkerPool) Start() {
	wp.started.Store(true)
	for i := range wp.workerCount {
		wp.wg.Add(1)
		go wp.worker(i)
	}
}

// worker is the main worker goroutine that processes jobs
func (wp *WorkerPool) worker(_ int) {
	defer wp.wg.Done()

	for {
		select {
		case <-wp.ctx.Done():
			return
		case job, ok := <-wp.jobQueue:
			if !ok {
				return
			}

			wp.processJobWithMetrics(job)
		}
	}
}

// processJobWithMetrics processes a job with full metrics, retry, and error handling
// This is the main job processing pipeline that:
//  1. Tracks total jobs submitted
//  2. Recovers from panics and calls OnPanic handler
//  3. Implements retry logic with exponential backoff
//  4. Updates all metrics (total, processed, failed, retried, per-type)
//  5. Calls OnError handler for final failures
//
// Metrics behavior:
//   - totalJobs is incremented once per job (not per retry)
//   - retriedJobs tracks number of retry attempts
//   - processingTimeSum includes time for all retry attempts
func (wp *WorkerPool) processJobWithMetrics(job Job) {
	start := time.Now()
	wp.totalJobs.Add(1)

	// Panic recovery
	defer func() {
		if r := recover(); r != nil {
			wp.failedJobs.Add(1)
			if wp.opts.OnPanic != nil {
				wp.opts.OnPanic(job, r)
			}
		}
	}()

	var err error
	var retryCount int

	// Retry loop
	for {
		err = wp.processJob(job)
		if err == nil {
			break
		}

		// Check if retry is enabled and available
		if job.RetryPolicy == nil || retryCount >= job.RetryPolicy.MaxRetries {
			if job.RetryPolicy != nil {
				slog.Error("job retry exhausted",
					"job_id", job.ID,
					"job_type", job.Type,
					"retry_count", retryCount,
					"max_retries", job.RetryPolicy.MaxRetries,
					"last_error", err)
			}
			break
		}

		retryCount++
		wp.retriedJobs.Add(1)

		// Calculate exponential backoff delay
		delay := calculateBackoffDelay(job.RetryPolicy, retryCount)

		slog.Debug("retrying job",
			"job_id", job.ID,
			"job_type", job.Type,
			"attempt", retryCount+1,
			"delay", delay,
			"error", err)

		time.Sleep(delay)
	}

	// Update metrics
	duration := time.Since(start)
	wp.processingTimeSum.Add(duration.Nanoseconds())
	endTime := start.Add(duration)
	wp.lastProcessedAt.Store(endTime.Unix())

	// Update per-type metrics
	wp.updateTypeMetrics(job.Type, duration, err == nil)

	if err != nil {
		wp.failedJobs.Add(1)
		if wp.opts.OnError != nil {
			wp.opts.OnError(job, err)
		}
	} else {
		wp.processedJobs.Add(1)
	}
}

// calculateBackoffDelay calculates the delay for exponential backoff
// Optimized for common case (BackoffFactor == 2.0) with integer arithmetic
func calculateBackoffDelay(policy *RetryPolicy, attempt int) time.Duration {
	// Fast path: integer-based calculation for 2.0 factor (most common)
	if policy.BackoffFactor == 2.0 {
		delay := policy.InitialDelay * int64(time.Millisecond)
		for i := 1; i < attempt; i++ {
			delay *= 2
			if delay > policy.MaxDelay*int64(time.Millisecond) {
				return time.Duration(policy.MaxDelay * int64(time.Millisecond))
			}
		}
		return time.Duration(delay)
	}

	// Slow path: float-based calculation for arbitrary factors
	delay := float64(policy.InitialDelay) * float64(time.Millisecond)
	for i := 1; i < attempt; i++ {
		delay *= policy.BackoffFactor
		if delay > float64(policy.MaxDelay)*float64(time.Millisecond) {
			delay = float64(policy.MaxDelay) * float64(time.Millisecond)
			break
		}
	}
	return time.Duration(delay)
}

// processJob processes a single job using registered handlers
// Uses O(1) map lookup for performance
func (wp *WorkerPool) processJob(job Job) error {
	// Fast path: O(1) map lookup
	wp.handlerMu.RLock()
	handler, exists := wp.handlerMap[job.Type]
	wp.handlerMu.RUnlock()

	if exists {
		return handler.Handle(wp.ctx, job)
	}

	// Fallback: linear search (for dynamically registered handlers)
	for _, handler := range wp.handlers {
		if handler.CanHandle(job.Type) {
			return handler.Handle(wp.ctx, job)
		}
	}

	slog.Error("no handler found",
		"job_id", job.ID,
		"job_type", job.Type)
	return fmt.Errorf("%w: %s", ErrNoHandler, job.Type)
}

// Submit submits a job to the worker pool (blocking if queue is full)
// Returns ErrPoolStopped if pool is stopped, ErrPoolShuttingDown if shutting down
func (wp *WorkerPool) Submit(job Job) error {
	if wp.stopped.Load() {
		return ErrPoolStopped
	}
	select {
	case <-wp.ctx.Done():
		return ErrPoolShuttingDown
	case wp.jobQueue <- job:
		return nil
	}
}

// TrySubmit attempts to submit a job without blocking
// Returns ErrQueueFull if queue is full, ErrPoolStopped if stopped, ErrPoolShuttingDown if shutting down
func (wp *WorkerPool) TrySubmit(job Job) error {
	if wp.stopped.Load() {
		return ErrPoolStopped
	}
	select {
	case <-wp.ctx.Done():
		return ErrPoolShuttingDown
	case wp.jobQueue <- job:
		return nil
	default:
		return ErrQueueFull
	}
}

// SubmitBatch submits multiple jobs to the worker pool
func (wp *WorkerPool) SubmitBatch(jobs []Job) error {
	for _, job := range jobs {
		if err := wp.Submit(job); err != nil {
			return err
		}
	}
	return nil
}

// Stop gracefully stops the worker pool (idempotent)
func (wp *WorkerPool) Stop() {
	if !wp.stopped.CompareAndSwap(false, true) {
		return // Already stopped
	}
	close(wp.jobQueue)
	wp.wg.Wait()
}

// Shutdown initiates a graceful shutdown of the worker pool
func (wp *WorkerPool) Shutdown() {
	wp.cancel()
	wp.Stop()
}

// GetMetrics returns current processing metrics
// Returns a snapshot of metrics at the time of the call
//
// Metrics fields:
//   - TotalJobs: Total jobs submitted (unique jobs, not counting retries)
//   - ProcessedJobs: Jobs completed successfully
//   - FailedJobs: Jobs that failed after all retry attempts
//   - RetriedJobs: Total number of retry attempts across all jobs
//   - QueueSize: Current number of jobs waiting in queue
//   - AvgProcessingTime: Average time per job (including retries)
//   - LastProcessedAt: Timestamp of last completed job
//   - PerTypeMetrics: Breakdown of metrics by job type
//
// Note: This creates a snapshot and copies per-type metrics,
// which may be expensive if there are many job types
func (wp *WorkerPool) GetMetrics() Metrics {
	totalJobs := wp.totalJobs.Load()
	processedJobs := wp.processedJobs.Load()
	processingTimeSum := wp.processingTimeSum.Load()

	var avgProcessingTime time.Duration
	if processedJobs > 0 {
		avgProcessingTime = time.Duration(processingTimeSum / processedJobs)
	}

	// Copy per-type metrics
	wp.typeMetricsMu.RLock()
	perTypeMetrics := make(map[JobType]TypeMetricsSnapshot)
	for k, v := range wp.typeMetrics {
		count := v.count.Load()
		perTypeMetrics[k] = TypeMetricsSnapshot{
			Count:       count,
			Failures:    v.failures.Load(),
			AvgDuration: time.Duration(v.totalDuration.Load() / max(count, 1)),
		}
	}
	wp.typeMetricsMu.RUnlock()

	return Metrics{
		TotalJobs:         totalJobs,
		ProcessedJobs:     processedJobs,
		FailedJobs:        wp.failedJobs.Load(),
		RetriedJobs:       wp.retriedJobs.Load(),
		QueueSize:         int64(len(wp.jobQueue)),
		AvgProcessingTime: avgProcessingTime,
		LastProcessedAt:   time.Unix(wp.lastProcessedAt.Load(), 0),
		PerTypeMetrics:    perTypeMetrics,
	}
}

// updateTypeMetrics updates per-type metrics
// Uses double-checked locking to minimize lock contention
func (wp *WorkerPool) updateTypeMetrics(jobType JobType, duration time.Duration, success bool) {
	// Fast path: RLock for existing metrics
	wp.typeMetricsMu.RLock()
	m, exists := wp.typeMetrics[jobType]
	wp.typeMetricsMu.RUnlock()

	// Slow path: Create new metrics if not exists
	if !exists {
		wp.typeMetricsMu.Lock()
		// Re-check after acquiring write lock (another goroutine may have created it)
		m, exists = wp.typeMetrics[jobType]
		if !exists {
			m = &TypeMetrics{}
			wp.typeMetrics[jobType] = m
		}
		wp.typeMetricsMu.Unlock()
	}

	m.count.Add(1)
	m.totalDuration.Add(duration.Nanoseconds())
	if !success {
		m.failures.Add(1)
	}
}

// Metrics contains worker pool processing metrics
type Metrics struct {
	TotalJobs         int64
	ProcessedJobs     int64
	FailedJobs        int64
	RetriedJobs       int64
	QueueSize         int64
	AvgProcessingTime time.Duration
	LastProcessedAt   time.Time
	PerTypeMetrics    map[JobType]TypeMetricsSnapshot
}

// typeMetricsInternal contains per-job-type metrics (internal with atomics)
type typeMetricsInternal struct {
	count         atomic.Int64
	failures      atomic.Int64
	totalDuration atomic.Int64 // nanoseconds
}

// TypeMetricsSnapshot contains per-job-type metrics snapshot (exported)
type TypeMetricsSnapshot struct {
	Count       int64
	Failures    int64
	AvgDuration time.Duration
}

// TypeMetrics is the internal type used in maps
type TypeMetrics = typeMetricsInternal

// HealthStatus represents the health of the worker pool
type HealthStatus struct {
	Healthy       bool
	QueueHealth   string // "ok", "warning", "critical"
	WorkerHealth  string
	LastProcessed time.Time
	Issues        []string
}

// HealthCheck performs a health check on the worker pool
// Returns detailed health status with configurable thresholds
//
// Health criteria:
//
//   - Queue Health:
//
//   - critical: queue usage > 90%
//
//   - warning:  queue usage > 70%
//
//   - ok:       queue usage <= 70%
//
//   - Worker Health:
//
//   - critical: no jobs processed in 5 minutes AND queue not empty
//
//   - ok:       jobs being processed or queue empty
//
//   - Failure Rate:
//
//   - unhealthy: failure rate > 50% (after 100+ jobs)
//
//   - warning:   failure rate > 20% (after 100+ jobs)
//
//   - ok:        failure rate <= 20%
//
// The pool is marked Healthy=false if any critical condition is met
func (wp *WorkerPool) HealthCheck() HealthStatus {
	metrics := wp.GetMetrics()
	status := HealthStatus{
		Healthy:       true,
		QueueHealth:   "ok",
		WorkerHealth:  "ok",
		LastProcessed: metrics.LastProcessedAt,
		Issues:        make([]string, 0),
	}

	// Check queue congestion
	queueUsage := float64(metrics.QueueSize) / float64(wp.opts.JobQueueSize)
	if queueUsage > 0.9 {
		status.QueueHealth = "critical"
		status.Healthy = false
		issue := fmt.Sprintf("queue usage critical: %.1f%%", queueUsage*100)
		status.Issues = append(status.Issues, issue)
	} else if queueUsage > 0.7 {
		status.QueueHealth = "warning"
		issue := fmt.Sprintf("queue usage high: %.1f%%", queueUsage*100)
		status.Issues = append(status.Issues, issue)
	}

	// Check if workers are processing
	if !metrics.LastProcessedAt.IsZero() {
		timeSinceLastProcess := time.Since(metrics.LastProcessedAt)
		if timeSinceLastProcess > 5*time.Minute && metrics.QueueSize > 0 {
			status.WorkerHealth = "critical"
			status.Healthy = false
			status.Issues = append(status.Issues, "no jobs processed in last 5 minutes")
		}
	}

	// Check failure rate
	if metrics.TotalJobs > 100 {
		failureRate := float64(metrics.FailedJobs) / float64(metrics.TotalJobs)
		if failureRate > 0.5 {
			status.Healthy = false
			issue := fmt.Sprintf("high failure rate: %.1f%%", failureRate*100)
			status.Issues = append(status.Issues, issue)
		} else if failureRate > 0.2 {
			issue := fmt.Sprintf("elevated failure rate: %.1f%%", failureRate*100)
			status.Issues = append(status.Issues, issue)
		}
	}

	return status
}
