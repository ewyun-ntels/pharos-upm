package schedules

import (
	"fmt"
	"slices"
	"strings"

	control_common "ntels.com/pharos/extensions/catv/business/pkg/control/common"
	"ntels.com/pharos/extensions/catv/business/pkg/control/common/tables"
)

type SchedulesResultsQueryFilter struct {
	ID             string   `form:"-"`               // Result ID from path parameter (required)
	K8sJobName     string   `form:"k8s_job_name"`    // Filter by specific task ID
	CmMacAddr      string   `form:"cm_mac_addr"`     // Cable Modem MAC address (exact or prefix with *)
	StbMacAddr     string   `form:"stb_mac_addr"`    // Set-Top Box MAC address (exact or prefix with *)
	StbMdlNm       string   `form:"stb_mdl_nm"`      // STB model name (exact match)
	ProgressStatus []string `form:"progress_status"` // Filter by task status (pending/running/succeeded/failed)
	ResultCode     string   `form:"result_code"`     // Filter by result code (1=success, 0=device error, -1=system error)
	Search         string   `form:"search"`          // Global search: k8s_job_name, cm_mac_addr, stb_mac_addr, stb_mdl_nm, result_message (case-insensitive LIKE)
	Limit          int      `form:"limit"`           // Maximum results to return
	Offset         uint64   `form:"offset"`          // Number of results to skip
}

type SchedulesResultsListQueryFilter struct {
	ID           string `form:"id"`            // Filter by specific job ID
	WorkType     string `form:"work_type"`     // Filter by work type (e.g., stb_request_info)
	ScheduleID   string `form:"schedule_id"`   // Filter by schedule ID
	ScheduleName string `form:"schedule_name"` // Filter by schedule name
	K8sJobName   string `form:"k8s_job_name"`  // Filter by specific task ID
	From         string `form:"from"`          // Start datetime (ISO 8601, e.g., 2025-01-01T00:00:00Z)
	To           string `form:"to"`            // End datetime (ISO 8601, e.g., 2025-12-31T23:59:59Z)
	Limit        int    `form:"limit"`         // Maximum results to return
	Offset       uint64 `form:"offset"`        // Number of results to skip
}

func (f *SchedulesResultsQueryFilter) Validate() error {
	for _, status := range f.ProgressStatus {
		if !isValidStatus(status) {
			return fmt.Errorf("invalid status value: %s (must be pending/running/succeeded/failed)", status)
		}
	}

	if f.ResultCode != "" && !isValidResultCode(f.ResultCode) {
		return fmt.Errorf("invalid result code: %s (must be 1/0/-1)", f.ResultCode)
	}

	if !isValidMacFormat(f.CmMacAddr) {
		return fmt.Errorf("invalid CM MAC address format: %s (wildcard * must be at the end only)", f.CmMacAddr)
	}

	if !isValidMacFormat(f.StbMacAddr) {
		return fmt.Errorf("invalid STB MAC address format: %s (wildcard * must be at the end only)", f.StbMacAddr)
	}

	return nil
}

func (f *SchedulesResultsQueryFilter) hasFilters() bool {
	return f.K8sJobName != "" || f.CmMacAddr != "" || f.StbMacAddr != "" || f.StbMdlNm != "" ||
		len(f.ProgressStatus) > 0 || f.ResultCode != "" || f.Search != ""
}

func isValidStatus(status string) bool {
	validStatuses := []string{
		tables.ProgressStatusPending,
		tables.ProgressStatusRunning,
		tables.ProgressStatusSucceeded,
		tables.ProgressStatusFailed,
	}
	return slices.Contains(validStatuses, status)
}

func isValidResultCode(code string) bool {
	return code == control_common.ResponseCodeSuccess ||
		code == control_common.ResponseCodeFailure ||
		code == control_common.ResponseCodeError
}

func isValidMacFormat(mac string) bool {
	if mac == "" {
		return true
	}

	asteriskCount := strings.Count(mac, "*")
	if asteriskCount > 1 {
		return false
	}

	if asteriskCount == 1 {
		if !strings.HasSuffix(mac, "*") {
			return false
		}
		mac = strings.TrimSuffix(mac, "*")
		if mac == "" {
			return false
		}
	}

	for _, char := range mac {
		if !strings.ContainsRune("0123456789ABCDEFabcdef:-", char) {
			return false
		}
	}

	return true
}
