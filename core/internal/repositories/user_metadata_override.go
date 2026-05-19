package repositories

import (
	"context"
	"time"

	sharedUser "ntels.com/pharos/shared/types/user"
)

// UserMetadataConfigEntity represents entire user metadata configuration (JSON Schema for additional fields)
// Stores complete UserMetadataConfig as atomic unit
type UserMetadataConfigEntity struct {
	ID          string                         // UUID primary key
	ConfigJSON  string                         // Complete UserMetadataConfig as JSON
	Config      *sharedUser.UserMetadataConfig // Parsed config (not stored in DB)
	CreatedAt   time.Time                      // Creation timestamp (latest = active)
	UpdatedBy   *string                        // Username who made this change
	Description *string                        // Optional change description
}

// UserMetadataConfigRepository manages user metadata configuration with version history
// Pattern: Base + Overlay = Final Config (similar to role metadata)
type UserMetadataConfigRepository interface {
	// GetActiveConfig retrieves the latest configuration (most recent by created_at)
	// Returns nil if no configuration exists
	GetActiveConfig(ctx context.Context) (*UserMetadataConfigEntity, error)

	// GetHistory retrieves configuration history (most recent first)
	// limit: maximum number of entries to return
	GetHistory(ctx context.Context, limit int) ([]UserMetadataConfigEntity, error)

	// SaveConfig saves new configuration and manages history retention
	// Creates new entry with generated UUID
	// maxHistory: maximum number of history entries to keep (older entries are deleted)
	SaveConfig(ctx context.Context, config UserMetadataConfigEntity, maxHistory int) error

	// GetByID retrieves specific configuration by UUID (for rollback)
	GetByID(ctx context.Context, id string) (*UserMetadataConfigEntity, error)
}
