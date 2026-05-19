package user

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"ntels.com/pharos/core/internal/repositories"
	sharedUser "ntels.com/pharos/shared/types/user"
)

const (
	// DefaultMaxHistory is the default number of history entries to keep
	DefaultMaxHistory = 50
)

// MetadataService manages user metadata configuration (JSON Schema for additional fields)
type MetadataService struct {
	repo repositories.UserMetadataConfigRepository
}

// NewMetadataService creates a new metadata service
func NewMetadataService(repo repositories.UserMetadataConfigRepository) *MetadataService {
	return &MetadataService{
		repo: repo,
	}
}

// GetActiveConfig retrieves the current active user metadata configuration
// Returns nil if no configuration exists (returns empty schema as fallback)
func (s *MetadataService) GetActiveConfig(ctx context.Context) (*sharedUser.UserMetadataConfig, error) {
	entity, err := s.repo.GetActiveConfig(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get active config: %w", err)
	}

	// Return empty schema if no config exists
	if entity == nil || entity.Config == nil {
		return &sharedUser.UserMetadataConfig{
			Schema: map[string]any{
				"type":       "object",
				"properties": map[string]any{},
			},
			UISchema: map[string]any{},
		}, nil
	}

	return entity.Config, nil
}

// SaveConfig saves new user metadata configuration with version history
func (s *MetadataService) SaveConfig(ctx context.Context, config *sharedUser.UserMetadataConfig, updatedBy *string, description *string) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}

	// Validate schema structure (basic validation)
	if err := validateUserMetadataConfig(config); err != nil {
		return fmt.Errorf("invalid config: %w", err)
	}

	// Marshal to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	entity := repositories.UserMetadataConfigEntity{
		ConfigJSON:  string(configJSON),
		UpdatedBy:   updatedBy,
		Description: description,
	}

	if err := s.repo.SaveConfig(ctx, entity, DefaultMaxHistory); err != nil {
		return fmt.Errorf("failed to save config: %w", err)
	}

	return nil
}

// GetHistory retrieves configuration history
func (s *MetadataService) GetHistory(ctx context.Context, limit int) ([]repositories.UserMetadataConfigEntity, error) {
	if limit <= 0 {
		limit = 10
	}
	if limit > 100 {
		limit = 100
	}

	history, err := s.repo.GetHistory(ctx, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to get history: %w", err)
	}

	return history, nil
}

// Rollback restores a previous configuration by ID
func (s *MetadataService) Rollback(ctx context.Context, historyID string, updatedBy *string) error {
	// Get the historical config
	historical, err := s.repo.GetByID(ctx, historyID)
	if err != nil {
		return fmt.Errorf("failed to get historical config: %w", err)
	}

	if historical == nil {
		return errors.New("historical config not found")
	}

	// Create new entry with same config but new timestamp
	desc := fmt.Sprintf("Rollback to version %s", historyID)
	entity := repositories.UserMetadataConfigEntity{
		ConfigJSON:  historical.ConfigJSON,
		UpdatedBy:   updatedBy,
		Description: &desc,
	}

	if err := s.repo.SaveConfig(ctx, entity, DefaultMaxHistory); err != nil {
		return fmt.Errorf("failed to rollback config: %w", err)
	}

	return nil
}

// validateUserMetadataConfig performs basic validation on the config
func validateUserMetadataConfig(config *sharedUser.UserMetadataConfig) error {
	if config == nil {
		return errors.New("config cannot be nil")
	}

	if config.Schema == nil {
		return errors.New("schema is required")
	}

	// Basic check: schema should have type and properties
	if schemaType, ok := config.Schema["type"].(string); !ok || schemaType != "object" {
		return errors.New("schema.type must be 'object'")
	}

	// Properties optional but if present, should be an object
	if props, ok := config.Schema["properties"]; ok {
		if _, isMap := props.(map[string]any); !isMap {
			return errors.New("schema.properties must be an object")
		}
	}

	return nil
}
