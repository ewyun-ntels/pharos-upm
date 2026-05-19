package tables

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
)

func TestNewStbControlScheduleResultTable_NotNil(t *testing.T) {
	table := NewStbControlScheduleResultTable(common.Config{})
	require.NotNil(t, table)
}

func TestStbControlScheduleResultTable_HasConfig(t *testing.T) {
	config := common.Config{}
	config.Catv.Database.Driver = "clickhouse"

	table := NewStbControlScheduleResultTable(config)

	assert.Equal(t, config, table.config)
}

func TestStbControlScheduleResultTable_GetDistributedTableName(t *testing.T) {
	table := NewStbControlScheduleResultTable(common.Config{})
	assert.Equal(t, "dist_stb_control_schedule_result", table.GetDistributedTableName())
}

// --- DB 없을 때 에러 반환 확인 ---

func TestStbControlScheduleResultTable_SelectCountByStatusByID_FailsWithoutDB(t *testing.T) {
	table := NewStbControlScheduleResultTable(common.Config{})

	total, items, err := table.SelectCountByStatusByID(ControlResultListFilter{Limit: 10})

	assert.Error(t, err)
	assert.Zero(t, total)
	assert.Nil(t, items)
}

func TestStbControlScheduleResultTable_SelectWithFilter_FailsWithoutDB(t *testing.T) {
	table := NewStbControlScheduleResultTable(common.Config{})

	total, items, err := table.SelectWithFilter("test-id", ControlResultFilter{Limit: 10})

	assert.Error(t, err)
	assert.Zero(t, total)
	assert.Nil(t, items)
}

func TestStbControlScheduleResultTable_SelectWhereK8sJobName_FailsWithoutDB(t *testing.T) {
	table := NewStbControlScheduleResultTable(common.Config{})

	results, err := table.SelectWhereK8sJobName("catv-control-job-001")

	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestStbControlScheduleResultTable_Inserts_FailsWithoutDB(t *testing.T) {
	config := common.Config{}
	config.Catv.Control.Batch.MaxCount = 1 // 0이면 guard에서 에러

	table := NewStbControlScheduleResultTable(config)

	err := table.Inserts([]StbControlScheduleResultRaw{{ID: "r-001"}})

	assert.Error(t, err)
}

// --- ControlResultListFilter 구조체 ---

func TestControlResultListFilter_ZeroValue(t *testing.T) {
	f := ControlResultListFilter{}
	assert.Empty(t, f.ID)
	assert.Empty(t, f.WorkType)
	assert.Zero(t, f.Limit)
	assert.Zero(t, f.Offset)
}

// --- ControlScheduleResultCountByStatusByID 구조체 ---

func TestControlScheduleResultCountByStatusByID_Fields(t *testing.T) {
	r := ControlScheduleResultCountByStatusByID{
		ID:          "result-001",
		RequestTime: "2025-01-01T00:00:00Z",
		ScheduleID:  "sched-001",
		WorkType:    "stb_request_info",
		CountByStatus: map[string]uint64{
			ProgressStatusPending:   1,
			ProgressStatusRunning:   2,
			ProgressStatusSucceeded: 10,
			ProgressStatusFailed:    3,
		},
	}

	assert.Equal(t, "result-001", r.ID)
	assert.Equal(t, "2025-01-01T00:00:00Z", r.RequestTime)
	assert.Equal(t, "sched-001", r.ScheduleID)
	assert.Equal(t, "stb_request_info", r.WorkType)
	assert.Equal(t, uint64(10), r.CountByStatus[ProgressStatusSucceeded])
	assert.Equal(t, uint64(3), r.CountByStatus[ProgressStatusFailed])
}
