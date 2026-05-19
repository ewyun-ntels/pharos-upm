package tables

import (
	"fmt"
	"log/slog"
	"strings"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/clickhouse_interface"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	control_common "ntels.com/pharos/extensions/catv/business/pkg/control/common"
)

type ControlResultListFilter struct {
	ID           string `json:"id" form:"id"`
	WorkType     string `json:"work_type" form:"work_type"`
	ScheduleID   string `json:"schedule_id" form:"schedule_id"`
	ScheduleName string `json:"schedule_name" form:"schedule_name"`
	K8sJobName   string `json:"k8s_job_name" form:"k8s_job_name"`
	From         string `json:"from" form:"from"`
	To           string `json:"to" form:"to"`
	Limit        int    `json:"limit" form:"limit"`
	Offset       uint64 `json:"offset" form:"offset"`
}

type ControlResultFilter struct {
	K8sJobName     string   `json:"k8s_job_name" form:"k8s_job_name"`
	CmMacAddr      string   `json:"cm_mac_addr" form:"cm_mac_addr"`
	StbMacAddr     string   `json:"stb_mac_addr" form:"stb_mac_addr"`
	StbMdlNm       string   `json:"stb_mdl_nm" form:"stb_mdl_nm"`
	ProgressStatus []string `json:"progress_status" form:"progress_status"`
	ResultCode     string   `json:"result_code" form:"result_code"`
	Search         string   `json:"search" form:"search"`
	Limit          int      `json:"limit" form:"limit"`
	Offset         uint64   `json:"offset" form:"offset"`
}

type ControlScheduleResultCountByStatusByID struct {
	ID             string            `json:"id" db:"id"`
	RequestTime    string            `json:"request_time" db:"request_time"`
	LastUpdateTime string            `json:"last_update_time" db:"last_update_time"`
	ScheduleID     string            `json:"schedule_id" db:"schedule_id"`
	ScheduleName   string            `json:"schedule_name" db:"schedule_name"`
	ScheduleType   string            `json:"schedule_type" db:"schedule_type"`
	K8sJobNames    []string          `json:"k8s_job_names" db:"k8s_job_names"`
	WorkType       string            `json:"work_type" db:"work_type"`
	CountByStatus  map[string]uint64 `json:"count_by_status" db:"count_by_status"`
}

func NewStbControlScheduleResultTable(config common.Config) StbControlScheduleResultTable {
	return StbControlScheduleResultTable{config: config}
}

type StbControlScheduleResultTable struct {
	config common.Config
}

func (t StbControlScheduleResultTable) GetDistributedTableName() string {
	return "dist_stb_control_schedule_result"
}

func (t StbControlScheduleResultTable) SelectWhereK8sJobName(k8sJobName string) ([]StbControlScheduleResultRaw, error) {
	var results []StbControlScheduleResultRaw

	handler := func(db *sqlx.DB) error {
		query := `SELECT * FROM ` + t.GetDistributedTableName() + ` FINAL WHERE k8s_job_name = $1;`
		return db.Select(&results, query, k8sJobName)
	}

	if err := orm.StatisticsHandler(t.config.Catv.Database.Driver, &t.config.Catv.Database, handler); err != nil {
		slog.Error("Failed to select control schedule results by k8s job name", "k8s_job_name", k8sJobName, "error", err)
		return nil, err
	}

	return results, nil
}

func (t StbControlScheduleResultTable) SelectCountByStatusByID(filter ControlResultListFilter) (int, []ControlScheduleResultCountByStatusByID, error) {
	var total int
	var results []ControlScheduleResultCountByStatusByID

	handler := func(db *sqlx.DB) error {
		var whereConditions []string
		var args []any

		if filter.ID != "" {
			whereConditions = append(whereConditions, "id = ?")
			args = append(args, filter.ID)
		}

		if filter.WorkType != "" {
			whereConditions = append(whereConditions, "work_type = ?")
			args = append(args, filter.WorkType)
		}

		if filter.ScheduleID != "" {
			whereConditions = append(whereConditions, "schedule_id = ?")
			args = append(args, filter.ScheduleID)
		}

		if filter.ScheduleName != "" {
			whereConditions = append(whereConditions, "schedule_name ilike ?")
			args = append(args, "%"+filter.ScheduleName+"%")
		}

		if filter.K8sJobName != "" {
			whereConditions = append(whereConditions, "k8s_job_name = ?")
			args = append(args, filter.K8sJobName)
		}

		if filter.From != "" {
			whereConditions = append(whereConditions, "request_time >= parseDateTimeBestEffort(?)")
			args = append(args, filter.From)
		}

		if filter.To != "" {
			whereConditions = append(whereConditions, "request_time <= parseDateTimeBestEffort(?)")
			args = append(args, filter.To)
		}

		whereClause := ""
		if len(whereConditions) > 0 {
			whereClause = "WHERE " + strings.Join(whereConditions, " AND ")
		}

		countQuery := `SELECT uniqExact(id, request_time, schedule_id, schedule_name, schedule_type, work_type) FROM ` + t.GetDistributedTableName() + ` FINAL ` + whereClause
		if err := db.Get(&total, countQuery, args...); err != nil {
			return err
		}

		query := `
SELECT
    id,
    request_time,
    max(update_time) AS last_update_time,
    schedule_id,
    schedule_name,
    schedule_type,
    groupArrayDistinct(k8s_job_name) AS k8s_job_names,
    work_type,
    sumMap(map(progress_status, 1)) AS count_by_status
FROM ` + t.GetDistributedTableName() + ` FINAL ` + whereClause + `
GROUP BY id, request_time, schedule_id, schedule_name, schedule_type, work_type
ORDER BY request_time desc, schedule_name desc
LIMIT ? OFFSET ?
`

		selectArgs := append(args, filter.Limit, filter.Offset)
		return db.Select(&results, query, selectArgs...)
	}

	if err := orm.StatisticsHandler(t.config.Catv.Database.Driver, &t.config.Catv.Database, handler); err != nil {
		return 0, nil, fmt.Errorf("failed to select count by status: %w", err)
	}

	return total, results, nil
}

func (t StbControlScheduleResultTable) SelectWithFilter(id string, filter ControlResultFilter) (int, []StbControlScheduleResultRaw, error) {
	var total int
	var results []StbControlScheduleResultRaw

	handler := func(db *sqlx.DB) error {
		var whereConditions []string
		var args []any

		whereConditions = append(whereConditions, "id = ?")
		args = append(args, id)

		if filter.K8sJobName != "" {
			whereConditions = append(whereConditions, "k8s_job_name = ?")
			args = append(args, filter.K8sJobName)
		}

		if filter.CmMacAddr != "" {
			if strings.HasSuffix(filter.CmMacAddr, "*") {
				whereConditions = append(whereConditions, "startsWith(cm_mac_addr, ?)")
				prefix := strings.TrimSuffix(filter.CmMacAddr, "*")
				args = append(args, prefix)
			} else {
				whereConditions = append(whereConditions, "cm_mac_addr = ?")
				args = append(args, filter.CmMacAddr)
			}
		}

		if filter.StbMacAddr != "" {
			if strings.HasSuffix(filter.StbMacAddr, "*") {
				whereConditions = append(whereConditions, "startsWith(stb_mac_addr, ?)")
				prefix := strings.TrimSuffix(filter.StbMacAddr, "*")
				args = append(args, prefix)
			} else {
				whereConditions = append(whereConditions, "stb_mac_addr = ?")
				args = append(args, filter.StbMacAddr)
			}
		}

		if filter.StbMdlNm != "" {
			whereConditions = append(whereConditions, "stb_mdl_nm = ?")
			args = append(args, filter.StbMdlNm)
		}

		if len(filter.ProgressStatus) > 0 {
			whereConditions = append(whereConditions, "progress_status IN (?)")
			args = append(args, filter.ProgressStatus)
		}

		if filter.ResultCode != "" {
			whereConditions = append(whereConditions, "result_code = ?")
			args = append(args, filter.ResultCode)
		}

		if filter.Search != "" {
			whereConditions = append(whereConditions, "(k8s_job_name ilike ? OR cm_mac_addr ilike ? OR stb_mac_addr ilike ? OR stb_mdl_nm ilike ? OR result_message ilike ?)")
			pattern := "%" + filter.Search + "%"
			args = append(args, pattern, pattern, pattern, pattern, pattern)
		}

		whereClause := strings.Join(whereConditions, " AND ")

		countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s FINAL WHERE %s", t.GetDistributedTableName(), whereClause)
		if len(filter.ProgressStatus) > 0 {
			q, countArgs, err := sqlx.In(countQuery, args...)
			if err != nil {
				return fmt.Errorf("failed to build count IN query: %w", err)
			}
			if err := db.Get(&total, db.Rebind(q), countArgs...); err != nil {
				return fmt.Errorf("failed to get total count: %w", err)
			}
		} else {
			if err := db.Get(&total, countQuery, args...); err != nil {
				return fmt.Errorf("failed to get total count: %w", err)
			}
		}

		selectQuery := fmt.Sprintf(`SELECT * FROM %s FINAL WHERE %s ORDER BY cm_mac_addr ASC LIMIT ? OFFSET ?`,
			t.GetDistributedTableName(), whereClause)
		selectArgs := append(args, filter.Limit, filter.Offset)

		if len(filter.ProgressStatus) > 0 {
			q, finalArgs, err := sqlx.In(selectQuery, selectArgs...)
			if err != nil {
				return fmt.Errorf("failed to build select IN query: %w", err)
			}
			return db.Select(&results, db.Rebind(q), finalArgs...)
		}

		return db.Select(&results, selectQuery, selectArgs...)
	}

	if err := orm.StatisticsHandler(t.config.Catv.Database.Driver, &t.config.Catv.Database, handler); err != nil {
		return 0, nil, fmt.Errorf("failed to select control schedule results with filter: %w", err)
	}

	return total, results, nil
}

func (t StbControlScheduleResultTable) Inserts(raws []StbControlScheduleResultRaw) error {
	if t.config.Catv.Control.Batch.MaxCount <= 0 {
		return fmt.Errorf("batch max count must be greater than 0")
	}

	var batch clickhouse_interface.ClickhouseBatch

	for index := range raws {
		switch raws[index].ResultCode {
		case control_common.ResponseCodeSuccess:
			raws[index].ProgressStatus = ProgressStatusSucceeded
		case control_common.ResponseCodeFailure:
			fallthrough
		case control_common.ResponseCodeError:
			raws[index].ProgressStatus = ProgressStatusFailed
		default:
		}

		if data, err := raws[index].ConvertBatchFormat(); err != nil {
			slog.Error("Failed to convert schedule result to batch format",
				"cm_mac_addr", raws[index].CmMacAddr,
				"stb_mac_addr", raws[index].StbMacAddr,
				"error_type", fmt.Sprintf("%T", err))
			return err
		} else if err := batch.AddBatchJson(data); err != nil {
			slog.Error("Failed to add schedule result to batch",
				"cm_mac_addr", raws[index].CmMacAddr,
				"stb_mac_addr", raws[index].StbMacAddr,
				"error_type", fmt.Sprintf("%T", err))
			return err
		}

		if (index+1)%t.config.Catv.Control.Batch.MaxCount != 0 &&
			batch.Len() < t.config.Catv.Control.Batch.MaxSize &&
			index != len(raws)-1 {
			continue
		}
		if err := batch.Insert(t.config.Catv.Database, t.GetDistributedTableName()); err != nil {
			slog.Error("Failed to insert schedule result batch", "error_type", fmt.Sprintf("%T", err))
			return err
		}
		batch = clickhouse_interface.ClickhouseBatch{}
	}

	return nil
}
