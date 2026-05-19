package test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/internal/repositories"
)

// DataMigrationTestable defines common test interface for DataMigrationRepository
type DataMigrationTestable interface {
	GetRepository() repositories.DataMigrationRepository
	GetDB() *sql.DB
	SetUpDb() error
	TearDownDb() error
}

// RunDataMigrationTestSuite runs a comprehensive test suite for DataMigrationRepository
func RunDataMigrationTestSuite(t *testing.T, testable DataMigrationTestable) {
	tests := []struct {
		name string
		fn   func(t *testing.T, testable DataMigrationTestable)
	}{
		{"CreateTable", testCreateTable},
		{"RecordAndGetMigrations", testRecordAndGetMigrations},
		{"BooleanComparison", testBooleanComparison},
		{"ConcurrentRecording", testConcurrentRecording},
		{"Idempotency", testIdempotency},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set up clean DB for each test
			err := testable.SetUpDb()
			if err != nil {
				t.Fatalf("Failed to set up DB: %v", err)
			}

			// Run test
			tt.fn(t, testable)

			// Clean up (optional, depends on implementation)
			// Some implementations may want to keep DB for debugging
		})
	}
}

func testCreateTable(t *testing.T, testable DataMigrationTestable) {
	repo := testable.GetRepository()
	db := testable.GetDB()

	// First creation should succeed
	err := repo.CreateMigrationTable()
	assert.NoError(t, err, "First table creation should succeed")

	// Verify table exists (works for both PostgreSQL and SQLite)
	var tableExists bool
	// Try PostgreSQL style first
	err = db.QueryRow(`
		SELECT EXISTS (
			SELECT 1 FROM information_schema.tables 
			WHERE table_name = 'core_data_version'
		)
	`).Scan(&tableExists)

	if err != nil {
		// Fall back to SQLite style
		var tableName string
		err = db.QueryRow(`
			SELECT name FROM sqlite_master 
			WHERE type='table' AND name='core_data_version'
		`).Scan(&tableName)
		tableExists = (err == nil && tableName == "core_data_version")
	}

	assert.True(t, tableExists, "Table should exist after creation")

	// Second creation should be idempotent (no error)
	err = repo.CreateMigrationTable()
	assert.NoError(t, err, "Creating existing table should succeed (idempotent)")
}

func testRecordAndGetMigrations(t *testing.T, testable DataMigrationTestable) {
	repo := testable.GetRepository()
	db := testable.GetDB()

	// Create table
	err := repo.CreateMigrationTable()
	assert.NoError(t, err)

	// Initially, no migrations applied
	applied, err := repo.GetAppliedMigrations()
	assert.NoError(t, err)
	assert.Empty(t, applied, "Initially no migrations should be applied")

	// Record some migrations
	tx, err := db.Begin()
	assert.NoError(t, err)

	versions := []int64{1, 2, 5, 10, 20250118000000}
	for _, version := range versions {
		err = repo.RecordMigration(tx, version)
		assert.NoError(t, err, "Recording migration %d should succeed", version)
	}

	err = tx.Commit()
	assert.NoError(t, err, "Transaction commit should succeed")

	// Verify migrations were recorded
	applied, err = repo.GetAppliedMigrations()
	assert.NoError(t, err)
	assert.Len(t, applied, len(versions), "All migrations should be recorded")

	for _, version := range versions {
		assert.True(t, applied[version], "Version %d should be applied", version)
	}
}

func testBooleanComparison(t *testing.T, testable DataMigrationTestable) {
	repo := testable.GetRepository()
	db := testable.GetDB()

	// Create table
	err := repo.CreateMigrationTable()
	assert.NoError(t, err)

	// Record migration
	tx, err := db.Begin()
	assert.NoError(t, err)

	err = repo.RecordMigration(tx, 100)
	assert.NoError(t, err)

	err = tx.Commit()
	assert.NoError(t, err)

	// Verify GetAppliedMigrations uses correct BOOLEAN comparison
	applied, err := repo.GetAppliedMigrations()
	assert.NoError(t, err)
	assert.True(t, applied[100], "Version 100 should be applied")
}

func testConcurrentRecording(t *testing.T, testable DataMigrationTestable) {
	repo := testable.GetRepository()
	db := testable.GetDB()

	// Create table
	err := repo.CreateMigrationTable()
	assert.NoError(t, err)

	// Record migration
	tx1, err := db.Begin()
	assert.NoError(t, err)

	err = repo.RecordMigration(tx1, 300)
	assert.NoError(t, err)

	err = tx1.Commit()
	assert.NoError(t, err)

	// Try to record same version again (should fail due to PRIMARY KEY constraint)
	tx2, err := db.Begin()
	assert.NoError(t, err)

	err = repo.RecordMigration(tx2, 300)
	assert.Error(t, err, "Recording duplicate version should fail")

	_ = tx2.Rollback()
}

func testIdempotency(t *testing.T, testable DataMigrationTestable) {
	repo := testable.GetRepository()

	// Multiple CreateTable calls should not error
	for i := range 3 {
		err := repo.CreateMigrationTable()
		assert.NoError(t, err, "CreateMigrationTable should be idempotent (iteration %d)", i)
	}
}
