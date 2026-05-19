package plugins

import (
	"log/slog"
	"maps"
	"sync"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/plugins/model"
	"ntels.com/pharos/core/pkg/plugins/shared"
	sharedRole "ntels.com/pharos/shared/types/role"
)

// Extension represents a datasource extension interface
type Extension interface {
	GetPlugin() (*model.Plugin, error)
	RegisterRoutes(router *gin.RouterGroup)
	GetName() string
	GetVersion() string
	GetType() string
	GetRoles() []sharedRole.RoleMetadata // Extension이 제공하는 role 목록 (UI용 메타데이터)
}

// ExtensionRegistry manages registered extensions
type ExtensionRegistry struct {
	extensions map[string]Extension
	mu         sync.RWMutex
}

// Global extension registry
var extensionRegistry = &ExtensionRegistry{
	extensions: make(map[string]Extension),
}

// RegisterExtension registers an extension with the global registry
func RegisterExtension(extension Extension) {
	extensionRegistry.Register(extension)
}

// GetExtensionRegistry returns the global extension registry
func GetExtensionRegistry() *ExtensionRegistry {
	return extensionRegistry
}

// Register registers an extension
func (r *ExtensionRegistry) Register(extension Extension) {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := extension.GetName()
	r.extensions[name] = extension

	slog.Info("Extension registered",
		"name", name,
		"version", extension.GetVersion(),
		"type", extension.GetType(),
	)
}

// Get retrieves an extension by name
func (r *ExtensionRegistry) Get(name string) (Extension, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	ext, exists := r.extensions[name]
	return ext, exists
}

// GetAll returns all registered extensions
func (r *ExtensionRegistry) GetAll() map[string]Extension {
	r.mu.RLock()
	defer r.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make(map[string]Extension)
	maps.Copy(result, r.extensions)
	return result
}

// LoadExtensions loads all registered extensions into the plugin system
func (a *Api) loadExtensions() error {
	extensions := extensionRegistry.GetAll()

	for name, extension := range extensions {
		slog.Info("Loading extension", "name", name)

		// Get the plugin from extension
		plugin, err := extension.GetPlugin()
		if err != nil {
			slog.Error("Failed to get plugin from extension",
				"name", name,
				"error", err.Error(),
			)
			continue
		}

		// Plugin may be nil for non-datasource extensions (e.g., feature extensions)
		if plugin == nil {
			slog.Info("Extension loaded without plugin (feature extension)",
				"name", name,
			)
			continue
		}

		// Register the plugin in the shared registry
		shared.PluginsByID[plugin.ID] = plugin
		if plugin.DatasourceClient != nil {
			shared.DatasourceClients.Set(plugin.ID, plugin.DatasourceClient)
		}

		slog.Info("Extension plugin loaded",
			"name", name,
			"pluginID", plugin.ID,
			"pluginName", plugin.Name,
		)
	}

	return nil
}

// RegisterExtensionRoutes registers all extension routes
func (a *Api) registerExtensionRoutes(router *gin.RouterGroup) {
	extensions := extensionRegistry.GetAll()

	// Create extension API group
	extensionAPI := router.Group("/ext")

	for name, extension := range extensions {
		slog.Info("Registering extension routes", "name", name)

		// Register extension routes under /plugins/ext/{extension-name}
		extensionGroup := extensionAPI.Group("/" + name)
		extension.RegisterRoutes(extensionGroup)
	}
}

// RegisterRoutesWithExtensions UpdatedRegisterRoutes includes extension route registration
func (a *Api) RegisterRoutesWithExtensions(routes gin.IRoutes) {
	// Register existing routes
	routes.GET("", pluginsHandler)

	routes.GET("/datasources", datasourcesHandler)
	routes.GET("/datasources/:name", datasourcesHandler)
	routes.POST("/datasources", datasourcesHandler)
	routes.PUT("/datasources/:name", datasourcesHandler)
	routes.DELETE("/datasources/:name", datasourcesHandler)

	routes.POST("/ds/query", dsQueryPostHandler)

	// Register extension routes
	if routerGroup, ok := routes.(*gin.RouterGroup); ok {
		a.registerExtensionRoutes(routerGroup)
	}
}
