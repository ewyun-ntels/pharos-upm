package sqlite

import (
	"fmt"
	"testing"

	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/test"
)

var _ test.DashboardTestable = (*TestSqliteDashboard)(nil)

type TestSqliteDashboard struct {
	dashboardRepository
	test.SqliteTestable
}

func (repo *TestSqliteDashboard) SetUpDb() error {
	if err := repo.SqliteTestable.SetUpDb(); err != nil {
		return err
	}
	repo.dashboardRepository = *newDashboardRepository(repo.DbConfig)
	return nil
}
func (repo *TestSqliteDashboard) TearDownDb() error {
	return repo.SqliteTestable.TearDownDb()
}
func (repo *TestSqliteDashboard) ClearTable() error {
	if repo.TestDb == nil {
		return fmt.Errorf("test db not initialized")
	}
	_, err := repo.TestDb.Exec("DELETE FROM dashboard WHERE true")
	return err
}
func (repo *TestSqliteDashboard) GetDashboard(id string) (*repositories.DashboardEntity, error) {
	if repo.TestDb == nil {
		return nil, fmt.Errorf("test db not initialized")
	}
	var d sqliteDashboard
	err := repo.TestDb.Get(&d, "SELECT * FROM dashboard WHERE id = ?", id)
	if err != nil {
		return nil, err
	}

	return d.toDashboardEntity()
}
func (repo *TestSqliteDashboard) GetDashboardCount() (int, error) {
	if repo.TestDb == nil {
		return 0, fmt.Errorf("test db not initialized")
	}
	var count int
	err := repo.TestDb.Get(&count, "SELECT COUNT(*) FROM dashboard")
	if err != nil {
		return 0, err
	}
	return count, nil
}

func TestDashboardRepository_Sqlite(t *testing.T) {
	test.DashboardSuite(t, &TestSqliteDashboard{})
}
