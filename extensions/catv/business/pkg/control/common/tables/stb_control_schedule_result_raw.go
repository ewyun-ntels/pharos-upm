package tables

import (
	"fmt"
	"time"

	"ntels.com/pharos/core/external/orm"
)

const (
	ProgressStatusPending   = "pending"
	ProgressStatusRunning   = "running"
	ProgressStatusSucceeded = "succeeded"
	ProgressStatusFailed    = "failed"
)

type StbControlScheduleResultRaw struct {
	ID string `json:"id" db:"id"`

	RequestTime orm.Datetime `json:"request_time" db:"request_time"`

	ControlStartTime *orm.Datetime `json:"control_start_time,omitempty" db:"control_start_time"`
	ControlEndTime   *orm.Datetime `json:"control_end_time,omitempty" db:"control_end_time"`

	ScheduleID   string `json:"schedule_id" db:"schedule_id"`
	ScheduleName string `json:"schedule_name" db:"schedule_name"`
	ScheduleType string `json:"schedule_type" db:"schedule_type"`

	K8sJobName string `json:"k8s_job_name" db:"k8s_job_name"`

	StbMdlNm   string  `json:"stb_mdl_nm" db:"stb_mdl_nm"`
	CmMacAddr  string  `json:"cm_mac_addr" db:"cm_mac_addr"`
	CmIpAddr   string  `json:"cm_ip_addr" db:"cm_ip_addr"`
	StbMacAddr string  `json:"stb_mac_addr" db:"stb_mac_addr"`
	SrcIpAddr  string  `json:"src_ip_addr" db:"src_ip_addr"`
	WorkType   string  `json:"work_type" db:"work_type"`
	WorkValue  *string `json:"work_value,omitempty" db:"work_value"`

	ProgressStatus string `json:"progress_status" db:"progress_status"`
	ResultCode     string `json:"result_code" db:"result_code"`
	ResultMessage  string `json:"result_message" db:"result_message"`

	UpdateTime orm.Datetime `json:"update_time" db:"update_time"`
}

func (r StbControlScheduleResultRaw) ConvertBatchFormat() (map[string]any, error) {
	requestTime, err := r.RequestTime.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to convert request time: %w", err)
	}

	var controlStartTime any
	if r.ControlStartTime != nil {
		controlStartTime, err = r.ControlStartTime.Value()
		if err != nil {
			return nil, fmt.Errorf("failed to convert control start time: %w", err)
		}
	} else {
		controlStartTime = nil
	}

	var controlEndTime any
	if r.ControlEndTime != nil {
		controlEndTime, err = r.ControlEndTime.Value()
		if err != nil {
			return nil, fmt.Errorf("failed to convert control end time: %w", err)
		}
	} else {
		controlEndTime = nil
	}

	r.UpdateTime = orm.Datetime{Time: time.Now().UTC()}
	updateTime, err := r.UpdateTime.Value()
	if err != nil {
		return nil, fmt.Errorf("failed to convert update time: %w", err)
	}

	var workValue any
	if r.WorkValue != nil {
		workValue = *r.WorkValue
	} else {
		workValue = nil
	}

	return map[string]any{
		"id": r.ID,

		"request_time": requestTime,

		"control_start_time": controlStartTime,
		"control_end_time":   controlEndTime,

		"schedule_id":   r.ScheduleID,
		"schedule_name": r.ScheduleName,
		"schedule_type": r.ScheduleType,

		"k8s_job_name": r.K8sJobName,

		"stb_mdl_nm":   r.StbMdlNm,
		"cm_mac_addr":  r.CmMacAddr,
		"cm_ip_addr":   r.CmIpAddr,
		"stb_mac_addr": r.StbMacAddr,
		"src_ip_addr":  r.SrcIpAddr,
		"work_type":    r.WorkType,
		"work_value":   workValue,

		"progress_status": r.ProgressStatus,
		"result_code":     r.ResultCode,
		"result_message":  r.ResultMessage,

		"update_time": updateTime,
	}, nil
}
