package tables

import (
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

// allowedOrderBy 는 GetAll 의 orderBy 인자로 허용된 값의 집합입니다.
// 컬럼 이름이 직접 쿼리에 삽입되므로 화이트리스트로만 허용합니다.
var allowedOrderBy = map[string]struct{}{
	"name ASC":         {},
	"name DESC":        {},
	"create_time ASC":  {},
	"create_time DESC": {},
	"update_time ASC":  {},
	"update_time DESC": {},
}

const distStbControlScheduleTableName = "dist_stb_control_schedule"

type StbControlScheduleTable struct {
	config common.Config
}

func NewStbControlScheduleTable(config common.Config) *StbControlScheduleTable {
	return &StbControlScheduleTable{
		config: config,
	}
}

func (r StbControlScheduleTable) ExistsByName(name string) (bool, error) {
	var exists bool
	handler := func(db *sqlx.DB) error {
		return db.Get(&exists, `SELECT EXISTS(SELECT 1 FROM `+distStbControlScheduleTableName+` FINAL WHERE name = ? AND is_deleted = 0);`, name)
	}
	if err := orm.StatisticsHandler(r.config.Catv.Database.Driver, &r.config.Catv.Database, handler); err != nil {
		return false, err
	}

	return exists, nil
}

func (r *StbControlScheduleTable) Create(raw StbControlScheduleRaw) error {
	if id, err := uuid.NewV7(); err != nil {
		return err
	} else {
		raw.ID = id.String()
	}

	return r.Update(raw)
}

func (r StbControlScheduleTable) GetByName(name string) (StbControlScheduleRaw, error) {
	var raw StbControlScheduleRaw
	handler := func(db *sqlx.DB) error {
		return db.Get(&raw, `SELECT * FROM `+distStbControlScheduleTableName+` FINAL WHERE name = ? AND is_deleted = 0;`, name)
	}
	if err := orm.StatisticsHandler(r.config.Catv.Database.Driver, &r.config.Catv.Database, handler); err != nil {
		return StbControlScheduleRaw{}, err
	}

	return raw, nil
}

func (r *StbControlScheduleTable) GetAll(orderBy string) ([]StbControlScheduleRaw, error) {
	var raws []StbControlScheduleRaw
	handler := func(db *sqlx.DB) error {
		query := `SELECT * FROM ` + distStbControlScheduleTableName + ` FINAL WHERE is_deleted = 0`
		if orderBy != "" {
			if _, ok := allowedOrderBy[orderBy]; !ok {
				return fmt.Errorf("invalid orderBy value: %q", orderBy)
			}
			query += " ORDER BY " + orderBy
		}
		return db.Select(&raws, query)
	}
	if err := orm.StatisticsHandler(r.config.Catv.Database.Driver, &r.config.Catv.Database, handler); err != nil {
		return nil, err
	}

	return raws, nil
}

func (r StbControlScheduleTable) Update(raw StbControlScheduleRaw) error {
	handler := func(db *sqlx.DB) error {
		var query = `INSERT INTO ` + distStbControlScheduleTableName +
			`(id, name, schedule_type, schedule_spec_once, schedule_spec_repeat, sos, l3s, cells, settopboxes, work_type, work_value, area_type, area_ids, create_time, update_time, is_deleted) ` +
			`VALUES (:id, :name, :schedule_type, :schedule_spec_once, :schedule_spec_repeat, :sos, :l3s, :cells, :settopboxes, :work_type, :work_value, :area_type, :area_ids, :create_time, :update_time, :is_deleted);`

		// Use map to safely handle nil pointers (prevents panic when calling Value() on nil *orm.Datetime)
		params := map[string]any{
			"id":                   raw.ID,
			"name":                 raw.Name,
			"schedule_type":        raw.ScheduleType,
			"schedule_spec_once":   nil,
			"schedule_spec_repeat": nil,
			"sos":                  raw.Sos,
			"l3s":                  raw.L3s,
			"cells":                raw.Cells,
			"settopboxes":          raw.Settopboxes,
			"work_type":            raw.WorkType,
			"work_value":           nil,
			"area_type":            raw.AreaType,
			"area_ids":             raw.AreaIDs,
			"create_time":          raw.CreateTime,
			"update_time":          orm.Datetime{Time: time.Now().UTC()},
			"is_deleted":           raw.IsDeleted,
		}

		if raw.ScheduleSpecOnce != nil {
			params["schedule_spec_once"] = raw.ScheduleSpecOnce.Time
		}
		if raw.ScheduleSpecRepeat != nil {
			params["schedule_spec_repeat"] = *raw.ScheduleSpecRepeat
		}
		if raw.WorkValue != nil {
			params["work_value"] = *raw.WorkValue
		}

		_, err := db.NamedExec(query, params)
		return err
	}
	return orm.StatisticsHandler(r.config.Catv.Database.Driver, &r.config.Catv.Database, handler)
}

func (r *StbControlScheduleTable) Delete(name string) error {
	controlScheduleRequest, err := r.GetByName(name)
	if err != nil {
		return err
	}

	controlScheduleRequest.IsDeleted = 1

	return r.Update(controlScheduleRequest)
}
