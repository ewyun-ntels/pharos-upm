package test

import (
	"context"
	"testing"
	"time"

	"ntels.com/pharos/core/internal/repositories"
)

type DashboardHistoryTestable interface {
	repositories.DashboardHistoryRepository
	SetUpDb() error
	TearDownDb() error
	ClearTable() error
	GetCount() (int, error)
}

func DashboardHistorySuite(t *testing.T, testable DashboardHistoryTestable) {
	t.Run("DashboardHistory Repository Suite", func(t *testing.T) {
		t.Run("Record", func(t *testing.T) { DashboardHistoryRecord(t, testable) })
		t.Run("Query", func(t *testing.T) { DashboardHistoryQuery(t, testable) })
		t.Run("QuerySearch", func(t *testing.T) { DashboardHistoryQuerySearch(t, testable) })
		t.Run("Count", func(t *testing.T) { DashboardHistoryCount(t, testable) })
		t.Run("CountSearch", func(t *testing.T) { DashboardHistoryCountSearch(t, testable) })
		t.Run("Prune", func(t *testing.T) { DashboardHistoryPrune(t, testable) })
		t.Run("QueryByDashboardIDs", func(t *testing.T) { DashboardHistoryQueryByDashboardIDs(t, testable) })
		t.Run("CountByDashboardIDs", func(t *testing.T) { DashboardHistoryCountByDashboardIDs(t, testable) })
	})
}

func DashboardHistoryRecord(t *testing.T, testable DashboardHistoryTestable) {
	if err := testable.SetUpDb(); err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})

	rec := repositories.DashboardHistoryRecord{
		DashboardID:    "dash-1",
		DashboardTitle: "My Dashboard",
		Action:         "created",
		ChangedBy:      "admin",
		ChangedAt:      time.Now().In(testTimeZone),
	}
	if err := testable.Record(context.Background(), rec); err != nil {
		t.Fatalf("Record failed: %v", err)
	}

	count, err := testable.GetCount()
	if err != nil {
		t.Fatalf("GetCount failed: %v", err)
	}
	if count != 1 {
		t.Errorf("Expected count 1, got %d", count)
	}
}

func DashboardHistoryQuery(t *testing.T, testable DashboardHistoryTestable) {
	if err := testable.SetUpDb(); err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})
	if err := testable.ClearTable(); err != nil {
		t.Fatalf("ClearTable failed: %v", err)
	}

	actions := []string{"created", "updated", "deleted"}
	base := time.Now()
	for i, action := range actions {
		err := testable.Record(context.Background(), repositories.DashboardHistoryRecord{
			DashboardID:    "dash-1",
			DashboardTitle: "Test",
			Action:         action,
			ChangedBy:      "user",
			ChangedAt:      base.Add(time.Duration(i) * time.Second),
		})
		if err != nil {
			t.Fatalf("Record failed: %v", err)
		}
	}

	rows, err := testable.Query(context.Background(), 10, 0, "")
	if err != nil {
		t.Fatalf("Query failed: %v", err)
	}
	if len(rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}
	// Most recent first
	if rows[0].Action != "deleted" {
		t.Errorf("Expected first row action=deleted, got %s", rows[0].Action)
	}

	// pagination: limit 1
	paged, err := testable.Query(context.Background(), 1, 1, "")
	if err != nil {
		t.Fatalf("Query with pagination failed: %v", err)
	}
	if len(paged) != 1 {
		t.Errorf("Expected 1 paged row, got %d", len(paged))
	}
}

func DashboardHistoryCount(t *testing.T, testable DashboardHistoryTestable) {
	if err := testable.SetUpDb(); err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})
	if err := testable.ClearTable(); err != nil {
		t.Fatalf("ClearTable failed: %v", err)
	}

	for range 5 {
		_ = testable.Record(context.Background(), repositories.DashboardHistoryRecord{
			DashboardID:    "dash",
			DashboardTitle: "Test",
			Action:         "updated",
			ChangedBy:      "user",
			ChangedAt:      time.Now(),
		})
	}
	n, err := testable.Count(context.Background(), "")
	if err != nil {
		t.Fatalf("Count failed: %v", err)
	}
	if n != 5 {
		t.Errorf("Expected count 5, got %d", n)
	}
}

func DashboardHistoryPrune(t *testing.T, testable DashboardHistoryTestable) {
	if err := testable.SetUpDb(); err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})
	if err := testable.ClearTable(); err != nil {
		t.Fatalf("ClearTable failed: %v", err)
	}

	// Insert 5 records and prune to max 3
	base := time.Now()
	for i := range 5 {
		_ = testable.Record(context.Background(), repositories.DashboardHistoryRecord{
			DashboardID:    "dash",
			DashboardTitle: "Test",
			Action:         "updated",
			ChangedBy:      "user",
			ChangedAt:      base.Add(time.Duration(i) * time.Second),
		})
	}
	if err := testable.Prune(context.Background(), 3); err != nil {
		t.Fatalf("Prune failed: %v", err)
	}
	n, err := testable.Count(context.Background(), "")
	if err != nil {
		t.Fatalf("Count after prune failed: %v", err)
	}
	if n != 3 {
		t.Errorf("Expected 3 records after prune, got %d", n)
	}
}

func DashboardHistoryQueryByDashboardIDs(t *testing.T, testable DashboardHistoryTestable) {
	if err := testable.SetUpDb(); err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})
	if err := testable.ClearTable(); err != nil {
		t.Fatalf("ClearTable failed: %v", err)
	}

	base := time.Now()
	// dash-A: 2 records, dash-B: 1 record, dash-C: 1 record (not permitted)
	records := []repositories.DashboardHistoryRecord{
		{DashboardID: "dash-A", DashboardTitle: "A", Action: "created", ChangedBy: "user", ChangedAt: base},
		{DashboardID: "dash-A", DashboardTitle: "A", Action: "updated", ChangedBy: "user", ChangedAt: base.Add(time.Second)},
		{DashboardID: "dash-B", DashboardTitle: "B", Action: "created", ChangedBy: "user", ChangedAt: base.Add(2 * time.Second)},
		{DashboardID: "dash-C", DashboardTitle: "C", Action: "created", ChangedBy: "user", ChangedAt: base.Add(3 * time.Second)},
	}
	for _, r := range records {
		if err := testable.Record(context.Background(), r); err != nil {
			t.Fatalf("Record failed: %v", err)
		}
	}

	// Query only dash-A and dash-B (permitted), dash-C must not appear
	rows, err := testable.QueryByDashboardIDs(context.Background(), []string{"dash-A", "dash-B"}, 10, 0, "")
	if err != nil {
		t.Fatalf("QueryByDashboardIDs failed: %v", err)
	}
	if len(rows) != 3 {
		t.Errorf("Expected 3 rows, got %d", len(rows))
	}
	for _, r := range rows {
		if r.DashboardID == "dash-C" {
			t.Errorf("dash-C must not appear in results")
		}
	}
	// Most recent first
	if rows[0].DashboardID != "dash-B" {
		t.Errorf("Expected first row dashboardID=dash-B, got %s", rows[0].DashboardID)
	}

	// Pagination: limit 1, offset 1
	paged, err := testable.QueryByDashboardIDs(context.Background(), []string{"dash-A", "dash-B"}, 1, 1, "")
	if err != nil {
		t.Fatalf("QueryByDashboardIDs with pagination failed: %v", err)
	}
	if len(paged) != 1 {
		t.Errorf("Expected 1 paged row, got %d", len(paged))
	}

	// Empty IDs → empty result
	empty, err := testable.QueryByDashboardIDs(context.Background(), []string{}, 10, 0, "")
	if err != nil {
		t.Fatalf("QueryByDashboardIDs empty ids failed: %v", err)
	}
	if len(empty) != 0 {
		t.Errorf("Expected 0 rows for empty ids, got %d", len(empty))
	}
}

func DashboardHistoryCountByDashboardIDs(t *testing.T, testable DashboardHistoryTestable) {
	if err := testable.SetUpDb(); err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})
	if err := testable.ClearTable(); err != nil {
		t.Fatalf("ClearTable failed: %v", err)
	}

	base := time.Now()
	for i, id := range []string{"dash-A", "dash-A", "dash-B", "dash-C"} {
		if err := testable.Record(context.Background(), repositories.DashboardHistoryRecord{
			DashboardID:    id,
			DashboardTitle: id,
			Action:         "updated",
			ChangedBy:      "user",
			ChangedAt:      base.Add(time.Duration(i) * time.Second),
		}); err != nil {
			t.Fatalf("Record failed: %v", err)
		}
	}

	// Count only dash-A and dash-B → must be 3
	n, err := testable.CountByDashboardIDs(context.Background(), []string{"dash-A", "dash-B"}, "")
	if err != nil {
		t.Fatalf("CountByDashboardIDs failed: %v", err)
	}
	if n != 3 {
		t.Errorf("Expected count 3, got %d", n)
	}

	// Empty IDs → 0
	n, err = testable.CountByDashboardIDs(context.Background(), []string{}, "")
	if err != nil {
		t.Fatalf("CountByDashboardIDs empty ids failed: %v", err)
	}
	if n != 0 {
		t.Errorf("Expected 0 for empty ids, got %d", n)
	}
}

// DashboardHistoryQuerySearch tests Query with a non-empty search term.
func DashboardHistoryQuerySearch(t *testing.T, testable DashboardHistoryTestable) {
	if err := testable.SetUpDb(); err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})
	if err := testable.ClearTable(); err != nil {
		t.Fatalf("ClearTable failed: %v", err)
	}

	base := time.Now()
	seed := []repositories.DashboardHistoryRecord{
		{DashboardID: "d1", DashboardTitle: "Network Overview", Action: "created", ChangedBy: "alice", ChangedAt: base},
		{DashboardID: "d2", DashboardTitle: "CPU Metrics", Action: "updated", ChangedBy: "bob", ChangedAt: base.Add(time.Second)},
		{DashboardID: "d3", DashboardTitle: "Memory Metrics", Action: "deleted", ChangedBy: "alice", ChangedAt: base.Add(2 * time.Second)},
		{DashboardID: "d4", DashboardTitle: "Disk Usage", Action: "created", ChangedBy: "carol", ChangedAt: base.Add(3 * time.Second)},
	}
	for _, r := range seed {
		if err := testable.Record(context.Background(), r); err != nil {
			t.Fatalf("Record failed: %v", err)
		}
	}

	// search by dashboardTitle partial match ("Metrics" → d2, d3)
	rows, err := testable.Query(context.Background(), 10, 0, "Metrics")
	if err != nil {
		t.Fatalf("Query(search=Metrics) failed: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("Expected 2 rows for search=Metrics, got %d", len(rows))
	}

	// search by action ("created" → d1, d4)
	rows, err = testable.Query(context.Background(), 10, 0, "created")
	if err != nil {
		t.Fatalf("Query(search=created) failed: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("Expected 2 rows for search=created, got %d", len(rows))
	}

	// search by changedBy ("alice" → d1, d3)
	rows, err = testable.Query(context.Background(), 10, 0, "alice")
	if err != nil {
		t.Fatalf("Query(search=alice) failed: %v", err)
	}
	if len(rows) != 2 {
		t.Errorf("Expected 2 rows for search=alice, got %d", len(rows))
	}

	// search with no match
	rows, err = testable.Query(context.Background(), 10, 0, "zzznomatch")
	if err != nil {
		t.Fatalf("Query(search=zzznomatch) failed: %v", err)
	}
	if len(rows) != 0 {
		t.Errorf("Expected 0 rows for no-match search, got %d", len(rows))
	}
}

// DashboardHistoryCountSearch tests Count with a non-empty search term.
func DashboardHistoryCountSearch(t *testing.T, testable DashboardHistoryTestable) {
	if err := testable.SetUpDb(); err != nil {
		t.Fatalf("DB setup failed: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("DB teardown failed: %v", err)
		}
	})
	if err := testable.ClearTable(); err != nil {
		t.Fatalf("ClearTable failed: %v", err)
	}

	base := time.Now()
	seed := []repositories.DashboardHistoryRecord{
		{DashboardID: "d1", DashboardTitle: "Network Overview", Action: "created", ChangedBy: "alice", ChangedAt: base},
		{DashboardID: "d2", DashboardTitle: "CPU Metrics", Action: "updated", ChangedBy: "bob", ChangedAt: base.Add(time.Second)},
		{DashboardID: "d3", DashboardTitle: "Memory Metrics", Action: "deleted", ChangedBy: "alice", ChangedAt: base.Add(2 * time.Second)},
		{DashboardID: "d4", DashboardTitle: "Disk Usage", Action: "created", ChangedBy: "carol", ChangedAt: base.Add(3 * time.Second)},
	}
	for _, r := range seed {
		if err := testable.Record(context.Background(), r); err != nil {
			t.Fatalf("Record failed: %v", err)
		}
	}

	// empty search → all 4
	n, err := testable.Count(context.Background(), "")
	if err != nil {
		t.Fatalf("Count(search='') failed: %v", err)
	}
	if n != 4 {
		t.Errorf("Expected count 4 for empty search, got %d", n)
	}

	// search by title ("Metrics" → 2)
	n, err = testable.Count(context.Background(), "Metrics")
	if err != nil {
		t.Fatalf("Count(search=Metrics) failed: %v", err)
	}
	if n != 2 {
		t.Errorf("Expected count 2 for search=Metrics, got %d", n)
	}

	// search by changedBy ("alice" → 2)
	n, err = testable.Count(context.Background(), "alice")
	if err != nil {
		t.Fatalf("Count(search=alice) failed: %v", err)
	}
	if n != 2 {
		t.Errorf("Expected count 2 for search=alice, got %d", n)
	}

	// no match → 0
	n, err = testable.Count(context.Background(), "zzznomatch")
	if err != nil {
		t.Fatalf("Count(search=zzznomatch) failed: %v", err)
	}
	if n != 0 {
		t.Errorf("Expected count 0 for no-match search, got %d", n)
	}
}
