package repositories

import (
	"context"
	"time"
)

// DashboardHistoryRecord represents a single dashboard change event
type DashboardHistoryRecord struct {
	ID             string
	DashboardID    string
	DashboardTitle string
	Action         string // "created" | "updated" | "deleted"
	ChangedBy      string
	ChangedAt      time.Time
	ConfigJSON     *string // JSON snapshot of the dashboard config at this point in time
}

// DashboardHistoryRepository defines the interface for dashboard history data access
type DashboardHistoryRepository interface {
	// Record inserts a new history entry
	Record(ctx context.Context, record DashboardHistoryRecord) error

	// Query retrieves history records ordered by changed_at DESC
	Query(ctx context.Context, limit, offset int, search string) ([]DashboardHistoryRecord, error)

	// QueryByDashboardIDs retrieves history records for specified dashboard IDs only
	QueryByDashboardIDs(ctx context.Context, ids []string, limit, offset int, search string) ([]DashboardHistoryRecord, error)

	// Count returns the total number of history records
	Count(ctx context.Context, search string) (int, error)

	// CountByDashboardIDs returns the total number of history records for specified dashboard IDs
	CountByDashboardIDs(ctx context.Context, ids []string, search string) (int, error)

	// Prune deletes oldest records so that at most maxKeep remain
	Prune(ctx context.Context, maxKeep int) error

	// GetByID retrieves a single history record by its ID
	GetByID(ctx context.Context, historyID string) (*DashboardHistoryRecord, error)

	// DeleteByDashboardID removes all history entries for the given dashboard
	DeleteByDashboardID(ctx context.Context, dashboardID string) error
}
