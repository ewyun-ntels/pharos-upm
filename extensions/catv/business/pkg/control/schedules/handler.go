package schedules

import (
	"log/slog"
	net_http "net/http"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/extensions/catv/business/pkg/control/common/tables"
)

const (
	MaxRequestBodySize  = 10 * 1024 * 1024 // 10MB maximum request body size
	MaxJobListLimit     = 1000             // Maximum number of jobs to return in list
	MaxJobResultsLimit  = 10000            // Maximum number of results to return
	DefaultJobListLimit = 100              // Default number of jobs in list
	DefaultResultsLimit = 1000             // Default number of results
)

func getSchedulesResultsListHandler(config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var filter SchedulesResultsListQueryFilter

		if err := c.ShouldBindQuery(&filter); err != nil {
			slog.Error("Failed to bind query parameters", "error", err)
			c.JSON(net_http.StatusBadRequest, external.NewErrorResponse("invalid query parameters"))
			return
		}

		if filter.Limit == 0 {
			filter.Limit = DefaultJobListLimit
		}
		if filter.Limit < 1 {
			filter.Limit = 1
		}
		if filter.Limit > MaxJobListLimit {
			filter.Limit = MaxJobListLimit
		}

		var stbControlScheduleResultTable = tables.NewStbControlScheduleResultTable(config)

		total, items, err := stbControlScheduleResultTable.SelectCountByStatusByID(tables.ControlResultListFilter{
			ID:           filter.ID,
			WorkType:     filter.WorkType,
			ScheduleID:   filter.ScheduleID,
			ScheduleName: filter.ScheduleName,
			K8sJobName:   filter.K8sJobName,
			From:         filter.From,
			To:           filter.To,
			Limit:        filter.Limit,
			Offset:       filter.Offset,
		})
		if err != nil {
			slog.Error("Failed to query job results", "error", err)
			c.JSON(net_http.StatusInternalServerError, external.NewErrorResponse("failed to retrieve job results"))
			return
		}

		c.JSON(net_http.StatusOK, map[string]any{
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
			"filters": map[string]any{
				"id":            filter.ID,
				"work_type":     filter.WorkType,
				"k8s_job_name":  filter.K8sJobName,
				"schedule_id":   filter.ScheduleID,
				"schedule_name": filter.ScheduleName,
				"from":          filter.From,
				"to":            filter.To,
			},
			"items": items,
		})
	}
}

func getSchedulesResultsHandler(config common.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		var filter SchedulesResultsQueryFilter
		filter.ID = c.Param("id")

		if filter.ID == "" {
			c.JSON(net_http.StatusBadRequest, external.NewErrorResponse("id is required"))
			return
		}

		if err := c.ShouldBindQuery(&filter); err != nil {
			slog.Error("Failed to bind query parameters", "error", err)
			c.JSON(net_http.StatusBadRequest, external.NewErrorResponse("invalid query parameters"))
			return
		}

		if filter.Limit == 0 {
			filter.Limit = DefaultResultsLimit
		}

		if filter.Limit < 1 {
			filter.Limit = 1
		}
		if filter.Limit > MaxJobResultsLimit {
			filter.Limit = MaxJobResultsLimit
		}

		if err := filter.Validate(); err != nil {
			slog.Error("Invalid filter parameters", "error", err)
			c.JSON(net_http.StatusBadRequest, external.NewErrorResponse(err.Error()))
			return
		}

		var stbControlScheduleResultTable = tables.NewStbControlScheduleResultTable(config)

		total, items, err := stbControlScheduleResultTable.SelectWithFilter(filter.ID, tables.ControlResultFilter{
			K8sJobName:     filter.K8sJobName,
			CmMacAddr:      filter.CmMacAddr,
			StbMacAddr:     filter.StbMacAddr,
			StbMdlNm:       filter.StbMdlNm,
			ProgressStatus: filter.ProgressStatus,
			ResultCode:     filter.ResultCode,
			Search:         filter.Search,
			Limit:          filter.Limit,
			Offset:         filter.Offset,
		})
		if err != nil {
			slog.Error("Failed to query job results", "id", filter.ID, "error", err)
			c.JSON(net_http.StatusInternalServerError, external.NewErrorResponse("failed to retrieve job results"))
			return
		}
		if total == 0 && filter.Offset == 0 && !filter.hasFilters() {
			c.JSON(net_http.StatusNotFound, external.NewErrorResponse("id not found"))
			return
		}

		c.JSON(net_http.StatusOK, map[string]any{
			"id":     filter.ID,
			"total":  total,
			"limit":  filter.Limit,
			"offset": filter.Offset,
			"filters": map[string]any{
				"k8s_job_name":    filter.K8sJobName,
				"cm_mac_addr":     filter.CmMacAddr,
				"stb_mac_addr":    filter.StbMacAddr,
				"stb_mdl_nm":      filter.StbMdlNm,
				"progress_status": filter.ProgressStatus,
				"result_code":     filter.ResultCode,
				"search":          filter.Search,
			},
			"items": items,
		})
	}
}
