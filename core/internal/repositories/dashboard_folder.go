package repositories

import (
	"context"
	"time"
)

// DashboardFolderEntity represents a dashboard folder
type DashboardFolderEntity struct {
	ID        string
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// DashboardFolderRepository defines the interface for dashboard folder data access
type DashboardFolderRepository interface {
	// Create inserts a new folder
	Create(ctx context.Context, folder *DashboardFolderEntity) error

	// GetByID retrieves a folder by its ID
	GetByID(ctx context.Context, id string) (*DashboardFolderEntity, error)

	// ListAll retrieves all folders
	ListAll(ctx context.Context) ([]*DashboardFolderEntity, error)

	// Update renames a folder
	Update(ctx context.Context, folder *DashboardFolderEntity) error

	// DeleteByID deletes a folder by its ID
	DeleteByID(ctx context.Context, id string) error

	// MoveDashboard sets (or clears) the folder_id on a dashboard row.
	// Pass nil folderID to remove the dashboard from its folder.
	MoveDashboard(ctx context.Context, dashboardID string, folderID *string) error
}
