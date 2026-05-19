package role

import (
	"fmt"
	"strings"

	sharedRole "ntels.com/pharos/shared/types/role"
)

// ValidationError represents a configuration validation error
type ValidationError struct {
	Field   string
	Message string
	Index   int // Index in config array where error occurred
}

func (e *ValidationError) Error() string {
	if e.Index >= 0 {
		return fmt.Sprintf("config[%d].%s: %s", e.Index, e.Field, e.Message)
	}
	return fmt.Sprintf("%s: %s", e.Field, e.Message)
}

// ValidateConfig validates role metadata configuration
func ValidateConfig(config []sharedRole.RoleMetadata) error {
	if len(config) == 0 {
		// Empty config is valid (equivalent to delete all)
		return nil
	}

	// Track keys to detect duplicates
	keysSeen := make(map[string]int)

	for i, role := range config {
		// 1. Key is required and must not be empty
		if strings.TrimSpace(role.Key) == "" {
			return &ValidationError{
				Field:   "key",
				Message: "key is required and cannot be empty",
				Index:   i,
			}
		}

		// 2. DisplayName is required and must not be empty
		if strings.TrimSpace(role.DisplayName) == "" {
			return &ValidationError{
				Field:   "displayName",
				Message: "displayName is required and cannot be empty",
				Index:   i,
			}
		}

		// 3. Check for duplicate keys
		if prevIndex, exists := keysSeen[role.Key]; exists {
			return &ValidationError{
				Field:   "key",
				Message: fmt.Sprintf("duplicate key '%s' (also at index %d)", role.Key, prevIndex),
				Index:   i,
			}
		}
		keysSeen[role.Key] = i

		// 4. Group should not be empty (optional but recommended)
		if strings.TrimSpace(role.Group) == "" {
			// This is a warning, not an error - we'll allow it but could log
			// For now, just skip validation
		}

		// 5. If Roles is specified (composite role), validate it's not empty
		if role.Roles != nil && len(role.Roles) == 0 {
			return &ValidationError{
				Field:   "roles",
				Message: "composite role must have at least one role specified",
				Index:   i,
			}
		}

		// 6. Validate role keys in composite roles (should follow key format)
		if role.Roles != nil {
			for roleKey := range role.Roles {
				if strings.TrimSpace(roleKey) == "" {
					return &ValidationError{
						Field:   "roles",
						Message: "role key in composite role cannot be empty",
						Index:   i,
					}
				}
			}
		}
	}

	return nil
}
