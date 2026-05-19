package sqlite

import (
	"fmt"
	"testing"

	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/test"
)

var _ test.DashboardFolderTestable = (*TestSqliteDashboardFolder)(nil)

type TestSqliteDashboardFolder struct {
	dashboardFolderRepository
	test.SqliteTestable
}

func (repo *TestSqliteDashboardFolder) SetUpDb() error {
	if err := repo.SqliteTestable.SetUpDb(); err != nil {
		return err
	}
	repo.dashboardFolderRepository = *newDashboardFolderRepository(repo.DbConfig)
	return nil
}
func (repo *TestSqliteDashboardFolder) TearDownDb() error {
	return repo.SqliteTestable.TearDownDb()
}
func (repo *TestSqliteDashboardFolder) ClearTable() error {
	if repo.TestDb == nil {
		return fmt.Errorf("test db not initialized")
	}
	if _, err := repo.TestDb.Exec("UPDATE dashboard SET folder_id = NULL WHERE folder_id IS NOT NULL"); err != nil {
		return err
	}
	_, err := repo.TestDb.Exec("DELETE FROM dashboard_folder WHERE true")
	return err
}
func (repo *TestSqliteDashboardFolder) GetFolder(id string) (*repositories.DashboardFolderEntity, error) {
	if repo.TestDb == nil {
		return nil, fmt.Errorf("test db not initialized")
	}
	var f sqliteDashboardFolder
	err := repo.TestDb.Get(&f, "SELECT * FROM dashboard_folder WHERE id = ?", id)
	if err != nil {
		return nil, err
	}
	return f.toEntity(), nil
}
func (repo *TestSqliteDashboardFolder) GetFolderCount() (int, error) {
	if repo.TestDb == nil {
		return 0, fmt.Errorf("test db not initialized")
	}
	var count int
	err := repo.TestDb.Get(&count, "SELECT COUNT(*) FROM dashboard_folder")
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (repo *TestSqliteDashboardFolder) CreateDashboard(id string) error {
	if repo.TestDb == nil {
		return fmt.Errorf("test db not initialized")
	}
	_, err := repo.TestDb.Exec(
		`INSERT INTO dashboard (id, config) VALUES (?, ?)`,
		id, `{"title":"test","type":"default"}`,
	)
	return err
}
func (repo *TestSqliteDashboardFolder) GetDashboardFolderID(dashboardID string) (*string, error) {
	if repo.TestDb == nil {
		return nil, fmt.Errorf("test db not initialized")
	}
	var folderID *string
	err := repo.TestDb.Get(&folderID, "SELECT folder_id FROM dashboard WHERE id = ?", dashboardID)
	if err != nil {
		return nil, err
	}
	return folderID, nil
}

func TestDashboardFolderRepository_Sqlite(t *testing.T) {
	test.DashboardFolderSuite(t, &TestSqliteDashboardFolder{})
}
