package common

import (
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/external/orm"
	pkgcommon "ntels.com/pharos/core/pkg/common"
	notifcommon "ntels.com/pharos/core/pkg/notification/common"
)

// Integration test for PostgreSQL statistics database. Skips if env is not configured.
// Env vars: PHAROS_TEST_PG_HOST, PHAROS_TEST_PG_PORT, PHAROS_TEST_PG_USER, PHAROS_TEST_PG_PASS, PHAROS_TEST_PG_DB
func TestHistory_Send_PostgreSQL_InsertsRows(t *testing.T) {
	host := os.Getenv("PHAROS_TEST_PG_HOST")
	port := os.Getenv("PHAROS_TEST_PG_PORT")
	user := os.Getenv("PHAROS_TEST_PG_USER")
	pass := os.Getenv("PHAROS_TEST_PG_PASS")
	dbname := os.Getenv("PHAROS_TEST_PG_DB")
	if host == "" || user == "" || dbname == "" {
		t.Skip("PostgreSQL env not set; skipping")
	}
	pgPort := 5432
	if port != "" {
		if v, err := strconv.Atoi(port); err == nil {
			pgPort = v
		}
	}

	cfg := pkgcommon.Config{}
	cfg.Statistics.Database = orm.DatabaseConfig{
		Driver: orm.DriverPostgreSQL,
		PostgreSQL: orm.PostgreSQLConfig{
			Host:     host,
			Port:     pgPort,
			Username: user,
			Password: pass,
			Database: dbname,
		},
	}

	// Prepare schema (drop and create for a clean slate)
	require.NoError(t, orm.StatisticsHandler(orm.DriverDefault, &cfg.Statistics.Database, func(db *sqlx.DB) error {
		if _, err := db.Exec("DROP TABLE IF EXISTS history_notification"); err != nil {
			return err
		}
		_, err := db.Exec(`CREATE TABLE history_notification (
			timestamp TIMESTAMPTZ,
			id TEXT,
			alert_id TEXT,
			name TEXT,
			alert_type TEXT,
			description TEXT,
			severity TEXT,
			value DOUBLE PRECISION,
			labels TEXT,
			status TEXT
		)`)
		return err
	}))

	h := History{Config: cfg}
	base := time.Now().UTC().Truncate(time.Second)
	vals := []notifcommon.AlertValue{
		{
			Name:        "pg-rule1",
			Description: "d1",
			AlertId:     "pg-aid-1",
			AlertType:   "type1",
			Value:       1.23,
			Severity:    "major",
			Status:      "firing",
			Timestamp:   base,
			UpdatedAt:   base,
			Labels:      map[string]string{"k": "v1"},
		},
		{
			Name:        "pg-rule2",
			Description: "d2",
			AlertId:     "pg-aid-2",
			AlertType:   "type2",
			Value:       2.34,
			Severity:    "minor",
			Status:      "resolved",
			Timestamp:   base.Add(time.Second),
			UpdatedAt:   base.Add(time.Second),
			Labels:      map[string]string{"k": "v2"},
		},
	}
	require.NoError(t, h.Send(vals))

	require.NoError(t, orm.StatisticsHandler(orm.DriverDefault, &cfg.Statistics.Database, func(db *sqlx.DB) error {
		var count int
		if err := db.Get(&count, "SELECT COUNT(*) FROM history_notification"); err != nil {
			return err
		}
		assert.Equal(t, len(vals), count)
		return nil
	}))
}

// Integration test for ClickHouse statistics database. Skips if env is not configured.
// Env vars: PHAROS_TEST_CH_HOST, PHAROS_TEST_CH_PORT, PHAROS_TEST_CH_USER, PHAROS_TEST_CH_PASS, PHAROS_TEST_CH_DB
func TestHistory_Send_ClickHouse_InsertsRows(t *testing.T) {
	host := os.Getenv("PHAROS_TEST_CH_HOST")
	port := os.Getenv("PHAROS_TEST_CH_PORT")
	user := os.Getenv("PHAROS_TEST_CH_USER")
	pass := os.Getenv("PHAROS_TEST_CH_PASS")
	dbname := os.Getenv("PHAROS_TEST_CH_DB")
	if host == "" || user == "" || dbname == "" {
		t.Skip("ClickHouse env not set; skipping")
	}
	chPort := 9000
	if port != "" {
		if v, err := strconv.Atoi(port); err == nil {
			chPort = v
		}
	}

	cfg := pkgcommon.Config{}
	cfg.Statistics.Database = orm.DatabaseConfig{
		Driver: orm.DriverClickHouse,
		ClickHouse: orm.ClickHouseConfig{
			Host:     host,
			Port:     chPort,
			Username: user,
			Password: pass,
			Database: dbname,
		},
	}

	// Prepare schema: drop then create MergeTree table
	require.NoError(t, orm.StatisticsHandler(orm.DriverDefault, &cfg.Statistics.Database, func(db *sqlx.DB) error {
		if _, err := db.Exec("DROP TABLE IF EXISTS history_notification"); err != nil {
			return err
		}
		_, err := db.Exec(`CREATE TABLE history_notification (
			timestamp DateTime,
			id String,
			alert_id String,
			name String,
			alert_type String,
			description String,
			severity String,
			value Float64,
			labels String,
			status String
		) ENGINE = MergeTree() ORDER BY (timestamp, alert_id)`)
		return err
	}))

	h := History{Config: cfg}
	base := time.Now().UTC().Truncate(time.Second)
	vals := []notifcommon.AlertValue{
		{
			Name:        "ch-rule1",
			Description: "d1",
			AlertId:     "ch-aid-1",
			AlertType:   "type1",
			Value:       1.23,
			Severity:    "major",
			Status:      "firing",
			Timestamp:   base,
			UpdatedAt:   base,
			Labels:      map[string]string{"k": "v1"},
		},
		{
			Name:        "ch-rule2",
			Description: "d2",
			AlertId:     "ch-aid-2",
			AlertType:   "type2",
			Value:       2.34,
			Severity:    "minor",
			Status:      "resolved",
			Timestamp:   base.Add(time.Second),
			UpdatedAt:   base.Add(time.Second),
			Labels:      map[string]string{"k": "v2"},
		},
	}
	require.NoError(t, h.Send(vals))

	require.NoError(t, orm.StatisticsHandler(orm.DriverDefault, &cfg.Statistics.Database, func(db *sqlx.DB) error {
		var count int
		if err := db.Get(&count, "SELECT count() FROM history_notification"); err != nil {
			return err
		}
		assert.Equal(t, len(vals), count)
		return nil
	}))
}
