package sqlite

import (
	"fmt"
	"testing"

	"ntels.com/pharos/core/internal/repositories/test"
)

var _ test.DashboardHistoryTestable = (*TestSqliteDashboardHistory)(nil)

type TestSqliteDashboardHistory struct {
	dashboardHistoryRepository
	test.SqliteTestable
}

func (repo *TestSqliteDashboardHistory) SetUpDb() error {
	if err := repo.SqliteTestable.SetUpDb(); err != nil {
		return err
	}
	repo.dashboardHistoryRepository = *newDashboardHistoryRepository(repo.DbConfig)
	return nil
}
func (repo *TestSqliteDashboardHistory) TearDownDb() error {
	return repo.SqliteTestable.TearDownDb()
}
func (repo *TestSqliteDashboardHistory) ClearTable() error {
	if repo.TestDb == nil {
		return fmt.Errorf("test db not initialized")
	}
	_, err := repo.TestDb.Exec("DELETE FROM dashboard_history")
	return err
}
func (repo *TestSqliteDashboardHistory) GetCount() (int, error) {
	if repo.TestDb == nil {
		return 0, fmt.Errorf("test db not initialized")
	}
	var count int
	err := repo.TestDb.Get(&count, "SELECT COUNT(*) FROM dashboard_history")
	return count, err
}

func TestDashboardHistoryRepository_Sqlite(t *testing.T) {
	test.DashboardHistorySuite(t, &TestSqliteDashboardHistory{})
}
