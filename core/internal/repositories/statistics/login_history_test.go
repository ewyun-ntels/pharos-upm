package statistics

import (
	"fmt"
	"testing"

	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/test"
)

var _ test.LoginHistoryTestable = (*TestStatisticsLoginHistory)(nil)

type TestStatisticsLoginHistory struct {
	loginHistoryRepository loginHistoryRepository
	test.StatisticsTestable
}

func (repo *TestStatisticsLoginHistory) SetUpDb() error {
	if err := repo.StatisticsTestable.SetUpDb(); err != nil {
		return err
	}
	repo.loginHistoryRepository = loginHistoryRepository{config: repo.DbConfig}
	return nil
}

func (repo *TestStatisticsLoginHistory) TearDownDb() error {
	return repo.StatisticsTestable.TearDownDb()
}

func (repo *TestStatisticsLoginHistory) GetLoginHistoryRepository() repositories.LoginHistoryRepository {
	return &repo.loginHistoryRepository
}

func (repo *TestStatisticsLoginHistory) InsertLoginHistory(record repositories.LoginHistoryRecord) error {
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
		"INSERT INTO history_login (session_id, user_id, client_ip, user_agent, started_at, ended_at, end_reason) VALUES (?, ?, ?, ?, ?, ?, ?)",
		record.SessionId, record.UserId, clientIp, userAgent, record.StartedAt, record.EndedAt, endReason,
	)
	return err
}

func (repo *TestStatisticsLoginHistory) ClearTable() error {
	if repo.TestDb == nil {
		return fmt.Errorf("test db not initialized")
	}
	_, err := repo.TestDb.Exec("DELETE FROM history_login WHERE 1=1")
	return err
}

func TestLoginHistoryRepository_Statistics(t *testing.T) {
	test.LoginHistorySuite(t, &TestStatisticsLoginHistory{})
}
