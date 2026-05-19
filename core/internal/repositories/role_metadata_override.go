package repositories

import (
	"context"
	"time"

	sharedRole "ntels.com/pharos/shared/types/role"
)

// RoleMetadataConfigEntity represents entire role metadata configuration (Kustomize-style overlay)
// Stores complete []RoleMetadata array as atomic unit
type RoleMetadataConfigEntity struct {
	ID          string                    // UUID primary key
	ConfigJSON  string                    // Complete []RoleMetadata array as JSON
	Config      []sharedRole.RoleMetadata // Parsed config (not stored in DB)
	CreatedAt   time.Time                 // Creation timestamp (latest = active)
	UpdatedBy   *string                   // Username who made this change
	Description *string                   // Optional change description
}

// RoleMetadataConfigRepository manages role metadata configuration with version history
// Implements Kustomize-style overlay pattern: Base + Overlay = Final Config
type RoleMetadataConfigRepository interface {
	// GetActiveConfig retrieves the latest configuration (most recent by created_at)
	// Returns nil if no configuration exists
	GetActiveConfig(ctx context.Context) (*RoleMetadataConfigEntity, error)

	// GetHistory retrieves configuration history (up to limit, ordered by timestamp desc)
	GetHistory(ctx context.Context, limit int) ([]RoleMetadataConfigEntity, error)

	// SaveConfig saves a new configuration version (appends to history)
	// Automatically maintains only the latest 'maxHistory' versions
	SaveConfig(ctx context.Context, config RoleMetadataConfigEntity, maxHistory int) error

	// Rollback rolls back to a specific version by id
	// Creates a new record with the old configuration
	Rollback(ctx context.Context, id string, updatedBy *string, maxHistory int) error

	// DeleteAll removes all configuration history (reset to base)
	DeleteAll(ctx context.Context) error
}
