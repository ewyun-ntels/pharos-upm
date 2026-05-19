package command

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/extensions/catv/business/pkg/control/common/tables"
)

// --- NewService ---

func TestNewService_DBError(t *testing.T) {
	// Empty config → DB connection fails → SelectWhereK8sJobName returns error
	service, err := NewService(newTestConfig(), "test-job")
	require.Error(t, err)
	assert.Nil(t, service)
}

// --- Service.GetResultChannel ---

func TestService_GetResultChannel_NilWhenUnloaded(t *testing.T) {
	svc := &Service{}
	assert.Nil(t, svc.resultChannel)
}

// --- Service.Start: no jobs ---

func TestService_Start_NoJobs(t *testing.T) {
	cfg := newTestConfig()
	cfg.Catv.Control.MaxConnectionsPerSecond = 300
	cfg.Catv.Control.Batch.MaxCount = 10
	cfg.Catv.Control.Batch.FlushInterval = "100ms"
	cfg.Catv.Control.Timeout.Connection = "100ms"
	cfg.Catv.Control.Timeout.Send = "100ms"
	cfg.Catv.Control.Retry.PortExhaustion.Count = 1
	cfg.Catv.Control.Retry.PortExhaustion.Delay = "1ms"
	cfg.Catv.Control.Retry.Total.Count = 1
	cfg.Catv.Control.Retry.Total.Delay = "1ms"
	cfg.Pool.Workers.WorkerCount = 2
	cfg.Pool.Workers.JobQueueSize = 10
	cfg.Pool.Workers.BatchSize = 5

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	svc := &Service{
		config:                        cfg,
		k8sJobName:                    "empty-job",
		stbControlScheduleResultRaws:  []tables.StbControlScheduleResultRaw{},
		ctx:                           ctx,
		cancel:                        cancel,
		stbControlScheduleResultTable: tables.NewStbControlScheduleResultTable(cfg),
	}

	err := svc.Start()
	require.NoError(t, err)
	defer svc.Stop()
}
