package postgres

import (
	"database/sql"
	"testing"

	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/test"
)

// TestDataMigrationRepository_Suite runs the common test suite for PostgreSQL
func TestDataMigrationRepository_Suite(t *testing.T) {
	testable := &PostgresDataMigrationTestable{}

	err := testable.Connect(t)
	if err != nil {
		t.Fatalf("Failed to connect: %v", err)
	}

	defer func() {
		_ = testable.TearDownDb()
	}()

	test.RunDataMigrationTestSuite(t, testable)
}

// PostgresDataMigrationTestable implements test.DataMigrationTestable for PostgreSQL
type PostgresDataMigrationTestable struct {
	pt   *test.PostgresTestable
	repo repositories.DataMigrationRepository
	db   *sql.DB
}

func (p *PostgresDataMigrationTestable) Connect(t *testing.T) error {
	p.pt = &test.PostgresTestable{
		DefaultDatabase: "postgres",
		TestDatabase:    "test_data_migration_suite",
	}

	return p.pt.Connect(t)
}

func (p *PostgresDataMigrationTestable) GetRepository() repositories.DataMigrationRepository {
	return p.repo
}

func (p *PostgresDataMigrationTestable) GetDB() *sql.DB {
	return p.db
}

func (p *PostgresDataMigrationTestable) SetUpDb() error {
	// Clean up previous test DB if exists
	if p.db != nil {
		_ = p.db.Close()
		_ = p.pt.TearDownDb()
	}

	// Set up fresh DB
	err := p.pt.SetUpDb()
	if err != nil {
		return err
	}

	// Open connection
	p.db, err = sql.Open("postgres", p.pt.Config.Database.GetDataSourceName())
	if err != nil {
		return err
	}

	// Create repository
	p.repo = newDataMigrationRepository(p.pt.Config.Database)

	return nil
}

func (p *PostgresDataMigrationTestable) TearDownDb() error {
	if p.db != nil {
		_ = p.db.Close()
		p.db = nil
	}

	if p.pt != nil {
		return p.pt.TearDownDb()
	}

	return nil
}
