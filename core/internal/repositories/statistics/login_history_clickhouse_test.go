package statistics

import (
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/test"
)

var _ test.LoginHistoryTestable = (*TestClickHouseStatisticsLoginHistory)(nil)

type TestClickHouseStatisticsLoginHistory struct {
	loginHistoryRepository loginHistoryRepository
	test.ClickHouseStatisticsTestable
}

func (repo *TestClickHouseStatisticsLoginHistory) SetUpDb() error {
	if err := repo.ClickHouseStatisticsTestable.SetUpDb(); err != nil {
		return err
	}
	repo.loginHistoryRepository = loginHistoryRepository{config: repo.DbConfig}
	return nil
}

func (repo *TestClickHouseStatisticsLoginHistory) TearDownDb() error {
	return repo.ClickHouseStatisticsTestable.TearDownDb()
}

func (repo *TestClickHouseStatisticsLoginHistory) GetLoginHistoryRepository() repositories.LoginHistoryRepository {
	return &repo.loginHistoryRepository
}

func (repo *TestClickHouseStatisticsLoginHistory) InsertLoginHistory(record repositories.LoginHistoryRecord) error {
	if repo.DbConfig.Driver == "" {
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
	return orm.StatisticsHandler("", &repo.DbConfig, func(db *sqlx.DB) error {
		_, err := db.Exec(
			"INSERT INTO history_login (session_id, user_id, client_ip, user_agent, started_at, ended_at, end_reason, version) VALUES (?, ?, ?, ?, ?, ?, ?, ?)",
			record.SessionId, record.UserId, clientIp, userAgent, record.StartedAt, record.EndedAt, endReason, time.Now(),
		)
		return err
	})
}

func (repo *TestClickHouseStatisticsLoginHistory) ClearTable() error {
	if repo.DbConfig.Driver == "" {
		return fmt.Errorf("test db not initialized")
	}
	return orm.StatisticsHandler("", &repo.DbConfig, func(db *sqlx.DB) error {
		_, err := db.Exec("TRUNCATE TABLE history_login")
		return err
	})
}

func TestLoginHistoryRepository_StatisticsClickHouse(t *testing.T) {
	testable := &TestClickHouseStatisticsLoginHistory{}
	testable.Connect(t)

	if err := testable.SetUpDb(); err != nil {
		t.Fatalf("Failed to set up ClickHouse test database: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("Failed to tear down ClickHouse test database: %v", err)
		}
	})

	wrapper := &testClickHouseLoginHistoryWrapper{TestClickHouseStatisticsLoginHistory: testable}
	test.LoginHistorySuite(t, wrapper)
}

type testClickHouseLoginHistoryWrapper struct {
	*TestClickHouseStatisticsLoginHistory
}

func (w *testClickHouseLoginHistoryWrapper) SetUpDb() error {
	return w.ClearTable()
}

func (w *testClickHouseLoginHistoryWrapper) TearDownDb() error {
	return w.ClearTable()
}
