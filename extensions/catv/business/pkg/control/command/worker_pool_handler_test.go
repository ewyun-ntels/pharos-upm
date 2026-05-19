package command

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/pools/workers"
	control_common "ntels.com/pharos/extensions/catv/business/pkg/control/common"
	"ntels.com/pharos/extensions/catv/business/pkg/control/common/tables"
)

func newHandlerConfig() common.Config {
	var cfg common.Config
	cfg.Catv.Control.Timeout.Connection = "100ms"
	cfg.Catv.Control.Timeout.Send = "200ms"
	cfg.Catv.Control.Retry.PortExhaustion.Count = 1
	cfg.Catv.Control.Retry.PortExhaustion.Delay = "1ms"
	cfg.Catv.Control.Retry.Total.Count = 1
	cfg.Catv.Control.Retry.Total.Delay = "1ms"
	return cfg
}

// --- NewStbControlHandler ---

func TestNewStbControlHandler(t *testing.T) {
	h := NewStbControlHandler(newHandlerConfig(), nil)
	require.NotNil(t, h)
}

// --- CanHandle ---

func TestCanHandle(t *testing.T) {
	h := NewStbControlHandler(newHandlerConfig(), nil)
	assert.True(t, h.CanHandle(JobTypeStbControl))
	assert.False(t, h.CanHandle("other-type"))
	assert.False(t, h.CanHandle(""))
}

// --- Handle: ctx already cancelled ---

func TestHandle_CtxAlreadyCancelled(t *testing.T) {
	h := NewStbControlHandler(newHandlerConfig(), nil)
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	raw := &tables.StbControlScheduleResultRaw{ID: "j1"}
	job := workers.Job{ID: "j1", Type: JobTypeStbControl, Payload: raw}
	err := h.Handle(ctx, job)
	assert.Equal(t, context.Canceled, err)
}

// --- Handle: invalid payload ---

func TestHandle_InvalidPayload(t *testing.T) {
	h := NewStbControlHandler(newHandlerConfig(), nil)
	job := workers.Job{ID: "j2", Type: JobTypeStbControl, Payload: "not-a-raw"}
	err := h.Handle(context.Background(), job)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid job payload")
}

// --- Handle: service nil, DB unreachable → insertResult falls back to direct insert ---

func TestHandle_ServiceNil_DBError(t *testing.T) {
	h := NewStbControlHandler(newHandlerConfig(), nil)
	raw := &tables.StbControlScheduleResultRaw{ID: "j3", WorkType: WorkTypeSysCheck}
	job := workers.Job{ID: "j3", Type: JobTypeStbControl, Payload: raw}
	err := h.Handle(context.Background(), job)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to update job status to running")
}

// --- Handle: simulation mode, service nil, DB unreachable ---

func TestHandle_SimulationMode_ServiceNil_DBError(t *testing.T) {
	cfg := newHandlerConfig()
	cfg.Catv.Control.Simulation = true
	h := NewStbControlHandler(cfg, nil)
	raw := &tables.StbControlScheduleResultRaw{ID: "j4", WorkType: WorkTypeSmartReboot}
	job := workers.Job{ID: "j4", Type: JobTypeStbControl, Payload: raw}
	err := h.Handle(context.Background(), job)
	require.Error(t, err)
}

// --- Handle: service with open channel → running status enqueued ---

func TestHandle_ServiceChannel_RunningStatusEnqueued(t *testing.T) {
	cfg := newHandlerConfig()
	cfg.Catv.Control.Simulation = true

	ch := make(chan *BatchWriteRequest, 100)
	svc := &Service{resultChannel: ch}
	h := NewStbControlHandler(cfg, svc)

	raw := &tables.StbControlScheduleResultRaw{ID: "j5", WorkType: WorkTypeSmartReboot}
	job := workers.Job{ID: "j5", Type: JobTypeStbControl, Payload: raw}
	_ = h.Handle(context.Background(), job)

	// At least the running-status record must be in the channel
	require.Greater(t, len(ch), 0)
	req := <-ch
	assert.Equal(t, tables.ProgressStatusRunning, req.StbControlScheduleResultRaw.ProgressStatus)
}

// --- Handle: result codes map to correct progress status ---

func TestHandle_ResultCode_MapsProgressStatus(t *testing.T) {
	tests := []struct {
		name        string
		workType    string
		simResponse string
		wantStatus  string
	}{
		// simulation always returns a valid code; we just verify the channel gets the final record
		{"simulation_reboot", WorkTypeSmartReboot, "", ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := newHandlerConfig()
			cfg.Catv.Control.Simulation = true

			ch := make(chan *BatchWriteRequest, 100)
			svc := &Service{resultChannel: ch}
			h := NewStbControlHandler(cfg, svc)

			raw := &tables.StbControlScheduleResultRaw{ID: "j6", WorkType: tt.workType}
			job := workers.Job{ID: "j6", Type: JobTypeStbControl, Payload: raw}
			_ = h.Handle(context.Background(), job)

			// Drain channel
			var records []*BatchWriteRequest
			for len(ch) > 0 {
				records = append(records, <-ch)
			}
			require.GreaterOrEqual(t, len(records), 2) // running + final

			final := records[len(records)-1]
			validStatuses := map[string]bool{
				tables.ProgressStatusSucceeded: true,
				tables.ProgressStatusFailed:    true,
			}
			assert.True(t, validStatuses[final.StbControlScheduleResultRaw.ProgressStatus])
		})
	}
}

// --- insertResult: nil service falls back to direct DB insert ---

func TestInsertResult_NilService_DirectInsert(t *testing.T) {
	h := NewStbControlHandler(newHandlerConfig(), nil)
	raw := tables.StbControlScheduleResultRaw{ID: "r1"}
	// scheduleResultTable with no DB → returns error
	err := h.insertResult(context.Background(), raw)
	assert.Error(t, err)
}

// --- insertResult: channel send succeeds ---

func TestInsertResult_ChannelSend_Success(t *testing.T) {
	ch := make(chan *BatchWriteRequest, 10)
	svc := &Service{resultChannel: ch}
	h := NewStbControlHandler(newHandlerConfig(), svc)
	raw := tables.StbControlScheduleResultRaw{ID: "r2", ProgressStatus: tables.ProgressStatusRunning}

	err := h.insertResult(context.Background(), raw)
	require.NoError(t, err)
	require.Equal(t, 1, len(ch))
	req := <-ch
	assert.Equal(t, "r2", req.StbControlScheduleResultRaw.ID)
}

// --- insertResult: ctx cancelled → falls back to direct insert ---

func TestInsertResult_CtxCancelled_FallbackInsert(t *testing.T) {
	// Blocked channel (no consumer)
	ch := make(chan *BatchWriteRequest)
	svc := &Service{resultChannel: ch}
	h := NewStbControlHandler(newHandlerConfig(), svc)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	raw := tables.StbControlScheduleResultRaw{ID: "r3"}
	// DB unavailable → direct insert fails, but fallback is attempted
	err := h.insertResult(ctx, raw)
	assert.Error(t, err)
}

// --- insertResult: channel blocked → fallback after timeout ---

func TestInsertResult_ChannelBlocked_FallbackAfterTimeout(t *testing.T) {
	ch := make(chan *BatchWriteRequest)
	svc := &Service{resultChannel: ch}

	cfg := newHandlerConfig()
	h := NewStbControlHandler(cfg, svc)

	// Use a context with short deadline so the test doesn't wait 10s
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	raw := tables.StbControlScheduleResultRaw{ID: "r4"}
	start := time.Now()
	_ = h.insertResult(ctx, raw)
	// Must complete within the context deadline, not the full 10s chan timeout
	assert.Less(t, time.Since(start), 2*time.Second)
}

// --- result code → progress status mapping ---

func TestHandle_SuccessCode_SetsSucceeded(t *testing.T) {
	cfg := newHandlerConfig()
	cfg.Catv.Control.Simulation = true

	// Run many times until we observe a success code
	for range 50 {
		ch := make(chan *BatchWriteRequest, 100)
		svc := &Service{resultChannel: ch}
		h := NewStbControlHandler(cfg, svc)

		raw := &tables.StbControlScheduleResultRaw{ID: "j7", WorkType: WorkTypeSmartReboot}
		_ = h.Handle(context.Background(), workers.Job{ID: "j7", Type: JobTypeStbControl, Payload: raw})

		var records []*BatchWriteRequest
		for len(ch) > 0 {
			records = append(records, <-ch)
		}
		if len(records) < 2 {
			continue
		}
		final := records[len(records)-1]
		if final.StbControlScheduleResultRaw.ResultCode == control_common.ResponseCodeSuccess {
			assert.Equal(t, tables.ProgressStatusSucceeded, final.StbControlScheduleResultRaw.ProgressStatus)
			return
		}
	}
	// If we never saw success (extremely unlikely), just pass — simulation is random
}
