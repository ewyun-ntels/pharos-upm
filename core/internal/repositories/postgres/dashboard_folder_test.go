package postgres

import (
	"fmt"
	"testing"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/test"
)

var _ test.DashboardFolderTestable = (*TestPostgresDashboardFolder)(nil)

type TestPostgresDashboardFolder struct {
	dashboardFolderRepository
	test.PostgresTestable
}

func (repo *TestPostgresDashboardFolder) SetUpDb() error {
	if err := repo.PostgresTestable.SetUpDb(); err != nil {
		return err
	}
	repo.dashboardFolderRepository = *newDashboardFolderRepository(repo.Config.Database)
	return nil
}
func (repo *TestPostgresDashboardFolder) TearDownDb() error {
	return repo.PostgresTestable.TearDownDb()
}
func (repo *TestPostgresDashboardFolder) ClearTable() error {
	if repo.Config == nil {
		return fmt.Errorf("test db not initialized")
	}
	return repo.Handler(func(db *sqlx.DB) error {
		if _, err := db.Exec("UPDATE dashboard SET folder_id = NULL WHERE folder_id IS NOT NULL"); err != nil {
			return err
		}
		_, err := db.Exec("DELETE FROM dashboard_folder WHERE true")
		return err
	})
}
func (repo *TestPostgresDashboardFolder) GetFolder(id string) (*repositories.DashboardFolderEntity, error) {
	if repo.Config == nil {
		return nil, fmt.Errorf("test db not initialized")
	}
	var f postgresDashboardFolder
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.Get(&f, "SELECT * FROM dashboard_folder WHERE id = $1", id)
	})
	if err != nil {
		return nil, err
	}
	return f.toEntity(), nil
}
func (repo *TestPostgresDashboardFolder) GetFolderCount() (int, error) {
	if repo.Config == nil {
		return 0, fmt.Errorf("test db not initialized")
	}
	var count int
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.Get(&count, "SELECT COUNT(*) FROM dashboard_folder")
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (repo *TestPostgresDashboardFolder) CreateDashboard(id string) error {
	if repo.Config == nil {
		return fmt.Errorf("test db not initialized")
	}
	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.Exec(
			`INSERT INTO dashboard (id, config) VALUES ($1, $2)`,
			id, `{"title":"test","type":"default"}`,
		)
		return err
	})
}
func (repo *TestPostgresDashboardFolder) GetDashboardFolderID(dashboardID string) (*string, error) {
	if repo.Config == nil {
		return nil, fmt.Errorf("test db not initialized")
	}
	var folderID *string
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.Get(&folderID, "SELECT folder_id FROM dashboard WHERE id = $1", dashboardID)
	})
	if err != nil {
		return nil, err
	}
	return folderID, nil
}

func TestDashboardFolderRepository_Postgres(t *testing.T) {
	testable := &TestPostgresDashboardFolder{
		PostgresTestable: test.PostgresTestable{
			DefaultDatabase: "postgres",
			TestDatabase:    "pharos_test_dashboard_folder_repo",
		},
	}

	if err := testable.Connect(t); err != nil {
		t.Fatalf("Failed to connect to test database: %v", err)
	}

	test.DashboardFolderSuite(t, testable)
}
