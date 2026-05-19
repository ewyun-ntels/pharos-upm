package test

import (
	"context"
	"testing"
	"time"

	"ntels.com/pharos/core/internal/repositories"
)

type LoginHistoryTestable interface {
	SetUpDb() error
	TearDownDb() error
	GetLoginHistoryRepository() repositories.LoginHistoryRepository
	InsertLoginHistory(record repositories.LoginHistoryRecord) error
}

func LoginHistorySuite(t *testing.T, repo LoginHistoryTestable) {
	t.Helper()

	if err := repo.SetUpDb(); err != nil {
		t.Fatalf("SetUpDb failed: %v", err)
	}
	defer func() {
		if err := repo.TearDownDb(); err != nil {
			t.Logf("TearDownDb error: %v", err)
		}
	}()

	r := repo.GetLoginHistoryRepository()
	ctx := context.Background()
	now := time.Now().UTC().Truncate(time.Second)

	t.Run("Query returns empty when no records", func(t *testing.T) {
		records, err := r.Query(ctx, "", 100, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 0 {
			t.Errorf("expected 0 records, got %d", len(records))
		}
	})

	t.Run("Query returns inserted records ordered by started_at DESC", func(t *testing.T) {
		entries := []repositories.LoginHistoryRecord{
			{SessionId: "sess-1", StartedAt: now.Add(-2 * time.Minute), UserId: "alice", ClientIp: "10.0.0.1", EndReason: "logout"},
			{SessionId: "sess-2", StartedAt: now.Add(-1 * time.Minute), UserId: "bob", ClientIp: "10.0.0.2", EndReason: "logout"},
			{SessionId: "sess-3", StartedAt: now, UserId: "alice", ClientIp: "10.0.0.1", EndReason: "expired"},
		}
		for _, e := range entries {
			if err := repo.InsertLoginHistory(e); err != nil {
				t.Fatalf("InsertLoginHistory failed: %v", err)
			}
		}
		defer clearLoginHistory(t, repo)

		records, err := r.Query(ctx, "", 100, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 3 {
			t.Fatalf("expected 3 records, got %d", len(records))
		}
		if records[0].UserId != "alice" || records[0].EndReason != "expired" {
			t.Errorf("first record should be the latest (alice expired), got user=%s reason=%s", records[0].UserId, records[0].EndReason)
		}
	})

	t.Run("Query filters by username", func(t *testing.T) {
		entries := []repositories.LoginHistoryRecord{
			{SessionId: "sess-4", StartedAt: now.Add(-2 * time.Minute), UserId: "alice", ClientIp: "10.0.0.1"},
			{SessionId: "sess-5", StartedAt: now.Add(-1 * time.Minute), UserId: "bob", ClientIp: "10.0.0.2"},
		}
		for _, e := range entries {
			if err := repo.InsertLoginHistory(e); err != nil {
				t.Fatalf("InsertLoginHistory failed: %v", err)
			}
		}
		defer clearLoginHistory(t, repo)

		records, err := r.Query(ctx, "alice", 100, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("expected 1 record for alice, got %d", len(records))
		}
		if records[0].UserId != "alice" {
			t.Errorf("expected alice, got %s", records[0].UserId)
		}
	})

	t.Run("Query respects limit", func(t *testing.T) {
		for i := range 5 {
			if err := repo.InsertLoginHistory(repositories.LoginHistoryRecord{
				SessionId: "sess-limit-" + string(rune('0'+i)),
				StartedAt: now.Add(time.Duration(i) * time.Second),
				UserId:    "charlie",
				ClientIp:  "10.0.0.3",
			}); err != nil {
				t.Fatalf("InsertLoginHistory failed: %v", err)
			}
		}
		defer clearLoginHistory(t, repo)

		records, err := r.Query(ctx, "charlie", 3, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 3 {
			t.Errorf("expected 3 records with limit=3, got %d", len(records))
		}
	})

	t.Run("Query respects offset", func(t *testing.T) {
		for i := range 4 {
			if err := repo.InsertLoginHistory(repositories.LoginHistoryRecord{
				SessionId: "sess-offset-" + string(rune('0'+i)),
				StartedAt: now.Add(time.Duration(i) * time.Second),
				UserId:    "dave",
				ClientIp:  "10.0.0.4",
			}); err != nil {
				t.Fatalf("InsertLoginHistory failed: %v", err)
			}
		}
		defer clearLoginHistory(t, repo)

		records, err := r.Query(ctx, "dave", 100, 2)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 2 {
			t.Errorf("expected 2 records with offset=2 (4 total), got %d", len(records))
		}
	})

	t.Run("Count returns total without filter", func(t *testing.T) {
		for i := range 3 {
			if err := repo.InsertLoginHistory(repositories.LoginHistoryRecord{
				SessionId: "sess-count-" + string(rune('0'+i)),
				StartedAt: now.Add(time.Duration(i) * time.Second),
				UserId:    "countuser",
				ClientIp:  "10.0.0.5",
			}); err != nil {
				t.Fatalf("InsertLoginHistory failed: %v", err)
			}
		}
		defer clearLoginHistory(t, repo)

		total, err := r.Count(ctx, "")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 3 {
			t.Errorf("expected count=3, got %d", total)
		}
	})

	t.Run("Count filters by username", func(t *testing.T) {
		entries := []repositories.LoginHistoryRecord{
			{SessionId: "sess-cf-1", StartedAt: now.Add(-2 * time.Minute), UserId: "userA", ClientIp: "10.0.0.1"},
			{SessionId: "sess-cf-2", StartedAt: now.Add(-1 * time.Minute), UserId: "userA", ClientIp: "10.0.0.1", EndReason: "logout"},
			{SessionId: "sess-cf-3", StartedAt: now, UserId: "userB", ClientIp: "10.0.0.2"},
		}
		for _, e := range entries {
			if err := repo.InsertLoginHistory(e); err != nil {
				t.Fatalf("InsertLoginHistory failed: %v", err)
			}
		}
		defer clearLoginHistory(t, repo)

		total, err := r.Count(ctx, "userA")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if total != 2 {
			t.Errorf("expected count=2 for userA, got %d", total)
		}
	})

	t.Run("Query handles NULL client_ip and end_reason as empty string", func(t *testing.T) {
		if err := repo.InsertLoginHistory(repositories.LoginHistoryRecord{
			SessionId: "sess-null-1",
			StartedAt: now,
			UserId:    "nulltest",
			ClientIp:  "",
			EndReason: "",
		}); err != nil {
			t.Fatalf("InsertLoginHistory failed: %v", err)
		}
		defer clearLoginHistory(t, repo)

		records, err := r.Query(ctx, "nulltest", 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(records))
		}
		if records[0].ClientIp != "" || records[0].EndReason != "" {
			t.Errorf("expected empty strings, got clientIp=%q endReason=%q", records[0].ClientIp, records[0].EndReason)
		}
	})

	t.Run("Query returns user_agent when set", func(t *testing.T) {
		if err := repo.InsertLoginHistory(repositories.LoginHistoryRecord{
			SessionId: "sess-ua-1",
			StartedAt: now,
			UserId:    "uatest",
			ClientIp:  "10.0.0.7",
			UserAgent: "Mozilla/5.0 (TestBrowser)",
		}); err != nil {
			t.Fatalf("InsertLoginHistory failed: %v", err)
		}
		defer clearLoginHistory(t, repo)

		records, err := r.Query(ctx, "uatest", 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(records))
		}
		if records[0].UserAgent != "Mozilla/5.0 (TestBrowser)" {
			t.Errorf("expected user_agent to be set, got %q", records[0].UserAgent)
		}
		if records[0].ClientIp != "10.0.0.7" {
			t.Errorf("expected client_ip=10.0.0.7, got %q", records[0].ClientIp)
		}
	})

	t.Run("Query returns empty user_agent when NULL", func(t *testing.T) {
		if err := repo.InsertLoginHistory(repositories.LoginHistoryRecord{
			SessionId: "sess-ua-null",
			StartedAt: now,
			UserId:    "uanulltest",
			UserAgent: "",
		}); err != nil {
			t.Fatalf("InsertLoginHistory failed: %v", err)
		}
		defer clearLoginHistory(t, repo)

		records, err := r.Query(ctx, "uanulltest", 10, 0)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("expected 1 record, got %d", len(records))
		}
		if records[0].UserAgent != "" {
			t.Errorf("expected empty user_agent, got %q", records[0].UserAgent)
		}
	})

	t.Run("UpsertSession inserts then updates on conflict", func(t *testing.T) {
		sid := "sess-upsert-1"
		defer clearLoginHistory(t, repo)

		// 1) 활성 세션 insert
		if err := r.UpsertSession(ctx, repositories.LoginHistoryRecord{
			SessionId: sid,
			UserId:    "upsertuser",
			ClientIp:  "10.0.0.9",
			StartedAt: now.Add(-2 * time.Minute),
		}); err != nil {
			t.Fatalf("UpsertSession (insert) failed: %v", err)
		}

		records, err := r.Query(ctx, "upsertuser", 10, 0)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(records) != 1 || records[0].EndedAt != nil {
			t.Fatalf("expected 1 active record, got %d", len(records))
		}

		// 2) 동일 session_id로 ended_at 채워 upsert → ended_at/end_reason 갱신
		endedAt := now
		if err := r.UpsertSession(ctx, repositories.LoginHistoryRecord{
			SessionId: sid,
			UserId:    "upsertuser",
			StartedAt: now.Add(-2 * time.Minute),
			EndedAt:   &endedAt,
			EndReason: "logout",
		}); err != nil {
			t.Fatalf("UpsertSession (update) failed: %v", err)
		}

		records, err = r.Query(ctx, "upsertuser", 10, 0)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(records) != 1 {
			t.Fatalf("expected 1 record after upsert, got %d", len(records))
		}
		if records[0].EndedAt == nil {
			t.Error("expected ended_at to be set after upsert")
		}
		if records[0].EndReason != "logout" {
			t.Errorf("expected end_reason=logout, got %q", records[0].EndReason)
		}
	})

	t.Run("CloseOrphanedSessions marks active sessions as server_restart", func(t *testing.T) {
		closedAt := now.Add(-1 * time.Minute)
		entries := []repositories.LoginHistoryRecord{
			// 활성 세션 (ended_at = nil) → CloseOrphanedSessions 대상
			{SessionId: "sess-orphan-1", StartedAt: now.Add(-10 * time.Minute), UserId: "orphan", ClientIp: "10.0.0.1"},
			{SessionId: "sess-orphan-2", StartedAt: now.Add(-5 * time.Minute), UserId: "orphan", ClientIp: "10.0.0.2"},
			// 이미 종료된 세션 → 영향받지 않아야 함
			{SessionId: "sess-orphan-3", StartedAt: now.Add(-3 * time.Minute), UserId: "orphan", ClientIp: "10.0.0.3", EndedAt: &closedAt, EndReason: "logout"},
		}
		for _, e := range entries {
			if err := repo.InsertLoginHistory(e); err != nil {
				t.Fatalf("InsertLoginHistory failed: %v", err)
			}
		}
		defer clearLoginHistory(t, repo)

		_, err := r.CloseOrphanedSessions(ctx, now)
		if err != nil {
			t.Fatalf("CloseOrphanedSessions failed: %v", err)
		}

		records, err := r.Query(ctx, "orphan", 100, 0)
		if err != nil {
			t.Fatalf("Query failed: %v", err)
		}
		if len(records) != 3 {
			t.Fatalf("expected 3 records, got %d", len(records))
		}

		for _, rec := range records {
			switch rec.SessionId {
			case "sess-orphan-1", "sess-orphan-2":
				if rec.EndedAt == nil {
					t.Errorf("session %s: expected ended_at to be set", rec.SessionId)
				}
				if rec.EndReason != "server_restart" {
					t.Errorf("session %s: expected end_reason=server_restart, got %q", rec.SessionId, rec.EndReason)
				}
			case "sess-orphan-3":
				if rec.EndReason != "logout" {
					t.Errorf("session %s: expected end_reason=logout, got %q", rec.SessionId, rec.EndReason)
				}
			}
		}
	})
}

func clearLoginHistory(t *testing.T, repo LoginHistoryTestable) {
	t.Helper()
	type clearer interface {
		ClearTable() error
	}
	if c, ok := repo.(clearer); ok {
		if err := c.ClearTable(); err != nil {
			t.Logf("ClearTable error: %v", err)
		}
	}
}
