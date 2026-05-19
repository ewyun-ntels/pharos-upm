package clickhousedatasource

import (
	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/model"
	"ntels.com/pharos/core/pkg/plugins/plugin/datasource/clickhouse"
	sharedRole "ntels.com/pharos/shared/types/role"
)

// ClickHouseExtension represents the ClickHouse datasource extension
type ClickHouseExtension struct {
	config common.Config
}

// NewClickHouseExtension creates a new ClickHouse extension instance
func NewClickHouseExtension(config common.Config) *ClickHouseExtension {
	return &ClickHouseExtension{
		config: config,
	}
}

// GetPlugin returns the ClickHouse plugin configuration
// This reuses the existing built-in clickhouse implementation
func (c *ClickHouseExtension) GetPlugin() (*model.Plugin, error) {
	// Use the existing clickhouse plugin implementation
	return clickhouse.GetPlugin(c.config)
}

// GetName returns the extension name
func (c *ClickHouseExtension) GetName() string {
	return "clickhouse-datasource"
}

// GetVersion returns the extension version
func (c *ClickHouseExtension) GetVersion() string {
	return "1.0.0"
}

// GetType returns the extension type
func (c *ClickHouseExtension) GetType() string {
	return "datasource"
}

// RegisterRoutes registers extension routes (currently no custom routes)
func (c *ClickHouseExtension) RegisterRoutes(_ *gin.RouterGroup) {
	// No custom routes for now - uses standard datasource API
	// Future: Add custom routes like:
	// router.GET("/schema", c.getSchemaInfo)
	// router.POST("/preview", c.previewQuery)
}

// GetRoles returns the roles provided by this extension
func (c *ClickHouseExtension) GetRoles() []sharedRole.RoleMetadata {
	// ClickHouse datasource extension does not provide custom roles
	return []sharedRole.RoleMetadata{}
}
