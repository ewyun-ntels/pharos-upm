package authhandler

import (
	"context"
	"log/slog"
	"time"

	"ntels.com/pharos/core/internal/repositories"
)

// SessionRecord holds the data for a session history UPSERT.
// EndedAt == nil means the session is still active.
// For login failures, EndedAt equals StartedAt (instant-closed session).
type SessionRecord struct {
	SessionId string
	UserId    string
	ClientIp  string
	UserAgent string
	StartedAt int64  // Unix timestamp
	EndedAt   *int64 // nil = active session
	EndReason string // "logout", "expired", "revoked", "failed: <msg>", or ""
}

func sendSession(ctx context.Context, record SessionRecord) {
	if loginHistoryRepo == nil || record.SessionId == "" {
		return
	}

	rec := repositories.LoginHistoryRecord{
		SessionId: record.SessionId,
		UserId:    record.UserId,
		ClientIp:  record.ClientIp,
		UserAgent: record.UserAgent,
		StartedAt: time.Unix(record.StartedAt, 0),
		EndReason: record.EndReason,
	}
	if record.EndedAt != nil {
		t := time.Unix(*record.EndedAt, 0)
		rec.EndedAt = &t
	}

	if err := loginHistoryRepo.UpsertSession(ctx, rec); err != nil {
		slog.Error("session history upsert failed", "error", err)
	}
}

// closeOrphanedSessions marks all active sessions (ended_at IS NULL) as ended with
// end_reason = "server_restart". Called on server startup to handle sessions that
// were lost when the server was killed (in-memory token store does not survive restarts).
func closeOrphanedSessions(repo repositories.LoginHistoryRepository, ctx context.Context) {
	n, err := repo.CloseOrphanedSessions(ctx, time.Now())
	if err != nil {
		slog.Error("closeOrphanedSessions failed", "error", err)
		return
	}
	if n > 0 {
		slog.Info("closed orphaned sessions on startup", "count", n, "reason", "server_restart")
	}
}
