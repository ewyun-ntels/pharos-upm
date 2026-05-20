package common

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/external/orm"
	alertcommon "ntels.com/pharos/core/pkg/alert/common"
	pkgcommon "ntels.com/pharos/core/pkg/common"
)

// createStatsSQLiteConfig creates a temporary SQLite statistics database config
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

// prepareAlertHistoryTables creates minimal tables used by History.Send
func prepareAlertHistoryTables(t *testing.T, cfg *pkgcommon.Config) {
	t.Helper()
	require.NoError(t, orm.StatisticsHandler(orm.DriverDefault, &cfg.Statistics.Database, func(db *sqlx.DB) error {
		// history_alert with snake_case columns
		if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS history_alert (
			timestamp DATETIME,
			id TEXT,
			alert_id TEXT,
			name TEXT,
			alert_type TEXT,
			description TEXT,
			previous_severity TEXT,
			previous_value REAL,
			severity TEXT,
			value REAL,
			labels TEXT,
			status TEXT,
			previous_timestamp DATETIME,
			version DATETIME
		)`); err != nil {
			return err
		}
		// unique index to support ON CONFLICT(id)
		if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_history_alert_id ON history_alert (id)`); err != nil {
			return err
		}

		// history_alert_row with snake_case columns
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS history_alert_row (
			timestamp DATETIME,
			id TEXT,
			alert_id TEXT,
			name TEXT,
			alert_type TEXT,
			description TEXT,
			previous_severity TEXT,
			previous_value REAL,
			severity TEXT,
			value REAL,
			labels TEXT,
			status TEXT,
			previous_timestamp DATETIME,
			version DATETIME,
			start_timestamp DATETIME,
			status_change_reason TEXT,
			status_changed_by TEXT,
			mask BOOLEAN,
			evaluation_epoch BIGINT
		)`)
		return err
	}))
}

func TestHistory_Send_SQLite_InsertsAndFields(t *testing.T) {
	cfg, _ := createStatsSQLiteConfig(t)
	prepareAlertHistoryTables(t, &cfg)

	h := History{Config: cfg}

	base := time.Now().UTC().Truncate(time.Second)
	vals := []alertcommon.Value{
		{
			Id:           "id-1",
			Name:         "rule1",
			Description:  "d1",
			AlertId:      "aid-1",
			AlertType:    string(alertcommon.TypeQuery),
			Value:        11,
			Severity:     SeverityMajor,
			Status:       alertcommon.StatusAlerting,
			Timestamp:    base,
			UpdatedAt:    base,
			StringLabels: `{"k":"v1"}`,
		},
		{
			Id:                "id-2",
			Name:              "rule2",
			Description:       "d2",
			AlertId:           "aid-2",
			AlertType:         string(alertcommon.TypeQuery),
			Value:             22,
			Severity:          SeverityMinor,
			Status:            alertcommon.StatusNormal,
			Timestamp:         base.Add(time.Second),
			UpdatedAt:         base.Add(time.Second),
			StringLabels:      `{"k":"v2"}`,
			PreviousSeverity:  SeverityMajor,
			PreviousValue:     func() *float64 { v := 11.0; return &v }(),
			PreviousTimestamp: base,
		},
	}

	user := "tester"
	require.NoError(t, h.Send(vals, alertcommon.StatusChangeReasonAuto, &user))

	// Verify row counts and some fields
	require.NoError(t, orm.StatisticsHandler(orm.DriverDefault, &cfg.Statistics.Database, func(db *sqlx.DB) error {
		var cntHist int
		if err := db.Get(&cntHist, "SELECT COUNT(*) FROM history_alert"); err != nil {
			return err
		}
		assert.Equal(t, len(vals), cntHist)

		var cntRow int
		if err := db.Get(&cntRow, "SELECT COUNT(*) FROM history_alert_row"); err != nil {
			return err
		}
		assert.Equal(t, len(vals), cntRow)

		// Check one row fields in history_alert
		var got struct {
			Name     string `db:"name"`
			AlertID  string `db:"alert_id"`
			Status   string `db:"status"`
			Severity string `db:"severity"`
		}
		if err := db.Get(&got, db.Rebind("SELECT name, alert_id, status, severity FROM history_alert WHERE id=?"), "id-1"); err != nil {
			return err
		}
		assert.Equal(t, "rule1", got.Name)
		assert.Equal(t, "aid-1", got.AlertID)
		assert.Equal(t, alertcommon.StatusAlerting, got.Status)
		assert.Equal(t, SeverityMajor, got.Severity)

		// Check status change fields exist in row table
		var row struct {
			StatusChangeReason string  `db:"status_change_reason"`
			StatusChangedBy    *string `db:"status_changed_by"`
		}
		if err := db.Get(&row, db.Rebind("SELECT status_change_reason, status_changed_by FROM history_alert_row WHERE id=?"), "id-2"); err != nil {
			return err
		}
		assert.Equal(t, alertcommon.StatusChangeReasonAuto, row.StatusChangeReason)
		assert.NotNil(t, row.StatusChangedBy)
		assert.Equal(t, &user, row.StatusChangedBy)
		return nil
	}))
}

func TestHistory_Send_SQLite_UpsertByIDVersion(t *testing.T) {
	cfg, _ := createStatsSQLiteConfig(t)
	prepareAlertHistoryTables(t, &cfg)
	h := History{Config: cfg}

	base := time.Now().UTC().Truncate(time.Second)
	// first send
	v1 := alertcommon.Value{
		Id:           "same-id",
		Name:         "rule",
		Description:  "first",
		AlertId:      "aid-x",
		AlertType:    string(alertcommon.TypeQuery),
		Value:        1,
		Severity:     SeverityMinor,
		Status:       alertcommon.StatusAlerting,
		Timestamp:    base,
		UpdatedAt:    base,
		StringLabels: `{"k":"v1"}`,
	}
	user := "tester"
	require.NoError(t, h.Send([]alertcommon.Value{v1}, alertcommon.StatusChangeReasonManual, &user))

	// Sleep over 1s to ensure rec.Version (seconds) is greater for the second send
	time.Sleep(1100 * time.Millisecond)

	// second send with changed fields and PreviousTimestamp set to first timestamp
	v2 := alertcommon.Value{
		Id:                "same-id",
		Name:              "rule",
		Description:       "second",
		AlertId:           "aid-x",
		AlertType:         string(alertcommon.TypeQuery),
		Value:             2,
		Severity:          SeverityMajor,
		Status:            alertcommon.StatusNormal,
		Timestamp:         base.Add(time.Second),
		UpdatedAt:         base.Add(time.Second),
		StringLabels:      `{"k":"v2"}`,
		PreviousSeverity:  SeverityMinor,
		PreviousValue:     func() *float64 { v := 1.0; return &v }(),
		PreviousTimestamp: base,
	}
	require.NoError(t, h.Send([]alertcommon.Value{v2}, alertcommon.StatusChangeReasonAuto, &user))

	// Validate upsert and row append behavior
	require.NoError(t, orm.StatisticsHandler(orm.DriverDefault, &cfg.Statistics.Database, func(db *sqlx.DB) error {
		var cntHist int
		if err := db.Get(&cntHist, "SELECT COUNT(*) FROM history_alert"); err != nil {
			return err
		}
		assert.Equal(t, 1, cntHist)

		var cntRow int
		if err := db.Get(&cntRow, "SELECT COUNT(*) FROM history_alert_row"); err != nil {
			return err
		}
		assert.Equal(t, 2, cntRow)

		// The history_alert should reflect second send values
		var got struct {
			Description string  `db:"description"`
			Severity    string  `db:"severity"`
			Value       float64 `db:"value"`
		}
		if err := db.Get(&got, db.Rebind("SELECT description, severity, value FROM history_alert WHERE id=?"), "same-id"); err != nil {
			return err
		}
		assert.Equal(t, "second", got.Description)
		assert.Equal(t, SeverityMajor, got.Severity)
		assert.Equal(t, 2.0, got.Value)
		return nil
	}))
}

func TestHistory_Send_SQLite_DeduplicatesAutoRowsByEvaluationEpoch(t *testing.T) {
	cfg, _ := createStatsSQLiteConfig(t)
	prepareAlertHistoryTables(t, &cfg)
	h := History{Config: cfg}

	require.NoError(t, orm.StatisticsHandler(orm.DriverDefault, &cfg.Statistics.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_history_alert_row_auto_epoch
			ON history_alert_row (id, status, severity, evaluation_epoch)
			WHERE evaluation_epoch IS NOT NULL AND status_change_reason = 'auto'`)
		return err
	}))

	checkTime := time.Unix(120, 0).UTC()
	v := alertcommon.Value{
		Id:           "same-alert",
		Name:         "rule",
		Description:  "dedup",
		AlertId:      "aid-dedup",
		AlertType:    string(alertcommon.TypeQuery),
		Value:        1,
		Severity:     SeverityMinor,
		Status:       alertcommon.StatusAlerting,
		Timestamp:    checkTime.Add(3 * time.Second),
		CheckTime:    &checkTime,
		UpdatedAt:    checkTime,
		StringLabels: `{"k":"v"}`,
	}
	dedup := &HistoryDedup{EvaluationIntervalSeconds: 60}

	require.NoError(t, h.Send([]alertcommon.Value{v}, alertcommon.StatusChangeReasonAuto, nil, dedup))
	require.NoError(t, h.Send([]alertcommon.Value{v}, alertcommon.StatusChangeReasonAuto, nil, dedup))

	require.NoError(t, orm.StatisticsHandler(orm.DriverDefault, &cfg.Statistics.Database, func(db *sqlx.DB) error {
		var cntRow int
		if err := db.Get(&cntRow, "SELECT COUNT(*) FROM history_alert_row"); err != nil {
			return err
		}
		assert.Equal(t, 1, cntRow)
		return nil
	}))
}

func TestHistory_Send_Empty_NoOp(t *testing.T) {
	cfg, _ := createStatsSQLiteConfig(t)
	prepareAlertHistoryTables(t, &cfg)
	h := History{Config: cfg}
	require.NoError(t, h.Send(nil, alertcommon.StatusChangeReasonAuto, nil))
	require.NoError(t, h.Send([]alertcommon.Value{}, alertcommon.StatusChangeReasonAuto, nil))

	require.NoError(t, orm.StatisticsHandler(orm.DriverDefault, &cfg.Statistics.Database, func(db *sqlx.DB) error {
		var cnt1, cnt2 int
		if err := db.Get(&cnt1, "SELECT COUNT(*) FROM history_alert"); err != nil {
			return err
		}
		if err := db.Get(&cnt2, "SELECT COUNT(*) FROM history_alert_row"); err != nil {
			return err
		}
		assert.Equal(t, 0, cnt1)
		assert.Equal(t, 0, cnt2)
		return nil
	}))
}
