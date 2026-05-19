package data

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories/factory"
	"ntels.com/pharos/core/pkg/common"
)

func TestFreshInstallDetection(t *testing.T) {
	t.Run("Fresh install without goose_db_version", func(t *testing.T) {
		// Create temp directory for test database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test_fresh.db")

		// Create database config
		config := common.Config{
			Database: orm.DatabaseConfig{
				Driver: orm.DriverSqlite,
				SQLite: orm.SQLiteConfig{
					Path: dbPath,
				},
			},
		}

		// Create data migration repository
		repo, err := factory.NewDataMigrationRepository(factory.RepositoryOptions{
			DatabaseConfig: config.Database,
		})
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		// Create migration table
		if err := repo.CreateMigrationTable(); err != nil {
			t.Fatalf("Failed to create migration table: %v", err)
		}

		// Test: Without goose_db_version table, should be fresh install
		isFresh, err := repo.IsFreshInstall()
		if err != nil {
			t.Fatalf("Failed to check fresh install: %v", err)
		}

		if !isFresh {
			t.Errorf("Expected fresh install (no goose_db_version), but got existing installation")
		}
	})

	t.Run("Existing install with core_db_version table", func(t *testing.T) {
		// Create temp directory for test database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test_existing.db")

		// Create database config
		config := common.Config{
			Database: orm.DatabaseConfig{
				Driver: orm.DriverSqlite,
				SQLite: orm.SQLiteConfig{
					Path: dbPath,
				},
			},
		}

		// Open DB and create core_db_version table (simulate existing schema)
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			t.Fatalf("Failed to open database: %v", err)
		}
		defer func() { _ = db.Close() }()

		// Create core_db_version table
		_, err = db.Exec(`
			CREATE TABLE core_db_version (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				version_id BIGINT NOT NULL UNIQUE,
				is_applied BOOLEAN NOT NULL DEFAULT 1,
				tstamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`)
		if err != nil {
			t.Fatalf("Failed to create core_db_version table: %v", err)
		}

		// Create data migration repository
		repo, err := factory.NewDataMigrationRepository(factory.RepositoryOptions{
			DatabaseConfig: config.Database,
		})
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		// Test: With core_db_version table, should NOT be fresh install
		isFresh, err := repo.IsFreshInstall()
		if err != nil {
			t.Fatalf("Failed to check fresh install: %v", err)
		}

		if isFresh {
			t.Errorf("Expected existing installation (has core_db_version table), but got fresh install")
		}
	})

	t.Run("MarkAllAsApplied functionality", func(t *testing.T) {
		// Create temp directory for test database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test_mark.db")

		// Create database config
		config := common.Config{
			Database: orm.DatabaseConfig{
				Driver: orm.DriverSqlite,
				SQLite: orm.SQLiteConfig{
					Path: dbPath,
				},
			},
		}

		// Create data migration repository
		repo, err := factory.NewDataMigrationRepository(factory.RepositoryOptions{
			DatabaseConfig: config.Database,
		})
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		// Create migration table
		if err := repo.CreateMigrationTable(); err != nil {
			t.Fatalf("Failed to create migration table: %v", err)
		}

		// Mark migrations as applied
		versions := []int64{20250118000000, 20250119000000}
		if err := repo.MarkAllAsApplied(versions); err != nil {
			t.Fatalf("Failed to mark migrations as applied: %v", err)
		}

		// Verify marked versions are recorded
		applied, err := repo.GetAppliedMigrations()
		if err != nil {
			t.Fatalf("Failed to get applied migrations: %v", err)
		}

		for _, v := range versions {
			if !applied[v] {
				t.Errorf("Version %d was not marked as applied", v)
			}
		}
	})

	t.Run("MarkAllAsApplied with duplicate versions (ON CONFLICT)", func(t *testing.T) {
		// Create temp directory for test database
		tmpDir := t.TempDir()
		dbPath := filepath.Join(tmpDir, "test_duplicate.db")

		// Create database config
		config := common.Config{
			Database: orm.DatabaseConfig{
				Driver: orm.DriverSqlite,
				SQLite: orm.SQLiteConfig{
					Path: dbPath,
				},
			},
		}

		// Create data migration repository
		repo, err := factory.NewDataMigrationRepository(factory.RepositoryOptions{
			DatabaseConfig: config.Database,
		})
		if err != nil {
			t.Fatalf("Failed to create repository: %v", err)
		}

		// Create migration table
		if err := repo.CreateMigrationTable(); err != nil {
			t.Fatalf("Failed to create migration table: %v", err)
		}

		// First insert
		versions := []int64{20250118000000, 20250119000000}
		if err := repo.MarkAllAsApplied(versions); err != nil {
			t.Fatalf("Failed to mark migrations as applied (first time): %v", err)
		}

		// Second insert with same versions (should not error due to ON CONFLICT)
		if err := repo.MarkAllAsApplied(versions); err != nil {
			t.Fatalf("Failed to mark migrations as applied (duplicate insert): %v", err)
		}

		// Verify versions are still recorded (only once)
		applied, err := repo.GetAppliedMigrations()
		if err != nil {
			t.Fatalf("Failed to get applied migrations: %v", err)
		}

		for _, v := range versions {
			if !applied[v] {
				t.Errorf("Version %d was not marked as applied", v)
			}
		}

		// Verify no duplicate records by counting
		db, err := sql.Open("sqlite", dbPath)
		if err != nil {
			t.Fatalf("Failed to open database: %v", err)
		}
		defer func() { _ = db.Close() }()

		var count int
		countQuery := "SELECT COUNT(*) FROM core_data_version WHERE version_id = ?"
		if err := db.QueryRow(countQuery, versions[0]).Scan(&count); err != nil {
			t.Fatalf("Failed to count records: %v", err)
		}

		if count != 1 {
			t.Errorf("Expected 1 record for version %d, got %d", versions[0], count)
		}
	})
}

func TestRunWithFreshInstall(t *testing.T) {
	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_run.db")

	// Open database to create users table (simulate schema migration)
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create users table
	createTableSQL := `
		CREATE TABLE users (
			username TEXT PRIMARY KEY,
			extra TEXT
		)
	`
	if _, err := db.Exec(createTableSQL); err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// NOTE: Do NOT create goose_db_version table
	// This simulates fresh install where schema migrations haven't run yet

	// Close DB to allow Run() to open it
	_ = db.Close()

	// Register test migration
	testMigrationExecuted := false
	testMigration := &Migration{
		Version: 20250118000000,
		Name:    "test_migration",
		Up: func(ctx context.Context, db *sql.DB) error {
			testMigrationExecuted = true
			return nil
		},
		Down: func(ctx context.Context, db *sql.DB) error {
			return nil
		},
	}

	// Temporarily add test migration
	originalMigrations := migrations
	migrations = []*Migration{testMigration}
	defer func() { migrations = originalMigrations }()

	// Create config
	config := common.Config{
		Database: orm.DatabaseConfig{
			Driver: orm.DriverSqlite,
			SQLite: orm.SQLiteConfig{
				Path: dbPath,
			},
		},
	}

	// Run migrations with fresh install flag
	if err := RunWithFreshInstallFlag(config, true); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Test: Migration should NOT have executed (fresh install)
	if testMigrationExecuted {
		t.Errorf("Migration was executed on fresh install, but should have been skipped")
	}

	// Verify version was recorded
	repo, err := factory.NewDataMigrationRepository(factory.RepositoryOptions{
		DatabaseConfig: config.Database,
	})
	if err != nil {
		t.Fatalf("Failed to create repository: %v", err)
	}

	applied, err := repo.GetAppliedMigrations()
	if err != nil {
		t.Fatalf("Failed to get applied migrations: %v", err)
	}

	if !applied[20250118000000] {
		t.Errorf("Test migration version was not recorded")
	}
}

func TestRunWithExistingInstall(t *testing.T) {
	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test_existing.db")

	// Open database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}

	// Create users table with OLD format data
	createTableSQL := `
		CREATE TABLE users (
			username TEXT PRIMARY KEY,
			extra TEXT
		)
	`
	if _, err := db.Exec(createTableSQL); err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	// Insert user with OLD attributes format
	insertSQL := `INSERT INTO users (username, extra) VALUES (?, ?)`
	oldAttrs := `{"role:super_admin": true, "userinfo": {"email": "test@example.com"}}`
	if _, err := db.Exec(insertSQL, "testuser", oldAttrs); err != nil {
		t.Fatalf("Failed to insert test user: %v", err)
	}

	// Create goose_db_version table (simulate existing installation)
	createGooseTableSQL := `
		CREATE TABLE goose_db_version (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			version_id INTEGER NOT NULL,
			is_applied INTEGER NOT NULL,
			tstamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := db.Exec(createGooseTableSQL); err != nil {
		t.Fatalf("Failed to create goose_db_version table: %v", err)
	}

	// Insert a version record (simulate that schema migrations have run)
	insertGooseSQL := `INSERT INTO goose_db_version (version_id, is_applied) VALUES (1, 1)`
	if _, err := db.Exec(insertGooseSQL); err != nil {
		t.Fatalf("Failed to insert goose version: %v", err)
	}

	// Create data_migrations table and insert old record (simulate upgrade scenario)
	createMigrationTableSQL := `
		CREATE TABLE core_data_version (
			version_id BIGINT PRIMARY KEY,
			is_applied BOOLEAN NOT NULL DEFAULT 1,
			tstamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
		)
	`
	if _, err := db.Exec(createMigrationTableSQL); err != nil {
		t.Fatalf("Failed to create migration table: %v", err)
	}

	// Insert old migration record (simulate previous data migrations)
	insertMigrationSQL := `INSERT INTO core_data_version (version_id, is_applied) VALUES (20250117000000, 1)`
	if _, err := db.Exec(insertMigrationSQL); err != nil {
		t.Fatalf("Failed to insert migration record: %v", err)
	}

	// Close DB
	_ = db.Close()

	// Register test migration
	testMigrationExecuted := false
	testMigration := &Migration{
		Version: 20250118000000,
		Name:    "test_existing_migration",
		Up: func(ctx context.Context, db *sql.DB) error {
			testMigrationExecuted = true
			return nil
		},
		Down: func(ctx context.Context, db *sql.DB) error {
			return nil
		},
	}

	// Temporarily add test migration
	originalMigrations := migrations
	migrations = []*Migration{testMigration}
	defer func() { migrations = originalMigrations }()

	// Create config
	config := common.Config{
		Database: orm.DatabaseConfig{
			Driver: orm.DriverSqlite,
			SQLite: orm.SQLiteConfig{
				Path: dbPath,
			},
		},
	}

	// Run migrations with existing install flag
	if err := RunWithFreshInstallFlag(config, false); err != nil {
		t.Fatalf("Failed to run migrations: %v", err)
	}

	// Test: Migration SHOULD have executed (existing install)
	if !testMigrationExecuted {
		t.Errorf("Migration was not executed on existing install, but should have run")
	}
}
