package sqlite

import (
	"database/sql"
	"testing"

	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/test"
)

// TestDataMigrationRepository_Suite runs the common test suite for SQLite
func TestDataMigrationRepository_Suite(t *testing.T) {
	testable := &SQLiteDataMigrationTestable{}

	defer func() {
		_ = testable.TearDownDb()
	}()

	test.RunDataMigrationTestSuite(t, testable)
}

// SQLiteDataMigrationTestable implements test.DataMigrationTestable for SQLite
type SQLiteDataMigrationTestable struct {
	st   *test.SqliteTestable
	repo repositories.DataMigrationRepository
	db   *sql.DB
}

func (s *SQLiteDataMigrationTestable) GetRepository() repositories.DataMigrationRepository {
	return s.repo
}

func (s *SQLiteDataMigrationTestable) GetDB() *sql.DB {
	return s.db
}

func (s *SQLiteDataMigrationTestable) SetUpDb() error {
	// Clean up previous test DB if exists
	if s.st != nil {
		_ = s.TearDownDb()
	}

	// Create new SQLite testable
	s.st = &test.SqliteTestable{}

	// Set up fresh DB
	err := s.st.SetUpDb()
	if err != nil {
		return err
	}

	// Get DB connection
	s.db = s.st.TestDb.DB

	// Create repository
	s.repo = newDataMigrationRepository(s.st.DbConfig)

	return nil
}

func (s *SQLiteDataMigrationTestable) TearDownDb() error {
	s.db = nil
	s.repo = nil

	if s.st != nil {
		err := s.st.TearDownDb()
		s.st = nil
		return err
	}

	return nil
}
