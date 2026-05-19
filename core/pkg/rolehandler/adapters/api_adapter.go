package adapters

import (
	"ntels.com/pharos/shared/types/role"
)

// APIAdapter handles conversion between role types and API response types
type APIAdapter struct{}

// ToAPIRoleMetadata converts role.RoleMetadata (internal) to role.RoleMetadata (API)
// Currently they are the same type, but this is kept for consistency with other adapters
func (a *APIAdapter) ToAPIRoleMetadata(internal role.RoleMetadata) role.RoleMetadata {
	return internal
}

// ToAPIRoleMetadataList converts slice of role.RoleMetadata
func (a *APIAdapter) ToAPIRoleMetadataList(internals []role.RoleMetadata) []role.RoleMetadata {
	return internals
}
