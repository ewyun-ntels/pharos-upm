package tables

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
)

// --- 생성 및 설정 ---

func TestNewStbInformationTable_NotNil(t *testing.T) {
	table := NewStbInformationTable(common.Config{})
	require.NotNil(t, table)
}

func TestStbInformationTable_HasConfig(t *testing.T) {
	config := common.Config{}
	config.Catv.Database.Driver = "clickhouse"

	table := NewStbInformationTable(config)

	assert.Equal(t, config, table.config)
}

// --- AreaType 상수 ---

func TestAreaTypeConstants(t *testing.T) {
	assert.Equal(t, "all", AreaTypeAll)
	assert.Equal(t, "so", AreaTypeSO)
	assert.Equal(t, "l3", AreaTypeL3)
	assert.Equal(t, "cell", AreaTypeCell)
	assert.Equal(t, "settopbox", AreaTypeSettopbox)
}

// --- DB 없을 때 에러 반환 확인 ---

func TestStbInformationTable_GetSos_FailsWithoutDB(t *testing.T) {
	table := NewStbInformationTable(common.Config{})

	results, err := table.GetSos()

	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestStbInformationTable_GetL3s_FailsWithoutDB(t *testing.T) {
	table := NewStbInformationTable(common.Config{})

	results, err := table.GetL3s("SO-01")

	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestStbInformationTable_GetCells_FailsWithoutDB(t *testing.T) {
	table := NewStbInformationTable(common.Config{})

	results, err := table.GetCells("SO-01", "L3-01")

	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestStbInformationTable_GetSettopboxes_FailsWithoutDB(t *testing.T) {
	table := NewStbInformationTable(common.Config{})

	results, err := table.GetSettopboxes("SO-01", "L3-01", "CELL-01")

	assert.Error(t, err)
	assert.Nil(t, results)
}

func TestStbInformationTable_GetStbInformationRaws_FailsWithoutDB(t *testing.T) {
	table := NewStbInformationTable(common.Config{})

	results, err := table.GetStbInformationRaws(AreaTypeAll, nil)

	assert.Error(t, err)
	assert.Nil(t, results)
}

// --- GetStbInformationRaws: 유효하지 않은 area type ---

func TestGetStbInformationRaws_InvalidAreaType(t *testing.T) {
	table := NewStbInformationTable(common.Config{})

	results, err := table.GetStbInformationRaws("invalid_area", []string{"id-01"})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid area type")
	assert.Nil(t, results)
}

// --- GetStbInformationRaws: 빈 IDs는 빈 슬라이스 반환 (DB 호출 없음) ---

func TestGetStbInformationRaws_EmptyIDs_SO(t *testing.T) {
	table := NewStbInformationTable(common.Config{})

	results, err := table.GetStbInformationRaws(AreaTypeSO, []string{})

	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestGetStbInformationRaws_EmptyIDs_L3(t *testing.T) {
	table := NewStbInformationTable(common.Config{})

	results, err := table.GetStbInformationRaws(AreaTypeL3, []string{})

	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestGetStbInformationRaws_EmptyIDs_Cell(t *testing.T) {
	table := NewStbInformationTable(common.Config{})

	results, err := table.GetStbInformationRaws(AreaTypeCell, []string{})

	assert.NoError(t, err)
	assert.Empty(t, results)
}

func TestGetStbInformationRaws_EmptyIDs_Settopbox(t *testing.T) {
	table := NewStbInformationTable(common.Config{})

	results, err := table.GetStbInformationRaws(AreaTypeSettopbox, []string{})

	assert.NoError(t, err)
	assert.Empty(t, results)
}
