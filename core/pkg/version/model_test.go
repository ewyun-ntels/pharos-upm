package version

import (
	"database/sql"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

// helper to init a sqlite db and create the version table
func setupSQLiteVersionDB(t *testing.T) common.Config {
	t.Helper()
	path := t.TempDir() + "/version_test.db"
	cfg := common.Config{}
	cfg.Database.Driver = orm.DriverSqlite
	cfg.Database.SQLite.Path = path

	// create table core_app_version
	err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS core_app_version (
		  module TEXT PRIMARY KEY,
		  version TEXT NOT NULL,
		  "commit" TEXT,
		  build_time TEXT,
		  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`)
		return err
	})
	if err != nil {
		t.Fatalf("failed to create table: %v", err)
	}
	return cfg
}

func getRow(t *testing.T, cfg common.Config) (module, versionStr, commitStr, buildTime string, updatedAt time.Time, ok bool) {
	t.Helper()
	err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		row := db.QueryRowx(`SELECT module, version, "commit", build_time, updated_at FROM core_app_version WHERE module = 'core';`)
		var u sql.NullTime
		if err := row.Scan(&module, &versionStr, &commitStr, &buildTime, &u); err != nil {
			return err
		}
		if u.Valid {
			updatedAt = u.Time
		}
		return nil
	})
	if err != nil {
		return "", "", "", "", time.Time{}, false
	}
	return module, versionStr, commitStr, buildTime, updatedAt, true
}

func TestUpsertVersion_InsertsAndUpdatesCoreModule(t *testing.T) {
	cfg := setupSQLiteVersionDB(t)

	// initial values
	Version = "v1"
	Commit = "c1"
	BuildTime = "b1"
	mv := Model{Config: &cfg}
	if err := mv.UpsertVersion(); err != nil {
		t.Fatalf("UpsertVersion failed: %v", err)
	}

	m, v, c, b, up1, ok := getRow(t, cfg)
	if !ok {
		t.Fatalf("failed to read back row after insert")
	}
	if m != "core" {
		t.Fatalf("module mismatch: want core got %s", m)
	}
	if v != "v1" || c != "c1" || b != "b1" {
		t.Fatalf("unexpected values after insert: v=%s c=%s b=%s", v, c, b)
	}

	// update values
	time.Sleep(1200 * time.Millisecond)
	Version = "v2"
	Commit = "c2"
	BuildTime = "b2"
	if err := mv.UpsertVersion(); err != nil {
		t.Fatalf("UpsertVersion failed on update: %v", err)
	}

	m, v, c, b, up2, ok := getRow(t, cfg)
	if !ok {
		t.Fatalf("failed to read back row after update")
	}
	if m != "core" {
		t.Fatalf("module mismatch after update: want core got %s", m)
	}
	if v != "v2" || c != "c2" || b != "b2" {
		t.Fatalf("unexpected values after update: v=%s c=%s b=%s", v, c, b)
	}
	if !up1.IsZero() && !up2.IsZero() && !up2.After(up1) {
		t.Fatalf("expected updated_at to advance: before=%v after=%v", up1, up2)
	}
}

// Test the Model.GetCoreVersionRow by inserting a row and fetching it.
func TestModel_GetCoreVersionRow(t *testing.T) {
	path := t.TempDir() + "/version_model_get.db"
	cfg := common.Config{}
	cfg.Database.Driver = orm.DriverSqlite
	cfg.Database.SQLite.Path = path

	// create table and insert a row
	err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS core_app_version (
		  module TEXT PRIMARY KEY,
		  version TEXT NOT NULL,
		  "commit" TEXT,
		  build_time TEXT,
		  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`); err != nil {
			return err
		}
		_, err := db.Exec(`INSERT INTO core_app_version(module, version, "commit", build_time) VALUES ('core','vX','cX','2025-08-26T15:15:40Z');`)
		return err
	})
	if err != nil {
		t.Fatalf("setup failed: %v", err)
	}

	m := Model{Config: &cfg}
	row, found, err := m.GetCoreVersionRow()
	if err != nil {
		t.Fatalf("GetCoreVersionRow error: %v", err)
	}
	if !found {
		t.Fatalf("expected to find row")
	}
	if row.Module != "core" || row.Version != "vX" {
		t.Fatalf("unexpected row: %+v", row)
	}
	if !row.Commit.Valid || row.Commit.String != "cX" {
		t.Fatalf("unexpected commit: %+v", row.Commit)
	}
	if !row.BuildTime.Valid || row.BuildTime.String == "" {
		t.Fatalf("expected build_time to be set")
	}
	// UpdatedAt should be valid
	if !row.UpdatedAt.Valid || row.UpdatedAt.Time.IsZero() {
		t.Fatalf("expected updated_at to be valid")
	}
}

// Ensure not-found case returns found=false and no error.
func TestModel_GetCoreVersionRow_NotFound(t *testing.T) {
	path := t.TempDir() + "/version_model_get_nf.db"
	cfg := common.Config{}
	cfg.Database.Driver = orm.DriverSqlite
	cfg.Database.SQLite.Path = path

	_ = orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS core_app_version (
		  module TEXT PRIMARY KEY,
		  version TEXT NOT NULL,
		  "commit" TEXT,
		  build_time TEXT,
		  updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
		);`)
		return err
	})

	m := Model{Config: &cfg}
	row, found, err := m.GetCoreVersionRow()
	if err != nil {
		t.Fatalf("GetCoreVersionRow error: %v", err)
	}
	if found {
		t.Fatalf("expected not found, got: %+v", row)
	}
	_ = sql.NullString{}
}
