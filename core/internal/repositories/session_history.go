package repositories

import (
	"context"
	"time"
)

type LoginHistoryRecord struct {
	SessionId string     `json:"session_id"`
	UserId    string     `json:"user_id"`
	ClientIp  string     `json:"client_ip"`
	UserAgent string     `json:"user_agent"`
	StartedAt time.Time  `json:"started_at"`
	EndedAt   *time.Time `json:"ended_at"`
	EndReason string     `json:"end_reason"`
}

type LoginHistoryRepository interface {
	Query(ctx context.Context, username string, limit, offset int) ([]LoginHistoryRecord, error)
	Count(ctx context.Context, username string) (int, error)
	UpsertSession(ctx context.Context, record LoginHistoryRecord) error
	CloseOrphanedSessions(ctx context.Context, endedAt time.Time) (int64, error)
}
