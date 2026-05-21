package common

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	alert_common "ntels.com/pharos/core/pkg/alert/common"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/websocket"
	"ntels.com/pharos/core/pkg/websocket/centrifuge/node/handler"
)

type History struct {
	Config common.Config
}

type HistoryDedup struct {
	EvaluationIntervalSeconds int64
}

func (h *History) Send(alerts []alert_common.Value, statusChangeReason string, statusChangeBy *string, dedupOptions ...*HistoryDedup) error {
	if len(alerts) == 0 {
		return nil
	}

	var dedup *HistoryDedup
	if len(dedupOptions) > 0 {
		dedup = dedupOptions[0]
	}

	type BatchHist struct {
		Timestamp          int64
		StartTimestamp     int64
		Id                 string
		Name               string
		Description        string
		AlertId            string
		AlertType          string
		Value              float64
		Severity           string
		Status             string
		Mask               bool
		Labels             string
		PreviousSeverity   string
		PreviousValue      *float64
		PreviousTimestamp  *int64
		Version            int64
		StatusChangeReason string
		StatusChangedBy    *string
		EvaluationEpoch    *int64
		DedupKey           string
	}

	records := make([]BatchHist, 0, len(alerts))

	now := time.Now()
	currentTime := now.Unix()
	batchID := now.UnixNano()
	for idx, alert := range alerts {
		bytes, err := json.Marshal(alert)
		if err != nil {
			slog.Error("json Marshal error", "error", err, "alert", alert)
			continue
		}

		startTimestamp := alert.Timestamp.Unix()
		var prevTimestamp *int64
		if !alert.PreviousTimestamp.IsZero() {
			prevUnix := alert.PreviousTimestamp.Unix()
			prevTimestamp = &prevUnix
			startTimestamp = prevUnix
		}

		var evaluationEpoch *int64
		if dedup != nil && dedup.EvaluationIntervalSeconds > 0 {
			evaluationTime := alert.Timestamp
			if alert.CheckTime != nil && !alert.CheckTime.IsZero() {
				evaluationTime = *alert.CheckTime
			}
			if !evaluationTime.IsZero() {
				epoch := evaluationTime.Unix() / dedup.EvaluationIntervalSeconds
				evaluationEpoch = &epoch
			}
		}

		dedupKey := fmt.Sprintf("row:%s:%d:%d:%d", alert.Id, alert.Timestamp.Unix(), batchID, idx)
		if evaluationEpoch != nil && statusChangeReason == alert_common.StatusChangeReasonAuto {
			dedupKey = fmt.Sprintf("auto:%s:%s:%s:%d", alert.Id, alert.Status, alert.Severity, *evaluationEpoch)
		}

		records = append(records, BatchHist{
			Timestamp:          alert.Timestamp.Unix(),
			StartTimestamp:     startTimestamp,
			Id:                 alert.Id,
			Name:               alert.Name,
			Description:        alert.Description,
			AlertId:            alert.AlertId,
			AlertType:          alert.AlertType,
			Value:              alert.Value,
			Status:             alert.Status,
			Severity:           alert.Severity,
			Mask:               alert.Mask,
			Labels:             alert.StringLabels,
			PreviousSeverity:   alert.PreviousSeverity,
			PreviousValue:      alert.PreviousValue,
			PreviousTimestamp:  prevTimestamp,
			Version:            currentTime,
			StatusChangeReason: statusChangeReason,
			StatusChangedBy:    statusChangeBy,
			EvaluationEpoch:    evaluationEpoch,
			DedupKey:           dedupKey,
		})

		for _, n := range websocket.Nodes {
			if err := n.Publish(handler.ChannelAlert, bytes); err != nil {
				slog.Error("publish error", "error", err, "bytes", string(bytes))
				continue
			}
		}
	}

	// Use statistics ORM pool to persist records into statistics database
	statsDB := &h.Config.Statistics.Database
	if statsDB == nil || statsDB.Driver == "" {
		// no statistics database configured
		return nil
	}

	return orm.StatisticsHandler("", statsDB, func(db *sqlx.DB) error {
		// Build unified SQL for SQLite and PostgreSQL; ClickHouse uses plain INSERT.
		insertHist := "INSERT INTO history_alert (timestamp, id, alert_id, name, alert_type, description, previous_severity, previous_value, severity, value, labels, status, previous_timestamp, version) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
		if statsDB.Driver != orm.DriverClickHouse {
			// Emulate ReplacingMergeTree: upsert by id only when incoming version is newer
			insertHist += " ON CONFLICT(id) DO UPDATE SET timestamp=excluded.timestamp, alert_id=excluded.alert_id, name=excluded.name, alert_type=excluded.alert_type, description=excluded.description, previous_severity=excluded.previous_severity, previous_value=excluded.previous_value, severity=excluded.severity, value=excluded.value, labels=excluded.labels, status=excluded.status, previous_timestamp=excluded.previous_timestamp, version=excluded.version WHERE history_alert.version IS NULL OR excluded.version > history_alert.version"
		}

		// history_alert_row is append-only, with optional query-evaluation dedup for multi-instance runs.
		insertRow := "INSERT INTO history_alert_row (timestamp, id, alert_id, name, alert_type, description, previous_severity, previous_value, severity, value, labels, status, previous_timestamp, version, start_timestamp, status_change_reason, status_changed_by, mask, evaluation_epoch) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
		if statsDB.Driver == orm.DriverClickHouse {
			insertRow = "INSERT INTO history_alert_row (timestamp, id, alert_id, name, alert_type, description, previous_severity, previous_value, severity, value, labels, status, previous_timestamp, version, start_timestamp, status_change_reason, status_changed_by, mask, evaluation_epoch, dedup_key) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"
		} else {
			insertRow += " ON CONFLICT DO NOTHING"
		}

		qHist := db.Rebind(insertHist)
		qRow := db.Rebind(insertRow)

		for _, rec := range records {
			// Convert timestamps to time.Time for DBs
			ts := time.Unix(rec.Timestamp, 0)
			ver := time.Unix(rec.Version, 0)
			var prevTS any
			if rec.PreviousTimestamp != nil {
				prevTS = time.Unix(*rec.PreviousTimestamp, 0)
			} else {
				prevTS = nil
			}
			startTS := time.Unix(rec.StartTimestamp, 0)

			if _, err := db.Exec(qHist, ts, rec.Id, rec.AlertId, rec.Name, rec.AlertType, rec.Description, rec.PreviousSeverity, rec.PreviousValue, rec.Severity, rec.Value, rec.Labels, rec.Status, prevTS, ver); err != nil {
				return err
			}
			if statsDB.Driver == orm.DriverClickHouse {
				if _, err := db.Exec(qRow, ts, rec.Id, rec.AlertId, rec.Name, rec.AlertType, rec.Description, rec.PreviousSeverity, rec.PreviousValue, rec.Severity, rec.Value, rec.Labels, rec.Status, prevTS, ver, startTS, rec.StatusChangeReason, rec.StatusChangedBy, rec.Mask, rec.EvaluationEpoch, rec.DedupKey); err != nil {
					return err
				}
			} else {
				if _, err := db.Exec(qRow, ts, rec.Id, rec.AlertId, rec.Name, rec.AlertType, rec.Description, rec.PreviousSeverity, rec.PreviousValue, rec.Severity, rec.Value, rec.Labels, rec.Status, prevTS, ver, startTS, rec.StatusChangeReason, rec.StatusChangedBy, rec.Mask, rec.EvaluationEpoch); err != nil {
					return err
				}
			}
		}

		return nil
	})
}
