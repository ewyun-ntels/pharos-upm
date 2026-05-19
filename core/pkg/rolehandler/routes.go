package rolehandler

import (
	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/authhandler"
	coreRole "ntels.com/pharos/core/pkg/role"
)

// RegisterRoutes registers role-related routes into the given router group.
// Requires authentication to access role configuration.
func RegisterRoutes(router gin.IRouter, metadataService *coreRole.MetadataService) {
	handler := NewHandler(metadataService)

	// /api/roles - 통합 role metadata API
	router.GET("/api/roles",
		authhandler.GetAuthenticationHandler(true),
		handler.GetAllRoles)

	// /api/roles/base - raw base roles without overlay (for admin config comparison)
	router.GET("/api/roles/base",
		authhandler.GetAuthenticationHandler(true),
		handler.GetBaseRoles)

	// /api/roles/metadata - metadata configuration management (Kustomize overlay)
	metadataGroup := router.Group("/api/roles/metadata")
	metadataGroup.Use(authhandler.GetAuthenticationHandler(true))
	{
		// GET /api/roles/metadata/config - get active configuration overlay
		metadataGroup.GET("/config", handler.GetRoleMetadataConfig)

		// PUT /api/roles/metadata/config - save entire configuration overlay
		metadataGroup.PUT("/config", handler.SaveRoleMetadataConfig)

		// DELETE /api/roles/metadata/config - delete configuration overlay (revert to base)
		metadataGroup.DELETE("/config", handler.DeleteRoleMetadataConfig)

		// GET /api/roles/metadata/config/history - get configuration history
		metadataGroup.GET("/config/history", handler.GetRoleMetadataConfigHistory)

		// POST /api/roles/metadata/config/rollback - rollback to specific version
		metadataGroup.POST("/config/rollback", handler.RollbackRoleMetadataConfig)
	}
}
