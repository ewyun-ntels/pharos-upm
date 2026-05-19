package tables

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/external/orm"
)

// --- 상수 값 검증 ---

func TestProgressStatusConstants(t *testing.T) {
	assert.Equal(t, "pending", ProgressStatusPending)
	assert.Equal(t, "running", ProgressStatusRunning)
	assert.Equal(t, "succeeded", ProgressStatusSucceeded)
	assert.Equal(t, "failed", ProgressStatusFailed)
}

// --- ConvertBatchFormat ---

func TestConvertBatchFormat_AllFields(t *testing.T) {
	workValue := "testValue"
	now := time.Now().UTC().Truncate(time.Second)
	controlStartTime := orm.Datetime{Time: now}
	controlEndTime := orm.Datetime{Time: now}

	r := StbControlScheduleResultRaw{
		ID:               "result-001",
		RequestTime:      orm.Datetime{Time: now},
		ControlStartTime: &controlStartTime,
		ControlEndTime:   &controlEndTime,
		ScheduleID:       "sched-001",
		ScheduleName:     "테스트 스케줄",
		ScheduleType:     ScheduleTypeImmediately,
		K8sJobName:       "catv-control-job-001",
		CmMacAddr:        "AA:BB:CC:DD:EE:FF",
		CmIpAddr:         "192.168.1.1",
		StbMacAddr:       "11:22:33:44:55:66",
		SrcIpAddr:        "10.0.0.1",
		WorkType:         "stb_request_info",
		WorkValue:        &workValue,
		ProgressStatus:   ProgressStatusSucceeded,
		ResultCode:       "1",
		ResultMessage:    "success",
	}

	m, err := r.ConvertBatchFormat()
	require.NoError(t, err)

	assert.Equal(t, "result-001", m["id"])
	assert.Equal(t, "sched-001", m["schedule_id"])
	assert.Equal(t, "테스트 스케줄", m["schedule_name"])
	assert.Equal(t, ScheduleTypeImmediately, m["schedule_type"])
	assert.Equal(t, "catv-control-job-001", m["k8s_job_name"])
	assert.Equal(t, "AA:BB:CC:DD:EE:FF", m["cm_mac_addr"])
	assert.Equal(t, "192.168.1.1", m["cm_ip_addr"])
	assert.Equal(t, "11:22:33:44:55:66", m["stb_mac_addr"])
	assert.Equal(t, "10.0.0.1", m["src_ip_addr"])
	assert.Equal(t, "stb_request_info", m["work_type"])
	assert.Equal(t, "testValue", m["work_value"])
	assert.Equal(t, ProgressStatusSucceeded, m["progress_status"])
	assert.Equal(t, "1", m["result_code"])
	assert.Equal(t, "success", m["result_message"])
	assert.NotNil(t, m["request_time"])
	assert.NotNil(t, m["control_start_time"])
	assert.NotNil(t, m["control_end_time"])
}

func TestConvertBatchFormat_NilControlTime(t *testing.T) {
	r := StbControlScheduleResultRaw{
		ID:               "result-002",
		RequestTime:      orm.Datetime{Time: time.Now().UTC()},
		ControlStartTime: nil,
		ControlEndTime:   nil,
		ScheduleID:       "sched-001",
		K8sJobName:       "catv-control-job-001",
		WorkType:         "stb_request_info",
	}

	m, err := r.ConvertBatchFormat()
	require.NoError(t, err)
	assert.Nil(t, m["control_start_time"])
	assert.Nil(t, m["control_end_time"])
}

func TestConvertBatchFormat_NilWorkValue(t *testing.T) {
	r := StbControlScheduleResultRaw{
		ID:          "result-003",
		RequestTime: orm.Datetime{Time: time.Now().UTC()},
		WorkValue:   nil,
		WorkType:    "stb_request_info",
	}

	m, err := r.ConvertBatchFormat()
	require.NoError(t, err)
	assert.Nil(t, m["work_value"])
}

func TestConvertBatchFormat_HasExpectedKeys(t *testing.T) {
	r := StbControlScheduleResultRaw{
		RequestTime: orm.Datetime{Time: time.Now().UTC()},
		WorkType:    "stb_request_info",
	}

	m, err := r.ConvertBatchFormat()
	require.NoError(t, err)

	expectedKeys := []string{
		"id", "request_time", "control_start_time", "control_end_time",
		"schedule_id", "schedule_name", "schedule_type",
		"k8s_job_name",
		"cm_mac_addr", "cm_ip_addr", "stb_mac_addr", "src_ip_addr",
		"work_type", "work_value",
		"progress_status", "result_code", "result_message",
	}
	for _, key := range expectedKeys {
		assert.Contains(t, m, key, "key %q should be present", key)
	}
}
