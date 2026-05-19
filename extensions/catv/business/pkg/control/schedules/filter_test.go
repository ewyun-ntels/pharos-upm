package schedules

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// SchedulesResultsQueryFilter.Validate

func TestSchedulesResultsQueryFilter_Validate_OK(t *testing.T) {
	f := SchedulesResultsQueryFilter{
		ProgressStatus: []string{"pending", "running", "succeeded", "failed"},
		ResultCode:     "1",
		CmMacAddr:      "AA:BB:CC:DD:EE:FF",
		StbMacAddr:     "11:22:33:44:55*",
	}
	assert.NoError(t, f.Validate())
}

func TestSchedulesResultsQueryFilter_Validate_InvalidStatus(t *testing.T) {
	f := SchedulesResultsQueryFilter{ProgressStatus: []string{"unknown"}}
	assert.Error(t, f.Validate())
}

func TestSchedulesResultsQueryFilter_Validate_InvalidResultCode(t *testing.T) {
	f := SchedulesResultsQueryFilter{ResultCode: "2"}
	assert.Error(t, f.Validate())
}

func TestSchedulesResultsQueryFilter_Validate_ValidResultCodes(t *testing.T) {
	for _, code := range []string{"1", "0", "-1"} {
		f := SchedulesResultsQueryFilter{ResultCode: code}
		assert.NoError(t, f.Validate(), "code=%s", code)
	}
}

func TestSchedulesResultsQueryFilter_Validate_InvalidCmMacWildcard(t *testing.T) {
	f := SchedulesResultsQueryFilter{CmMacAddr: "*AA:BB"}
	assert.Error(t, f.Validate())
}

func TestSchedulesResultsQueryFilter_Validate_MultipleWildcard(t *testing.T) {
	f := SchedulesResultsQueryFilter{StbMacAddr: "AA:BB:*:*"}
	assert.Error(t, f.Validate())
}

func TestSchedulesResultsQueryFilter_Validate_InvalidChar(t *testing.T) {
	f := SchedulesResultsQueryFilter{CmMacAddr: "ZZ:ZZ:ZZ"}
	assert.Error(t, f.Validate())
}

func TestSchedulesResultsQueryFilter_Validate_TrailingWildcardOK(t *testing.T) {
	f := SchedulesResultsQueryFilter{CmMacAddr: "AA:BB:CC*"}
	assert.NoError(t, f.Validate())
}

func TestSchedulesResultsQueryFilter_Validate_EmptyMacOK(t *testing.T) {
	f := SchedulesResultsQueryFilter{}
	assert.NoError(t, f.Validate())
}

// hasFilters

func TestSchedulesResultsQueryFilter_HasFilters_False(t *testing.T) {
	f := SchedulesResultsQueryFilter{}
	assert.False(t, f.hasFilters())
}

func TestSchedulesResultsQueryFilter_HasFilters_K8sJobName(t *testing.T) {
	f := SchedulesResultsQueryFilter{K8sJobName: "job-1"}
	assert.True(t, f.hasFilters())
}

func TestSchedulesResultsQueryFilter_HasFilters_ProgressStatus(t *testing.T) {
	f := SchedulesResultsQueryFilter{ProgressStatus: []string{"running"}}
	assert.True(t, f.hasFilters())
}

func TestSchedulesResultsQueryFilter_HasFilters_Search(t *testing.T) {
	f := SchedulesResultsQueryFilter{Search: "keyword"}
	assert.True(t, f.hasFilters())
}

// isValidMacFormat

func TestIsValidMacFormat_Empty(t *testing.T) {
	assert.True(t, isValidMacFormat(""))
}

func TestIsValidMacFormat_Valid(t *testing.T) {
	assert.True(t, isValidMacFormat("AA:BB:CC:DD:EE:FF"))
	assert.True(t, isValidMacFormat("aa-bb-cc-dd-ee-ff"))
}

func TestIsValidMacFormat_WildcardOnlyInvalid(t *testing.T) {
	assert.False(t, isValidMacFormat("*"))
}

func TestIsValidMacFormat_PrefixWildcardInvalid(t *testing.T) {
	assert.False(t, isValidMacFormat("*AA:BB"))
}
