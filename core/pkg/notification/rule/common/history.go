package common

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	notification_common "ntels.com/pharos/core/pkg/notification/common"
)

type History struct {
	Config common.Config
}

func (h *History) Send(alerts []notification_common.AlertValue) error {
	if len(alerts) == 0 {
		return nil
	}

	// If statistics DB is not configured, do nothing
	if h.Config.Statistics.Database.Driver == "" {
		return nil
	}

	statsDB := &h.Config.Statistics.Database

	// Use a single SQL with '?' placeholders and let sqlx rebind for the driver
	insertSQL := "INSERT INTO history_notification (timestamp, id, alert_id, name, alert_type, description, severity, value, labels, status) VALUES (?,?,?,?,?,?,?,?,?,?)"

	return orm.StatisticsHandler("", statsDB, func(db *sqlx.DB) error {
		q := db.Rebind(insertSQL)

		for _, alert := range alerts {
			var labels []byte
			if len(alert.Labels) != 0 {
				b, err := json.Marshal(alert.Labels)
				if err != nil {
					slog.Error("json Marshal error", "error", err, "alert", alert)
					continue
				}
				labels = b
			}

			if _, err := db.Exec(
				q,
				time.Unix(alert.Timestamp.Unix(), 0),
				"", // Id not provided by AlertValue; keep empty for compatibility
				alert.AlertId,
				alert.Name,
				alert.AlertType,
				alert.Description,
				alert.Severity,
				alert.Value,
				string(labels),
				alert.Status,
			); err != nil {
				slog.Error("insert history_notification failed", "error", err)
				return err
			}
		}
		return nil
	})
}
