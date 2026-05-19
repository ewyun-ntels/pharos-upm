package statistics

import (
	"fmt"
	"testing"

	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/test"
)

var _ test.LoginHistoryTestable = (*TestPostgresStatisticsLoginHistory)(nil)

type TestPostgresStatisticsLoginHistory struct {
	loginHistoryRepository loginHistoryRepository
	test.PostgresStatisticsTestable
}

func (repo *TestPostgresStatisticsLoginHistory) SetUpDb() error {
	if err := repo.PostgresStatisticsTestable.SetUpDb(); err != nil {
		return err
	}
	repo.loginHistoryRepository = loginHistoryRepository{config: repo.DbConfig}
	return nil
}

func (repo *TestPostgresStatisticsLoginHistory) TearDownDb() error {
	return repo.PostgresStatisticsTestable.TearDownDb()
}

func (repo *TestPostgresStatisticsLoginHistory) GetLoginHistoryRepository() repositories.LoginHistoryRepository {
	return &repo.loginHistoryRepository
}

func (repo *TestPostgresStatisticsLoginHistory) InsertLoginHistory(record repositories.LoginHistoryRecord) error {
	if repo.TestDb == nil {
		return fmt.Errorf("test db not initialized")
	}
	var clientIp any
	if record.ClientIp != "" {
		clientIp = record.ClientIp
	}
	var userAgent any
	if record.UserAgent != "" {
		userAgent = record.UserAgent
	}
	var endReason any
	if record.EndReason != "" {
		endReason = record.EndReason
	}
	_, err := repo.TestDb.Exec(
		"INSERT INTO history_login (session_id, user_id, client_ip, user_agent, started_at, ended_at, end_reason) VALUES ($1, $2, $3, $4, $5, $6, $7)",
		record.SessionId, record.UserId, clientIp, userAgent, record.StartedAt, record.EndedAt, endReason,
	)
	return err
}

func (repo *TestPostgresStatisticsLoginHistory) ClearTable() error {
	if repo.TestDb == nil {
		return fmt.Errorf("test db not initialized")
	}
	_, err := repo.TestDb.Exec("DELETE FROM history_login WHERE true")
	return err
}

func TestLoginHistoryRepository_StatisticsPostgres(t *testing.T) {
	testable := &TestPostgresStatisticsLoginHistory{
		PostgresStatisticsTestable: test.PostgresStatisticsTestable{
			DefaultDatabase: "postgres",
			TestDatabase:    "pharos_test_stats_login_history",
		},
	}

	testable.Connect(t)

	if err := testable.SetUpDb(); err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("Failed to tear down test database: %v", err)
		}
	})

	// LoginHistorySuite는 SetUpDb/TearDownDb를 직접 호출하므로
	// 래퍼를 통해 DB를 재설정하지 않고 테이블만 초기화하도록 한다.
	wrapper := &testStatisticsLoginHistoryWrapper{TestPostgresStatisticsLoginHistory: testable}
	test.LoginHistorySuite(t, wrapper)
}

// testStatisticsLoginHistoryWrapper는 DB를 재설정하지 않고 테이블만 초기화한다.
type testStatisticsLoginHistoryWrapper struct {
	*TestPostgresStatisticsLoginHistory
}

func (w *testStatisticsLoginHistoryWrapper) SetUpDb() error {
	return w.ClearTable()
}

func (w *testStatisticsLoginHistoryWrapper) TearDownDb() error {
	return w.ClearTable()
}
