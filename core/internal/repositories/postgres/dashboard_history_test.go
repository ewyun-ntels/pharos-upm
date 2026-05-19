package postgres

import (
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/internal/repositories/test"
)

var _ test.DashboardHistoryTestable = (*TestPostgresDashboardHistory)(nil)

type TestPostgresDashboardHistory struct {
	dashboardHistoryRepository
	test.PostgresTestable
}

func (repo *TestPostgresDashboardHistory) SetUpDb() error {
	if err := repo.PostgresTestable.SetUpDb(); err != nil {
		return err
	}
	repo.dashboardHistoryRepository = *newDashboardHistoryRepository(repo.Config.Database)
	return nil
}
func (repo *TestPostgresDashboardHistory) TearDownDb() error {
	return repo.PostgresTestable.TearDownDb()
}
func (repo *TestPostgresDashboardHistory) ClearTable() error {
	if repo.Config == nil {
		return fmt.Errorf("test db not initialized")
	}
	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.Exec("DELETE FROM dashboard_history")
		return err
	})
}
func (repo *TestPostgresDashboardHistory) GetCount() (int, error) {
	if repo.Config == nil {
		return 0, fmt.Errorf("test db not initialized")
	}
	var count int
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.Get(&count, "SELECT COUNT(*) FROM dashboard_history")
	})
	return count, err
}

func TestDashboardHistoryRepository_Postgres(t *testing.T) {
	testable := &TestPostgresDashboardHistory{
		PostgresTestable: test.PostgresTestable{
			DefaultDatabase: "postgres",
			TestDatabase:    "pharos_test_dashboard_history_repo",
		},
	}
	if err := testable.Connect(t); err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}
	test.DashboardHistorySuite(t, testable)
}
