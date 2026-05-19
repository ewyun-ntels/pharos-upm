package common

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/external/orm"
	pkgcommon "ntels.com/pharos/core/pkg/common"
	notifcommon "ntels.com/pharos/core/pkg/notification/common"
)

// createStatsSQLiteConfig creates a temporary SQLite database config for statistics
func createStatsSQLiteConfig(t *testing.T) (pkgcommon.Config, string) {
	t.Helper()
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "stats.db")

	cfg := pkgcommon.Config{}
	cfg.Statistics.Database = orm.DatabaseConfig{
		Driver: orm.DriverSqlite,
		SQLite: orm.SQLiteConfig{Path: dbPath},
	}
	return cfg, dbPath
}

// prepareNotificationHistoryTable ensures the table exists in the SQLite DB
func prepareNotificationHistoryTable(t *testing.T, cfg *pkgcommon.Config) {
	t.Helper()
	require.NoError(t, orm.StatisticsHandler(orm.DriverDefault, &cfg.Statistics.Database, func(db *sqlx.DB) error {
		// minimal schema compatible with inserts from History.Send
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS history_notification (
			timestamp DATETIME,
			id TEXT,
			alert_id TEXT,
			name TEXT,
			alert_type TEXT,
			description TEXT,
			severity TEXT,
			value REAL,
			labels TEXT,
			status TEXT
		)`)
		return err
	}))
}

func TestHistory_Send_SQLite_InsertsRows(t *testing.T) {
	cfg, _ := createStatsSQLiteConfig(t)
	prepareNotificationHistoryTable(t, &cfg)

	h := History{Config: cfg}

	base := time.Now().UTC().Truncate(time.Second)
	vals := []notifcommon.AlertValue{
		{
			Name:        "rule1",
			Description: "d1",
			AlertId:     "aid-1",
			AlertType:   "type1",
			Value:       1.1,
			Severity:    "major",
			Status:      "firing",
			Timestamp:   base,
			UpdatedAt:   base,
			Labels:      map[string]string{"k": "v1"},
		},
		{
			Name:        "rule2",
			Description: "d2",
			AlertId:     "aid-2",
			AlertType:   "type2",
			Value:       2.2,
			Severity:    "minor",
			Status:      "resolved",
			Timestamp:   base.Add(time.Second),
			UpdatedAt:   base.Add(time.Second),
			Labels:      map[string]string{"k": "v2"},
		},
	}

	require.NoError(t, h.Send(vals))

	// Verify count and a couple of fields
	require.NoError(t, orm.StatisticsHandler(orm.DriverDefault, &cfg.Statistics.Database, func(db *sqlx.DB) error {
		var count int
		if err := db.Get(&count, "SELECT COUNT(*) FROM history_notification"); err != nil {
			return err
		}
		assert.Equal(t, len(vals), count)

		// Verify one row by name and alert_id
		var got struct {
			Name    string `json:"name" db:"name"`
			AlertID string `db:"alert_id"`
			Status  string `db:"status"`
		}
		if err := db.Get(&got, "SELECT name, alert_id, status FROM history_notification WHERE alert_id=?", "aid-1"); err != nil {
			return err
		}
		assert.Equal(t, "rule1", got.Name)
		assert.Equal(t, "aid-1", got.AlertID)
		assert.Equal(t, "firing", got.Status)
		return nil
	}))
}

func TestHistory_Send_Empty_NoOp(t *testing.T) {
	cfg, _ := createStatsSQLiteConfig(t)
	prepareNotificationHistoryTable(t, &cfg)
	h := History{Config: cfg}
	// empty slice
	require.NoError(t, h.Send([]notifcommon.AlertValue{}))
	// nil slice
	require.NoError(t, h.Send(nil))

	require.NoError(t, orm.StatisticsHandler(orm.DriverDefault, &cfg.Statistics.Database, func(db *sqlx.DB) error {
		var count int
		if err := db.Get(&count, "SELECT COUNT(*) FROM history_notification"); err != nil {
			return err
		}
		assert.Equal(t, 0, count)
		return nil
	}))
}
