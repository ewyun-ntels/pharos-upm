package tables

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
)

func TestNewStbControlScheduleTable_NotNil(t *testing.T) {
	config := common.Config{}

	repo := NewStbControlScheduleTable(config)

	require.NotNil(t, repo)
}

func TestStbControlScheduleRepository_HasConfig(t *testing.T) {
	config := common.Config{}
	config.Catv.Database.Driver = "sqlite"

	repo := NewStbControlScheduleTable(config)

	assert.Equal(t, config, repo.config)
}

func TestDistStbControlScheduleTableName(t *testing.T) {
	assert.Equal(t, "dist_stb_control_schedule", distStbControlScheduleTableName)
}

func TestStbControlScheduleRaw_Fields(t *testing.T) {
	raw := StbControlScheduleRaw{
		ID:           "id-001",
		Name:         "테스트 스케줄",
		ScheduleType: ScheduleTypeImmediately,
		AreaType:     AreaTypeAll,
		WorkType:     "stb_request_info",
		IsDeleted:    0,
	}

	assert.Equal(t, "id-001", raw.ID)
	assert.Equal(t, "테스트 스케줄", raw.Name)
	assert.Equal(t, ScheduleTypeImmediately, raw.ScheduleType)
	assert.Equal(t, AreaTypeAll, raw.AreaType)
	assert.Equal(t, "stb_request_info", raw.WorkType)
	assert.Equal(t, uint8(0), raw.IsDeleted)
}

func TestStbControlScheduleRepository_Create_FailsWithoutDB(t *testing.T) {
	config := common.Config{}

	repo := NewStbControlScheduleTable(config)
	raw := StbControlScheduleRaw{
		Name:         "test",
		ScheduleType: ScheduleTypeImmediately,
		AreaType:     AreaTypeAll,
		WorkType:     "stb_request_info",
	}

	// DB 없이 Create 호출하면 에러 반환
	err := repo.Create(raw)
	assert.Error(t, err)
}

func TestStbControlScheduleRepository_GetByName_FailsWithoutDB(t *testing.T) {
	config := common.Config{}
	repo := NewStbControlScheduleTable(config)

	_, err := repo.GetByName("test")
	assert.Error(t, err)
}

func TestStbControlScheduleRepository_GetAll_FailsWithoutDB(t *testing.T) {
	config := common.Config{}
	repo := NewStbControlScheduleTable(config)

	_, err := repo.GetAll("")
	assert.Error(t, err)
}

func TestStbControlScheduleRepository_Delete_FailsWithoutDB(t *testing.T) {
	config := common.Config{}
	repo := NewStbControlScheduleTable(config)

	err := repo.Delete("test")
	assert.Error(t, err)
}
