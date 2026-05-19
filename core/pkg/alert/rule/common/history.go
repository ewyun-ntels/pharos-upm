package common

import (
	"encoding/json"
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

func (h *History) Send(alerts []alert_common.Value, statusChangeReason string, statusChangeBy *string) error {
	if len(alerts) == 0 {
		return nil
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
	}

	records := make([]BatchHist, 0, len(alerts))

	currentTime := time.Now().Unix()
	for _, alert := range alerts {
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

		// history_alert_row remains simple insert across all drivers
		insertRow := "INSERT INTO history_alert_row (timestamp, id, alert_id, name, alert_type, description, previous_severity, previous_value, severity, value, labels, status, previous_timestamp, version, start_timestamp, status_change_reason, status_changed_by, mask) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)"

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
			if _, err := db.Exec(qRow, ts, rec.Id, rec.AlertId, rec.Name, rec.AlertType, rec.Description, rec.PreviousSeverity, rec.PreviousValue, rec.Severity, rec.Value, rec.Labels, rec.Status, prevTS, ver, startTS, rec.StatusChangeReason, rec.StatusChangedBy, rec.Mask); err != nil {
				return err
			}
		}

		return nil
	})
}
