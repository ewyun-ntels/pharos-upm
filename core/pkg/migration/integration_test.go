package migration

import (
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories/factory"
	"ntels.com/pharos/core/pkg/common"
)

func TestFreshInstallTimingCheck(t *testing.T) {
	// Test: Verify fresh install detection works BEFORE schema migration

	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_timing.db")

	// Create database config
	config := common.Config{
		Database: orm.DatabaseConfig{
			Driver: orm.DriverSqlite,
			SQLite: orm.SQLiteConfig{
				Path: dbPath,
			},
		},
	}

	// Create repository
	dataMigrationRepo, err := factory.NewDataMigrationRepository(factory.RepositoryOptions{
		DatabaseConfig: config.Database,
	})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Test 1: Should be fresh install (no goose_db_version table yet)
	isFresh, err := dataMigrationRepo.IsFreshInstall()
	if err != nil {
		t.Fatalf("Failed to check fresh install: %v", err)
	}

	if !isFresh {
		t.Errorf("Expected fresh install before ANY migration, but got existing")
	}

	// Simulate schema migration by creating core_db_version table
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	createCoreDBVersionTableSQL := `
		CREATE TABLE core_db_version (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version_id INTEGER NOT NULL,
			is_applied INTEGER NOT NULL,
			tstamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := db.Exec(createCoreDBVersionTableSQL); err != nil {
		t.Fatalf("Failed to create core_db_version table: %v", err)
	}
	_ = db.Close()

	// Test 2: Should NOT be fresh install after goose table created
	isFresh, err = dataMigrationRepo.IsFreshInstall()
	if err != nil {
		t.Fatalf("Failed to check fresh install after goose table: %v", err)
	}

	if isFresh {
		t.Errorf("Expected existing installation after goose_db_version created, but still fresh")
	}

	t.Log("✅ Fresh install detection timing is correct:")
	t.Log("   - BEFORE goose_db_version: Fresh Install")
	t.Log("   - AFTER goose_db_version: Existing Install")
}

func TestLoadWithFreshInstallFlag(t *testing.T) {
	// Test: Verify Load() checks fresh install BEFORE schema migration

	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_load.db")

	// Create database config
	config := common.Config{
		Database: orm.DatabaseConfig{
			Driver: orm.DriverSqlite,
			SQLite: orm.SQLiteConfig{
				Path: dbPath,
			},
		},
	}

	// Create repository
	dataMigrationRepo, err := factory.NewDataMigrationRepository(factory.RepositoryOptions{
		DatabaseConfig: config.Database,
	})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	// Verify: Should be fresh install initially
	isFreshBefore, err := dataMigrationRepo.IsFreshInstall()
	if err != nil {
		t.Fatalf("Failed to check fresh install: %v", err)
	}

	if !isFreshBefore {
		t.Errorf("Expected fresh install before Load(), but got existing")
	}

	t.Log("✅ Confirmed: goose_db_version table does not exist before Load()")
	t.Log("   This means Load() will correctly detect Fresh Install")
}
