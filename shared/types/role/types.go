// Code generated from JSON Schema using quicktype. DO NOT EDIT.
// To parse and unparse this JSON data, add this code to your project and do:
//
//    permissionKeys, err := UnmarshalPermissionKeys(bytes)
//    bytes, err = permissionKeys.Marshal()
//
//    roleGroup, err := UnmarshalRoleGroup(bytes)
//    bytes, err = roleGroup.Marshal()
//
//    roleMetadata, err := UnmarshalRoleMetadata(bytes)
//    bytes, err = roleMetadata.Marshal()
//
//    roleMetadataConfigEntity, err := UnmarshalRoleMetadataConfigEntity(bytes)
//    bytes, err = roleMetadataConfigEntity.Marshal()
//
//    roleMetadataConfigHistoryResponse, err := UnmarshalRoleMetadataConfigHistoryResponse(bytes)
//    bytes, err = roleMetadataConfigHistoryResponse.Marshal()
//
//    roleMetadataConfigResponse, err := UnmarshalRoleMetadataConfigResponse(bytes)
//    bytes, err = roleMetadataConfigResponse.Marshal()

package role

import "encoding/json"

func UnmarshalPermissionKeys(data []byte) (PermissionKeys, error) {
	var r PermissionKeys
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *PermissionKeys) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalRoleGroup(data []byte) (RoleGroup, error) {
	var r RoleGroup
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *RoleGroup) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalRoleMetadata(data []byte) (RoleMetadata, error) {
	var r RoleMetadata
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *RoleMetadata) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalRoleMetadataConfigEntity(data []byte) (RoleMetadataConfigEntity, error) {
	var r RoleMetadataConfigEntity
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *RoleMetadataConfigEntity) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalRoleMetadataConfigHistoryResponse(data []byte) (RoleMetadataConfigHistoryResponse, error) {
	var r RoleMetadataConfigHistoryResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *RoleMetadataConfigHistoryResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

func UnmarshalRoleMetadataConfigResponse(data []byte) (RoleMetadataConfigResponse, error) {
	var r RoleMetadataConfigResponse
	err := json.Unmarshal(data, &r)
	return r, err
}

func (r *RoleMetadataConfigResponse) Marshal() ([]byte, error) {
	return json.Marshal(r)
}

// Represents a group of roles (e.g., Core, CATV, etc.)
type RoleGroup struct {
	// User-friendly display name for the group
	DisplayName string `json:"displayName"`
	// Group identifier (e.g., 'core', 'catv')
	Group string         `json:"group"`
	Roles []RoleMetadata `json:"roles"`
}

// Metadata for a role (for UI display purposes). Supports both atomic roles and composite
// roles
type RoleMetadata struct {
	// Description of what this role allows
	Description string `json:"description"`
	// User-friendly display name (e.g., '사용자 조회', '품질 지표 조회')
	DisplayName string `json:"displayName"`
	// Role group type for categorization (recommended: role, attribute, permission, or custom
	// category)
	Group string `json:"group"`
	// Whether to hide this role from UI selection (e.g., internal or system roles). If false or
	// nil, the role is visible and selectable in UI
	Hide *bool `json:"hide,omitempty"`
	// Role key (e.g., 'role:user_read', 'extension:catv:view_quality')
	Key string `json:"key"`
	// Composite role definition: map of sub-roles this role includes. If nil or empty, this is
	// an atomic role. If non-empty, this is a composite role that expands to included roles
	Roles map[string]bool `json:"roles,omitempty"`
}

type RoleMetadataConfigHistoryResponse struct {
	History []RoleMetadataConfigEntity `json:"history"`
}

type RoleMetadataConfigEntity struct {
	Config      []RoleMetadata `json:"config"`
	CreatedAt   string         `json:"created_at"`
	Description *string        `json:"description,omitempty"`
	ID          string         `json:"id"`
	UpdatedBy   *string        `json:"updated_by,omitempty"`
}

type RoleMetadataConfigResponse struct {
	Config  []RoleMetadata `json:"config,omitempty"`
	Message *string        `json:"message,omitempty"`
}

// All available permission keys in the system
type PermissionKeys string

const (
	AttrPasswordChange         PermissionKeys = "attr:password_change"
	AttrPasswordRetryUnlimited PermissionKeys = "attr:password_retry_unlimited"
	AttrSkipTemporarilyBlock   PermissionKeys = "attr:skip_temporarily_block"
	AttrTemporaryUser          PermissionKeys = "attr:temporary_user"
	RoleAlertCreate            PermissionKeys = "role:alert_create"
	RoleAlertDelete            PermissionKeys = "role:alert_delete"
	RoleAlertRead              PermissionKeys = "role:alert_read"
	RoleAlertUpdate            PermissionKeys = "role:alert_update"
	RoleDashboardRead          PermissionKeys = "role:dashboard_read"
	RoleDashboardCreate        PermissionKeys = "role:dashboard_create"
	RoleGroupAdd               PermissionKeys = "role:group_add"
	RoleNotificationCreate     PermissionKeys = "role:notification_create"
	RoleNotificationDelete     PermissionKeys = "role:notification_delete"
	RoleNotificationRead       PermissionKeys = "role:notification_read"
	RoleNotificationUpdate     PermissionKeys = "role:notification_update"
	RolePodDelete              PermissionKeys = "role:pod_delete"
	RoleSuperAdmin             PermissionKeys = "role:super_admin"
	RoleUserCreate             PermissionKeys = "role:user_create"
	RoleUserDelete             PermissionKeys = "role:user_delete"
	RoleUserRead               PermissionKeys = "role:user_read"
	RoleUserUpdate             PermissionKeys = "role:user_update"
)
