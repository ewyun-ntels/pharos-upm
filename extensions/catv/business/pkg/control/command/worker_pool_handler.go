package command

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
	"time"

	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/pools/workers"
	control_common "ntels.com/pharos/extensions/catv/business/pkg/control/common"
	"ntels.com/pharos/extensions/catv/business/pkg/control/common/tables"
)

// JobTypeStbControl is the job type identifier for settopbox control operations.
// This constant is used by the worker pool to route jobs to the appropriate handler.
const JobTypeStbControl workers.JobType = "stb-control"

// responseBufferPool provides a pool of pre-allocated byte slices for TCP responses.
// This reduces GC pressure by reusing buffers across requests instead of allocating new ones.
var responseBufferPool = sync.Pool{
	New: func() any {
		// Allocate 8KB buffer (enough for most responses)
		b := make([]byte, 8192)
		return &b
	},
}

// StbControlHandler implements the JobHandler interface for processing
// settopbox control operations. It handles TCP communication with settopbox devices
// and stores results to the database.
//
// The handler performs the following operations:
//  1. Validates job payload (request_received_at, stb_mac, ip, stb_control_request)
//  2. Executes the control command via direct TCP connection
//  3. Stores the result (success or error) to the database
//
// This handler is registered with the worker pool during API.Load() and processes
// all jobs with type JobTypeStbControl.
//
// Note: Connection pooling is NOT used because each device has a unique IP address,
// resulting in 0% connection reuse. Direct net.DialContext eliminates unnecessary
// overhead (hash calculation, lock contention, map operations, singleflight).
//
// Dependencies:
//   - config: provides controller port, timeout, and database configuration
//   - service: provides result channel for batch writes
type StbControlHandler struct {
	config              common.Config                        // System configuration for controller and database
	service             *Service                             // Service instance for accessing result channel
	scheduleResultTable tables.StbControlScheduleResultTable // Pre-created for fallback direct inserts
}

func NewStbControlHandler(config common.Config, service *Service) *StbControlHandler {
	return &StbControlHandler{
		config:              config,
		service:             service,
		scheduleResultTable: tables.NewStbControlScheduleResultTable(config),
	}
}

func (h *StbControlHandler) Handle(ctx context.Context, job workers.Job) error {
	slog.Debug("Received job for settopbox control", "job_id", job.ID, "job_type", job.Type)

	select {
	case <-ctx.Done():
		slog.Warn("Context cancelled before starting job execution", "job_id", job.ID)
		return ctx.Err()
	default:
	}

	scheduleResultRaw, ok := job.Payload.(*tables.StbControlScheduleResultRaw)
	if !ok {
		return fmt.Errorf("invalid job payload: expected *tables.StbControlScheduleResultRaw, got %T", job.Payload)
	}
	slog.Debug("Job payload received", "job_id", job.ID, "payload", scheduleResultRaw)

	scheduleResultRaw.ControlStartTime = &orm.Datetime{Time: time.Now().UTC()}
	scheduleResultRaw.ProgressStatus = tables.ProgressStatusRunning
	if err := h.insertResult(ctx, *scheduleResultRaw); err != nil {
		return fmt.Errorf("failed to update job status to running: %w", err)
	}

	response := NewSettopbox(h.config, ctx, *scheduleResultRaw).Control(true)

	scheduleResultRaw.ControlEndTime = &orm.Datetime{Time: time.Now().UTC()}
	scheduleResultRaw.ResultCode = response.Code
	scheduleResultRaw.ResultMessage = response.Message
	switch scheduleResultRaw.ResultCode {
	case control_common.ResponseCodeSuccess:
		scheduleResultRaw.ProgressStatus = tables.ProgressStatusSucceeded
	case control_common.ResponseCodeFailure, control_common.ResponseCodeError:
		scheduleResultRaw.ProgressStatus = tables.ProgressStatusFailed
	}
	if err := h.insertResult(context.Background(), *scheduleResultRaw); err != nil {
		return fmt.Errorf("failed to record control result: %w", err)
	}

	slog.Info("Completed settopbox control job",
		"job_id", job.ID,
		"work_type", scheduleResultRaw.WorkType,
		"stb_mdl_nm", scheduleResultRaw.StbMdlNm,
		"cm_mac_addr", scheduleResultRaw.CmMacAddr,
		"stb_mac_addr", scheduleResultRaw.StbMacAddr,
		"result_code", scheduleResultRaw.ResultCode,
		"result_message", scheduleResultRaw.ResultMessage,
		"duration_seconds", scheduleResultRaw.ControlEndTime.Time.Sub(scheduleResultRaw.ControlStartTime.Time).Seconds(),
	)

	return nil
}

func (h *StbControlHandler) CanHandle(jobType workers.JobType) bool {
	return jobType == JobTypeStbControl
}

func (h *StbControlHandler) insertResult(ctx context.Context, raw tables.StbControlScheduleResultRaw) error {
	if h.service != nil {
		select {
		case h.service.resultChannel <- &BatchWriteRequest{StbControlScheduleResultRaw: &raw}:
			return nil
		case <-time.After(10 * time.Second):
			slog.Error("Result channel blocked, falling back to direct insert", "result_id", raw.ID, "progress_status", raw.ProgressStatus)
		case <-ctx.Done():
			slog.Warn("Context cancelled, persisting result via direct insert", "result_id", raw.ID, "progress_status", raw.ProgressStatus)
		}
	}
	return h.scheduleResultTable.Inserts([]tables.StbControlScheduleResultRaw{raw})
}
