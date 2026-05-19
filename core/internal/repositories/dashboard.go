package repositories

import (
	"context"

	"ntels.com/pharos/shared/types/dashboard"
)

// DashboardEntity represents a dashboard entity
type DashboardEntity struct {
	ID          string
	Config      dashboard.DashboardConfig
	FolderID    *string
	Annotations []dashboard.Annotation
}

// DashboardRepository defines the interface for dashboard data access
type DashboardRepository interface {
	// Create inserts a new dashboard into the database
	Create(ctx context.Context, dashboard *DashboardEntity) error

	// GetByID retrieves a dashboard by its ID
	GetByID(ctx context.Context, id string) (*DashboardEntity, error)

	// ListAll retrieves all dashboards from the database
	ListAll(ctx context.Context) ([]*DashboardEntity, error)

	// Update updates an existing dashboard's config
	Update(ctx context.Context, dashboard *DashboardEntity) error

	// UpdateAnnotations replaces the annotations list for a dashboard
	UpdateAnnotations(ctx context.Context, id string, annotations []dashboard.Annotation) error

	// DeleteByID deletes a dashboard by its ID
	DeleteByID(ctx context.Context, id string) error
}
