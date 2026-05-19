package model

import (
	"encoding/json"
	"errors"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	alert_common "ntels.com/pharos/core/pkg/alert/common"
	"ntels.com/pharos/core/pkg/common"
)

type Rule struct {
	ID        string       `db:"id"`
	AlertType string       `db:"alert_type"`
	Name      string       `db:"name"`
	Rule      string       `db:"rule"`
	Timestamp orm.Datetime `db:"timestamp"` // formating 문제로 인하여 string형식으로 변경
	UpdatedAt orm.Datetime `db:"updated_at"`
}

type Alert struct {
	Id               string       `db:"id"` // 알람 중복 삭제등을 위해 사용(alertId는 alert 구분(severity는 달라질수 있음)
	Timestamp        orm.Datetime `db:"timestamp"`
	UpdatedAt        orm.Datetime `db:"updated_at"`
	CheckTime        orm.Datetime `db:"check_time"`
	Name             string       `db:"name"`
	Mask             bool         `db:"mask"`
	AlertType        string       `db:"alert_type"`
	AlertId          string       `db:"alert_id"`
	Value            float64      `db:"value"`
	Status           string       `db:"status"`
	Severity         string       `db:"severity"`
	PreviousSeverity string       `db:"previous_severity"`
	Description      string       `db:"description"`
	PreviousValue    *float64     `db:"previous_value"`
	StringLabels     string       `db:"labels"`
}

type Model struct {
	Config *common.Config
}

func (m *Model) UpdateAlert(update *alert_common.Value) (err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		// TODO TIME FORMAT에 대해 확인 필요 응답과 insert가 다른 포맷으로 나옴
		alert := Alert{
			Id:               update.Id,
			Timestamp:        orm.Datetime{Time: update.Timestamp.UTC()},
			UpdatedAt:        orm.Datetime{Time: update.Timestamp.UTC()},
			CheckTime:        orm.Datetime{Time: update.CheckTime.UTC()},
			Name:             update.Name,
			Mask:             update.Mask,
			AlertType:        update.AlertType,
			Description:      update.Description,
			AlertId:          update.AlertId,
			Value:            update.Value,
			Status:           update.Status,
			Severity:         update.Severity,
			PreviousSeverity: update.PreviousSeverity,
			PreviousValue:    update.PreviousValue,
			StringLabels:     update.StringLabels,
		}
		_, err = db.NamedExec(`
	INSERT INTO alert_status (id, alert_id, name, mask, alert_type, description, status, severity, previous_severity, previous_value, value, labels, timestamp, updated_at, check_time)
	VALUES (:id, :alert_id, :name, :mask, :alert_type, :description, :status, :severity, :previous_severity, :previous_value, :value, :labels, :timestamp, :updated_at, :check_time)
	ON CONFLICT (alert_id, name)
	DO UPDATE SET
	id = EXCLUDED.id,
	severity = EXCLUDED.severity,
	alert_type = EXCLUDED.alert_type,
	previous_severity = EXCLUDED.previous_severity,
	previous_value = EXCLUDED.previous_value,
	value = EXCLUDED.value,
	status = EXCLUDED.status,
	labels = EXCLUDED.labels,
	timestamp = EXCLUDED.timestamp,
	updated_at = EXCLUDED.updated_at,
	check_time = EXCLUDED.check_time
	`, alert)
		if err != nil {
			slog.Error("db.insert failed", "error", err)
			return err
		}

		return nil
	})

	return
}

func (m *Model) GetAlert(name, alertId string) (alert *alert_common.Value, err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		var get []alert_common.Value
		err = db.Select(&get, `
	SELECT id, alert_id, status, severity, previous_severity, previous_value, name, mask, alert_type, description, value, labels, timestamp, updated_at, check_time
	FROM alert_status
	WHERE name = $1 AND alert_id = $2
	`, name, alertId)
		if err != nil {
			return err
		}

		if len(get) == 0 {
			return nil
		}

		if len(get) > 1 {
			return errors.New("too many alert found")
		}

		alert = &get[0]

		return nil
	})

	return
}

func (m *Model) GetAllAlert(name string) (alerts []alert_common.Value, err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		err = db.Select(&alerts, `
	SELECT id, alert_id, status, severity, previous_severity, previous_value, name, mask, alert_type, description, value, labels, timestamp, updated_at, check_time
	FROM alert_status
	WHERE name = $1
	`, name)
		if err != nil {
			slog.Error("alert select failed", "error", err)
			return err
		}

		return nil
	})

	return
}

func (m *Model) DeleteAlertById(name string, id string) (err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		_, err = db.Exec(`
	DELETE FROM alert_status WHERE name = $1 AND id = $2
	`, name, id)
		if err != nil {
			slog.Error("alert delete failed", "error", err)
			return err
		}

		return nil
	})

	return
}

func (m *Model) DeleteAlert(name, alertId string) (err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		_, err = db.Exec(`
	DELETE FROM alert_status WHERE alert_id = $1 AND name = $2
	`, alertId, name)
		if err != nil {
			slog.Error("alert delete failed", "error", err)
			return err
		}

		return nil
	})

	return
}

func (m *Model) DeleteOldAlerts(alerts *[]alert_common.Value) (err error) {
	// 관련 있는 alert만 제거하고 나머지는 디버그를 위해 일단 남겨두도록 구현
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		tx := db.MustBegin()

		for _, alert := range *alerts {
			tx.MustExec("DELETE FROM alert_status WHERE alert_id=$1", alert.AlertId)
		}

		err := tx.Commit()
		if err != nil {
			slog.Error("tx.Commit failed", "error", err)
			return err
		}

		return nil
	})

	return
}

func (m *Model) InsertRule(rule *Rule) (err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		_, err = db.NamedExec(`
	INSERT INTO alert_rule (id, name, alert_type, rule, timestamp, updated_at) VALUES
	(:id, :name, :alert_type, :rule, :timestamp, :updated_at);
	`, rule)
		if err != nil {
			slog.Error("rule insert failed", "error", err)
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return
}

func (m *Model) UpdateRule(rule *Rule) (err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		_, err = db.NamedExec(`
	UPDATE alert_rule SET
	id =:id,
	name =:name,
	alert_type =:alert_type,
	rule =:rule,
	updated_at =:updated_at
	WHERE id = :id
	`, rule)
		if err != nil {
			slog.Error("rule update failed", "error", err)
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return
}

func (m *Model) DeleteRule(id string) (err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		_, err = db.Exec(`
	DELETE FROM alert_rule WHERE id = $1
	`, id)
		if err != nil {
			slog.Error("rule delete failed", "error", err)
			return err
		}
		return nil
	})
	if err != nil {
		slog.Error("rule delete failed", "error", err)
		return err
	}

	return nil
}

// DeleteAlertsByName removes all current alert statuses for a given rule name.
func (m *Model) DeleteAlertsByName(name string) (err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		_, err = db.Exec(`
	DELETE FROM alert_status WHERE name = $1
	`, name)
		if err != nil {
			slog.Error("alerts delete by name failed", "error", err)
			return err
		}
		return nil
	})
	return
}

func (m *Model) GetRuleByName(name string) (rule *Rule, err error) {
	var res []Rule
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		err = db.Select(&res, `SELECT * FROM alert_rule WHERE name = $1`, name)
		if err != nil {
			return err
		}

		return nil
	})

	if len(res) == 0 {
		return nil, err
	}

	if len(res) > 1 {
		return nil, errors.New("too many results")
	}

	return &res[0], err
}

func (m *Model) GetRule(id string) (rule *Rule, err error) {
	var res []Rule
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		err = db.Select(&res, `SELECT * FROM alert_rule WHERE id = $1`, id)
		if err != nil {
			slog.Error("rule get failed", "error", err)
			return err
		}

		return nil
	})

	if len(res) == 0 {
		return nil, err
	}

	if len(res) > 1 {
		return nil, errors.New("too many results")
	}

	return &res[0], err
}

func (m *Model) GetAllRuleDb() (ruleDbs []Rule, err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		err = db.Select(&ruleDbs, `SELECT * FROM alert_rule`)
		if err != nil {
			slog.Error("rule list load failed", "error", err)
			return err
		}

		return nil
	})

	return
}

type RuleHistDbStructure struct {
	Timestamp        orm.Datetime `db:"timestamp"`
	UpdatedAt        orm.Datetime `db:"updated_at"`
	AlertId          string       `db:"alert_id"`
	Name             string       `db:"name"`
	AlertType        string       `db:"alert_type"`
	Description      string       `db:"description"`
	PreviousSeverity string       `db:"previous_severity"`
	PreviousValue    *float64     `db:"previous_value"`
	Severity         string       `db:"severity"`
	Value            float64      `db:"value"`
	Labels           string       `db:"labels"`
}

func (m *Model) InsertRuleHist(alerts []alert_common.Value) (err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		tx := db.MustBegin()
		for _, alert := range alerts {
			hist := RuleHistDbStructure{
				Timestamp:        orm.Datetime{Time: alert.Timestamp.UTC()},
				UpdatedAt:        orm.Datetime{Time: alert.Timestamp.UTC()},
				AlertId:          alert.AlertId,
				Name:             alert.Name,
				AlertType:        alert.AlertType,
				Description:      alert.Description,
				Severity:         alert.Severity,
				PreviousSeverity: alert.PreviousSeverity,
				PreviousValue:    alert.PreviousValue,
				Value:            alert.Value,
				Labels:           alert.StringLabels,
			}
			_, err = tx.NamedExec(`
	INSERT INTO alert_hist (timestamp, updated_at, alert_id, name, alert_type, description, previous_severity, previous_value, severity, value, labels) VALUES
	(:timestamp, :updated_at, :alert_id, :name, :alert_type, :description, :previous_severity, :previous_value, :severity, :value, :labels)
`, hist)
			if err != nil {
				slog.Error("rule insert failed", "error",
					err)

				err := tx.Rollback()
				if err != nil {
					slog.Error("tx.Rollback failed", "error", err)
					return err
				}
				return err
			}
		}

		err := tx.Commit()
		if err != nil {
			slog.Error("tx.Commit failed", "error", err)
			return err
		}
		return nil
	})

	return
}

func (m *Model) QueryRuleHist(ruleName string, startTime, endTime orm.Datetime, count int) (value []alert_common.Value, err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		query := `SELECT * FROM alert_hist WHERE 1=1`
		if ruleName != "" {
			query += ` AND name = $1`
		}
		if !startTime.IsZero() && !endTime.IsZero() {
			query += ` AND timestamp BETWEEN $2 AND $3`
		}

		query += ` ORDER BY timestamp DESC`

		if count > 0 {
			query += ` LIMIT $4`
		}

		result := []RuleHistDbStructure{}
		err := db.Select(&result, query,
			ruleName,
			startTime, endTime,
			count)
		if err != nil {
			slog.Error("rule hist query failed", "error", err)
			return err
		}

		for _, v := range result {
			// TODO time parsing 공식 확인 필요 insert는 dateTime으로 했는데 응답은 RFC3339형식으로 반환
			//timestamp, _ := time.Parse(time.RFC3339, v.Timestamp)
			//updateAt, _ := time.Parse(time.RFC3339, v.UpdatedAt)
			var labels map[string]string
			if v.Labels != "" {
				err = json.Unmarshal([]byte(v.Labels), &labels)
				if err != nil {
					slog.Error("rule hist parsing failed(label parsing error)", "error", err)
					return err
				}
			}

			value = append(value, alert_common.Value{
				Timestamp:        v.Timestamp.Time,
				UpdatedAt:        v.Timestamp.Time,
				Name:             v.Name,
				AlertType:        v.AlertType,
				AlertId:          v.AlertId,
				Value:            v.Value,
				Description:      v.Description,
				Severity:         v.Severity,
				PreviousSeverity: v.PreviousSeverity,
				PreviousValue:    v.PreviousValue,
				Labels:           labels,
			})
		}
		return nil
	})

	return
}

func (m *Model) GetStatuses() (*[]alert_common.Value, error) {
	var value []alert_common.Value
	err := orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		err := db.Select(&value, "select * from alert_status")
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	for i := range value {
		err = value[i].ConvertStringLabelsToLabels()
		if err != nil {
			slog.Error("alert status convert string labels to labels failed", "error", err)
			continue
		}
	}

	return &value, nil
}

func (m *Model) GetStatus(name, alertId string) (*alert_common.Value, error) {
	var value []alert_common.Value
	err := orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		err := db.Select(&value, "select * from alert_status WHERE name=$1 AND alert_id=$2", name, alertId)
		if err != nil {
			return err
		}

		return nil
	})
	if err != nil {
		return nil, err
	}

	if len(value) == 0 {
		return nil, nil
	}

	if len(value) > 1 {
		return nil, errors.New("alert status query returned more than one result")
	}

	err = value[0].ConvertStringLabelsToLabels()
	if err != nil {
		return nil, err
	}

	return &value[0], nil
}

func (m *Model) SetStatusMask(name, alertId string, mask bool) (err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		_ = db.MustExec("UPDATE alert_status SET mask=$1 WHERE name = $2 AND alert_id = $3", mask, name, alertId)

		return nil
	})

	return
}
