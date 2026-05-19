package tables

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/external/orm"
)

// --- Validate ---

func TestValidate_Immediately_OK(t *testing.T) {
	r := StbControlScheduleRaw{
		ScheduleType: ScheduleTypeImmediately,
		AreaType:     AreaTypeAll,
		WorkType:     "stb_request_info",
	}
	assert.NoError(t, r.Validate())
}

func TestValidate_Once_OK(t *testing.T) {
	future := orm.Datetime{Time: time.Now().UTC().Add(time.Hour)}
	r := StbControlScheduleRaw{
		ScheduleType:     ScheduleTypeOnce,
		ScheduleSpecOnce: &future,
		AreaType:         AreaTypeAll,
		WorkType:         "stb_request_info",
	}
	assert.NoError(t, r.Validate())
}

func TestValidate_Once_MissingSpec(t *testing.T) {
	r := StbControlScheduleRaw{
		ScheduleType: ScheduleTypeOnce,
		AreaType:     AreaTypeAll,
		WorkType:     "stb_request_info",
	}
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "once_spec is required")
}

func TestValidate_Once_PastTime(t *testing.T) {
	past := orm.Datetime{Time: time.Now().UTC().Add(-time.Hour)}
	r := StbControlScheduleRaw{
		ScheduleType:     ScheduleTypeOnce,
		ScheduleSpecOnce: &past,
		AreaType:         AreaTypeAll,
		WorkType:         "stb_request_info",
	}
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "future time")
}

func TestValidate_Repeat_OK(t *testing.T) {
	spec := "0 * * * * *"
	r := StbControlScheduleRaw{
		ScheduleType:       ScheduleTypeRepeat,
		ScheduleSpecRepeat: &spec,
		AreaType:           AreaTypeAll,
		WorkType:           "stb_request_info",
	}
	assert.NoError(t, r.Validate())
}

func TestValidate_Repeat_MissingSpec(t *testing.T) {
	r := StbControlScheduleRaw{
		ScheduleType: ScheduleTypeRepeat,
		AreaType:     AreaTypeAll,
		WorkType:     "stb_request_info",
	}
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "repeat_spec is required")
}

func TestValidate_InvalidScheduleType(t *testing.T) {
	r := StbControlScheduleRaw{
		ScheduleType: "unknown",
		AreaType:     AreaTypeAll,
		WorkType:     "stb_request_info",
	}
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid schedule_type")
}

func TestValidate_NonAllAreaType_MissingAreaIDs(t *testing.T) {
	r := StbControlScheduleRaw{
		ScheduleType: ScheduleTypeImmediately,
		AreaType:     AreaTypeSO,
		AreaIDs:      []string{},
		WorkType:     "stb_request_info",
	}
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "area_ids is required")
}

func TestValidate_NonAllAreaType_WithAreaIDs_OK(t *testing.T) {
	r := StbControlScheduleRaw{
		ScheduleType: ScheduleTypeImmediately,
		AreaType:     AreaTypeSO,
		AreaIDs:      []string{"SO-01"},
		WorkType:     "stb_request_info",
	}
	assert.NoError(t, r.Validate())
}

func TestValidate_MissingWorkType(t *testing.T) {
	r := StbControlScheduleRaw{
		ScheduleType: ScheduleTypeImmediately,
		AreaType:     AreaTypeAll,
		WorkType:     "",
	}
	err := r.Validate()
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "work_type is required")
}

// AreaType이 대소문자 무관하게 "all"로 처리되는지 확인
func TestValidate_AreaType_CaseInsensitive(t *testing.T) {
	r := StbControlScheduleRaw{
		ScheduleType: ScheduleTypeImmediately,
		AreaType:     "ALL",
		WorkType:     "stb_request_info",
	}
	assert.NoError(t, r.Validate())
}

// --- 상수 값 검증 ---

func TestScheduleTypeConstants(t *testing.T) {
	assert.Equal(t, "immediately", ScheduleTypeImmediately)
	assert.Equal(t, "once", ScheduleTypeOnce)
	assert.Equal(t, "repeat", ScheduleTypeRepeat)
}
