package postgres

import (
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/test"
)

var _ test.DashboardTestable = (*TestPostgresDashboard)(nil)

type TestPostgresDashboard struct {
	dashboardRepository
	test.PostgresTestable
}

func (repo *TestPostgresDashboard) SetUpDb() error {
	if err := repo.PostgresTestable.SetUpDb(); err != nil {
		return err
	}
	repo.dashboardRepository = *newDashboardRepository(repo.Config.Database)
	return nil
}
func (repo *TestPostgresDashboard) TearDownDb() error {
	return repo.PostgresTestable.TearDownDb()
}
func (repo *TestPostgresDashboard) ClearTable() error {
	if repo.Config == nil {
		return fmt.Errorf("test db not initialized")
	}
	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.Exec("DELETE FROM dashboard WHERE true")
		return err
	})
}
func (repo *TestPostgresDashboard) GetDashboard(id string) (*repositories.DashboardEntity, error) {
	if repo.Config == nil {
		return nil, fmt.Errorf("test db not initialized")
	}
	var d postgresDashboard
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.Get(&d, "SELECT * FROM dashboard WHERE id = $1", id)
	})
	if err != nil {
		return nil, err
	}
	return d.toDashboardEntity()
}
func (repo *TestPostgresDashboard) GetDashboardCount() (int, error) {
	if repo.Config == nil {
		return 0, fmt.Errorf("test db not initialized")
	}
	var count int
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.Get(&count, "SELECT COUNT(*) FROM dashboard")
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}

func TestDashboardRepository_Postgres(t *testing.T) {
	testable := &TestPostgresDashboard{
		PostgresTestable: test.PostgresTestable{
			DefaultDatabase: "postgres",
			TestDatabase:    "pharos_test_dashboard_repo",
		},
	}

	if err := testable.Connect(t); err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	// Common Dashboard repository test
	test.DashboardSuite(t, testable)
}
