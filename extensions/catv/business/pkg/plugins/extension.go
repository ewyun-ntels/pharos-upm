package plugins

import (
	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/plugins/model"
	"ntels.com/pharos/extensions/catv/business/pkg/permissions"
	sharedRole "ntels.com/pharos/shared/types/role"
)

// CatvPluginExtension implements plugins.Extension interface for CATV
// This is separate from the server.Extension implementation
type CatvPluginExtension struct{}

// NewCatvPluginExtension creates a new CATV plugin extension
func NewCatvPluginExtension() *CatvPluginExtension {
	return &CatvPluginExtension{}
}

// GetPlugin returns nil as CATV is not a datasource plugin
func (e *CatvPluginExtension) GetPlugin() (*model.Plugin, error) {
	// CATV is not a datasource plugin, so return nil
	return nil, nil
}

// GetName returns the extension name
func (e *CatvPluginExtension) GetName() string {
	return "CATV Quality Monitor"
}

// GetVersion returns the extension version
func (e *CatvPluginExtension) GetVersion() string {
	return "1.0.0"
}

// GetType returns the extension type
func (e *CatvPluginExtension) GetType() string {
	return "feature"
}

// RegisterRoutes registers CATV extension routes
// Routes are already registered by catv server, so this is a no-op
func (e *CatvPluginExtension) RegisterRoutes(_ *gin.RouterGroup) {
	// Routes already registered by catv/pkg/server
}

// GetRoles returns the roles provided by CATV extension
func (e *CatvPluginExtension) GetRoles() []sharedRole.RoleMetadata {
	return []sharedRole.RoleMetadata{
		{
			Key:         permissions.Read,
			DisplayName: "CATV Read",
			Description: "View CATV monitoring data and schedules",
			Group:       "permission",
		},
		{
			Key:         permissions.Create,
			DisplayName: "CATV Create",
			Description: "Create CATV monitoring schedules",
			Group:       "permission",
		},
		{
			Key:         permissions.Update,
			DisplayName: "CATV Update",
			Description: "Modify CATV monitoring schedules",
			Group:       "permission",
		},
		{
			Key:         permissions.Delete,
			DisplayName: "CATV Delete",
			Description: "Remove CATV monitoring schedules",
			Group:       "permission",
		},
	}
}
