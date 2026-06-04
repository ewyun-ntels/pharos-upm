package role

import (
	sharedRole "ntels.com/pharos/shared/types/role"
)

// Recommended group values for role categorization
const (
	GroupRole      = "role"      // Standard roles (e.g., user_read, dashboard_create)
	GroupAttribute = "attribute" // User attributes (e.g., password_change, temporary_user)
)

// CoreRoleMetadata provides metadata for core atomic roles
// This is used for UI display purposes only
var CoreRoleMetadata = []sharedRole.RoleMetadata{
	// User management roles
	{
		Key:         string(sharedRole.RoleUserRead),
		DisplayName: "User Read",
		Description: "View user information",
		Group:       GroupRole,
	},
	{
		Key:         string(sharedRole.RoleUserCreate),
		DisplayName: "User Create",
		Description: "Create new users",
		Group:       GroupRole,
	},
	{
		Key:         string(sharedRole.RoleUserUpdate),
		DisplayName: "User Update",
		Description: "Update user information",
		Group:       GroupRole,
	},
	{
		Key:         string(sharedRole.RoleUserDelete),
		DisplayName: "User Delete",
		Description: "Delete users",
		Group:       GroupRole,
	},

	// Dashboard roles
	{
		Key:         string(sharedRole.RoleDashboardRead),
		DisplayName: "Dashboard Read",
		Description: "View dashboard menu and dashboards",
		Group:       GroupRole,
	},
	{
		Key:         string(sharedRole.RoleDashboardCreate),
		DisplayName: "Dashboard Create",
		Description: "Create new dashboards",
		Group:       GroupRole,
	},

	// Group roles
	{
		Key:         string(sharedRole.RoleGroupAdd),
		DisplayName: "Group Add",
		Description: "Add user groups",
		Group:       GroupRole,
	},

	// Alert roles
	{
		Key:         string(sharedRole.RoleAlertRead),
		DisplayName: "Alert Read",
		Description: "View alerts",
		Group:       GroupRole,
	},
	{
		Key:         string(sharedRole.RoleAlertCreate),
		DisplayName: "Alert Create",
		Description: "Create new alerts",
		Group:       GroupRole,
	},
	{
		Key:         string(sharedRole.RoleAlertUpdate),
		DisplayName: "Alert Update",
		Description: "Update alerts",
		Group:       GroupRole,
	},
	{
		Key:         string(sharedRole.RoleAlertDelete),
		DisplayName: "Alert Delete",
		Description: "Delete alerts",
		Group:       GroupRole,
	},

	// Notification roles
	{
		Key:         string(sharedRole.RoleNotificationRead),
		DisplayName: "Notification Read",
		Description: "View notifications",
		Group:       GroupRole,
		Hide:        boolPtr(true),
	},
	{
		Key:         string(sharedRole.RoleNotificationCreate),
		DisplayName: "Notification Create",
		Description: "Create new notifications",
		Group:       GroupRole,
		Hide:        boolPtr(true),
	},
	{
		Key:         string(sharedRole.RoleNotificationUpdate),
		DisplayName: "Notification Update",
		Description: "Update notifications",
		Group:       GroupRole,
		Hide:        boolPtr(true),
	},
	{
		Key:         string(sharedRole.RoleNotificationDelete),
		DisplayName: "Notification Delete",
		Description: "Delete notifications",
		Group:       GroupRole,
		Hide:        boolPtr(true),
	},

	// UPM roles
	{
		Key:         string(sharedRole.RolePodDelete),
		DisplayName: "Pod Delete",
		Description: "Delete Kubernetes pods from the home topology",
		Group:       GroupRole,
	},

	// Super admin
	{
		Key:         string(sharedRole.RoleSuperAdmin),
		DisplayName: "Super Admin",
		Description: "Full system administrator with all permissions",
		Group:       GroupRole,
	},

	// Attributes - hidden from UI
	{
		Key:         string(sharedRole.AttrPasswordChange),
		DisplayName: "Password Change",
		Description: "Can change password",
		Group:       GroupAttribute,
		Hide:        boolPtr(true),
	},
	{
		Key:         string(sharedRole.AttrPasswordRetryUnlimited),
		DisplayName: "Unlimited Password Retries",
		Description: "No limit on password retry attempts",
		Group:       GroupAttribute,
		Hide:        boolPtr(true),
	},
	{
		Key:         string(sharedRole.AttrSkipTemporarilyBlock),
		DisplayName: "Skip Temporary Block",
		Description: "Skip temporary block due to login failures",
		Group:       GroupAttribute,
		Hide:        boolPtr(true),
	},
	{
		Key:         string(sharedRole.AttrTemporaryUser),
		DisplayName: "Temporary User",
		Description: "Temporary user attribute",
		Group:       GroupAttribute,
		Hide:        boolPtr(true),
	},
}

// GetCoreRoleMetadata returns metadata for core atomic roles
func GetCoreRoleMetadata() []sharedRole.RoleMetadata {
	return CoreRoleMetadata
}

// boolPtr returns a pointer to a bool value
func boolPtr(b bool) *bool {
	return &b
}
