package role

import (
	"context"
	"encoding/json"

	"ntels.com/pharos/core/internal/repositories"
	sharedRole "ntels.com/pharos/shared/types/role"
)

const (
	// MaxHistoryVersions is the maximum number of history versions to keep
	MaxHistoryVersions = 10
)

// MetadataService handles role metadata with Kustomize-style configuration overlay
type MetadataService struct {
	configRepo repositories.RoleMetadataConfigRepository
}

// NewMetadataService creates a new metadata service
func NewMetadataService(configRepo repositories.RoleMetadataConfigRepository) *MetadataService {
	return &MetadataService{
		configRepo: configRepo,
	}
}

// ApplyOverlay applies configuration overlay to base metadata.
// - Existing base entries are overridden by matching overlay entries (patch).
// - Composite roles (overlay entries with a non-empty Roles map) are appended.
// Use this for core roles where the overlay can also define new composite roles.
func (s *MetadataService) ApplyOverlay(ctx context.Context, base []sharedRole.RoleMetadata) ([]sharedRole.RoleMetadata, error) {
	if s.configRepo == nil {
		return base, nil
	}

	configEntity, err := s.configRepo.GetActiveConfig(ctx)
	if err != nil {
		return nil, err
	}
	if configEntity == nil {
		return base, nil
	}

	return mergeMetadata(base, configEntity.Config), nil
}

// mergeMetadata applies the overlay onto the base:
//   - base entries whose key appears in the overlay are replaced by the overlay entry
//   - overlay entries whose key is NOT in base are appended (e.g. composite roles)
func mergeMetadata(base, overlay []sharedRole.RoleMetadata) []sharedRole.RoleMetadata {
	overlayMap := make(map[string]sharedRole.RoleMetadata, len(overlay))
	for _, m := range overlay {
		overlayMap[m.Key] = m
	}
	baseKeys := make(map[string]bool, len(base))
	result := make([]sharedRole.RoleMetadata, 0, len(base))
	for _, m := range base {
		baseKeys[m.Key] = true
		if ov, ok := overlayMap[m.Key]; ok {
			result = append(result, ov)
		} else {
			result = append(result, m)
		}
	}
	for _, m := range overlay {
		if !baseKeys[m.Key] {
			result = append(result, m)
		}
	}
	return result
}

// GetActiveConfig retrieves the active configuration overlay from DB
func (s *MetadataService) GetActiveConfig(ctx context.Context) ([]sharedRole.RoleMetadata, error) {
	if s.configRepo == nil {
		return nil, nil
	}

	configEntity, err := s.configRepo.GetActiveConfig(ctx)
	if err != nil {
		return nil, err
	}

	if configEntity == nil {
		return nil, nil
	}

	return configEntity.Config, nil
}

// GetHistory retrieves configuration history
func (s *MetadataService) GetHistory(ctx context.Context, limit int) ([]repositories.RoleMetadataConfigEntity, error) {
	if s.configRepo == nil {
		return nil, nil
	}

	if limit <= 0 {
		limit = MaxHistoryVersions
	}

	return s.configRepo.GetHistory(ctx, limit)
}

// SaveConfig saves entire configuration overlay (replaces all)
func (s *MetadataService) SaveConfig(ctx context.Context, config []sharedRole.RoleMetadata, updatedBy *string, description *string) error {
	if s.configRepo == nil {
		return nil
	}

	// Marshal config to JSON
	configJSON, err := json.Marshal(config)
	if err != nil {
		return err
	}

	entity := repositories.RoleMetadataConfigEntity{
		ConfigJSON:  string(configJSON),
		Config:      config,
		UpdatedBy:   updatedBy,
		Description: description,
	}

	return s.configRepo.SaveConfig(ctx, entity, MaxHistoryVersions)
}

// Rollback rolls back to a specific configuration version
func (s *MetadataService) Rollback(ctx context.Context, id string, updatedBy *string) error {
	if s.configRepo == nil {
		return nil
	}

	return s.configRepo.Rollback(ctx, id, updatedBy, MaxHistoryVersions)
}

// DeleteConfig removes all configuration overlay (revert to base)
func (s *MetadataService) DeleteConfig(ctx context.Context) error {
	if s.configRepo == nil {
		return nil
	}
	return s.configRepo.DeleteAll(ctx)
}

// ExpandCompositeRoles expands composite roles in the given roles map
// Composite roles (roles with non-empty Roles field) are recursively expanded to atomic roles
// Returns expanded map with all atomic roles
func (s *MetadataService) ExpandCompositeRoles(ctx context.Context, rolesMap map[string]bool) (map[string]bool, error) {
	// Get merged metadata (base + overlay from DB)
	allMetadata, err := s.ApplyOverlay(ctx, GetCoreRoleMetadata())
	if err != nil {
		return rolesMap, err
	}

	// Create metadata lookup map
	metadataMap := make(map[string]sharedRole.RoleMetadata)
	for _, m := range allMetadata {
		metadataMap[m.Key] = m
	}

	// Expand roles recursively
	visited := make(map[string]bool)
	result := make(map[string]bool)

	var expand func(roleKey string)
	expand = func(roleKey string) {
		// Prevent infinite loop
		if visited[roleKey] {
			return
		}
		visited[roleKey] = true

		// Add current role to result
		result[roleKey] = true

		// Check if it's a composite role
		if metadata, exists := metadataMap[roleKey]; exists {
			if len(metadata.Roles) > 0 {
				// Composite role - expand sub-roles
				for subRole := range metadata.Roles {
					expand(subRole)
				}
			}
		}
	}

	// Expand all roles in input map
	for roleKey := range rolesMap {
		expand(roleKey)
	}

	return result, nil
}
