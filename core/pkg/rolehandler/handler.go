package rolehandler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite/token/jwt"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins"
	coreRole "ntels.com/pharos/core/pkg/role"
	"ntels.com/pharos/core/pkg/rolehandler/adapters"
	sharedRole "ntels.com/pharos/shared/types/role"
)

// Handler - Role API handler
type Handler struct {
	metadataService *coreRole.MetadataService
	adapter         *adapters.APIAdapter
}

// NewHandler - Role handler 생성
func NewHandler(metadataService *coreRole.MetadataService) *Handler {
	return &Handler{
		metadataService: metadataService,
		adapter:         &adapters.APIAdapter{},
	}
}

// collectBase returns the full flat atomic role list: core roles + all extension roles.
// Extension roles have their Group set to the extension's registry name so they
// form their own named group in the UI without any extra bookkeeping.
func collectBase() []sharedRole.RoleMetadata {
	base := coreRole.GetCoreRoleMetadata()
	for name, ext := range plugins.GetExtensionRegistry().GetAll() {
		for _, r := range ext.GetRoles() {
			r.Group = name
			base = append(base, r)
		}
	}
	return base
}

// groupByField groups a flat role list by each role's Group field.
// Unknown groups use their key as display name.
func groupByField(roles []sharedRole.RoleMetadata) []sharedRole.RoleGroup {
	groupDisplayNames := map[string]string{
		"role":      "Roles",
		"attribute": "User Attributes",
		"composite": "Composite Roles",
		"custom":    "Custom Roles",
	}
	groupMap := make(map[string][]sharedRole.RoleMetadata)
	for _, r := range roles {
		g := r.Group
		if g == "" {
			g = "custom"
		}
		groupMap[g] = append(groupMap[g], r)
	}
	result := make([]sharedRole.RoleGroup, 0, len(groupMap))
	for g, rs := range groupMap {
		dn := groupDisplayNames[g]
		if dn == "" {
			dn = g
		}
		result = append(result, sharedRole.RoleGroup{Group: g, DisplayName: dn, Roles: rs})
	}
	return result
}

// GetBaseRoles - GET /api/roles/base
// Raw base roles without overlay (core + extension) — for admin config page comparison.
func (h *Handler) GetBaseRoles(c *gin.Context) {
	c.JSON(http.StatusOK, groupByField(collectBase()))
}

// GetAllRoles - GET /api/roles
// Step 1: flat base = core roles + extension roles
// Step 2: apply DB overlay once (patch existing + append composite)
// Step 3: filter hidden (unless includeHidden=true)
// Step 4: group by Group field → response
func (h *Handler) GetAllRoles(c *gin.Context) {
	includeHidden := c.Query("includeHidden") == "true"

	// Step 1: flat base
	base := collectBase()

	// Step 2: overlay
	merged := base
	if h.metadataService != nil {
		var err error
		merged, err = h.metadataService.ApplyOverlay(c.Request.Context(), base)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to apply role overlay", "details": err.Error()})
			return
		}
	}

	// Step 3: filter hidden
	if !includeHidden {
		visible := make([]sharedRole.RoleMetadata, 0, len(merged))
		for _, r := range merged {
			if r.Hide == nil || !*r.Hide {
				visible = append(visible, r)
			}
		}
		merged = visible
	}

	// Step 4: group and respond
	c.JSON(http.StatusOK, groupByField(merged))
}

// GetRoleMetadataConfig - GET /api/roles/metadata/config
// Get active configuration overlay from DB
func (h *Handler) GetRoleMetadataConfig(c *gin.Context) {
	if h.metadataService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "metadata service not available"})
		return
	}

	config, err := h.metadataService.GetActiveConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get configuration", "details": err.Error()})
		return
	}

	if config == nil {
		c.JSON(http.StatusOK, sharedRole.RoleMetadataConfigResponse{
			Config:  []sharedRole.RoleMetadata{},
			Message: ptr("no configuration overlay (using base only)"),
		})
		return
	}

	c.JSON(http.StatusOK, sharedRole.RoleMetadataConfigResponse{
		Config: config,
	})
}

func ptr[T any](v T) *T {
	return &v
}

// queryLimit parses the "limit" query parameter (1–100, default 10).
func queryLimit(c *gin.Context) int {
	limit := 10
	if s := c.Query("limit"); s != "" {
		var n int
		if _, err := fmt.Sscanf(s, "%d", &n); err == nil && n > 0 && n <= 100 {
			limit = n
		}
	}
	return limit
}

// extractUpdatedBy pulls the JWT subject (username) from the Gin context for audit logging.
func extractUpdatedBy(c *gin.Context) *string {
	if value, exists := c.Get(common.ContextKeyJWTClaims); exists {
		if claims, ok := value.(*jwt.JWTClaims); ok {
			if sub := claims.Subject; sub != "" {
				return &sub
			}
		}
	}
	return nil
}

// GetRoleMetadataConfigHistory - GET /api/roles/metadata/config/history
// Get configuration history
func (h *Handler) GetRoleMetadataConfigHistory(c *gin.Context) {
	if h.metadataService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "metadata service not available"})
		return
	}

	limit := queryLimit(c)

	history, err := h.metadataService.GetHistory(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get history", "details": err.Error()})
		return
	}

	// Convert internal entities to shared response types
	apiHistory := make([]sharedRole.RoleMetadataConfigEntity, 0, len(history))
	for _, h := range history {
		apiHistory = append(apiHistory, sharedRole.RoleMetadataConfigEntity{
			ID:          h.ID,
			Config:      h.Config,
			CreatedAt:   h.CreatedAt.Format(time.RFC3339),
			UpdatedBy:   h.UpdatedBy,
			Description: h.Description,
		})
	}

	c.JSON(http.StatusOK, sharedRole.RoleMetadataConfigHistoryResponse{
		History: apiHistory,
	})
}

// SaveRoleMetadataConfig - PUT /api/roles/metadata/config
// Save entire configuration overlay (replaces all)
func (h *Handler) SaveRoleMetadataConfig(c *gin.Context) {
	if h.metadataService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "metadata service not available"})
		return
	}

	var req struct {
		Config      *[]sharedRole.RoleMetadata `json:"config"`
		Description *string                    `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	if req.Config == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "missing required field: config"})
		return
	}

	// Validate configuration before saving
	if err := coreRole.ValidateConfig(*req.Config); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "configuration validation failed", "details": err.Error()})
		return
	}

	updatedBy := extractUpdatedBy(c)

	if err := h.metadataService.SaveConfig(c.Request.Context(), *req.Config, updatedBy, req.Description); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save configuration", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "configuration saved successfully", "count": len(*req.Config)})
}

// RollbackRoleMetadataConfig - POST /api/roles/metadata/config/rollback
// Rollback to a specific configuration version by id
func (h *Handler) RollbackRoleMetadataConfig(c *gin.Context) {
	if h.metadataService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "metadata service not available"})
		return
	}

	var req struct {
		ID string `json:"id" binding:"required"` // History entry UUID
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	updatedBy := extractUpdatedBy(c)

	if err := h.metadataService.Rollback(c.Request.Context(), req.ID, updatedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rollback", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rollback successful", "id": req.ID})
}

// DeleteRoleMetadataConfig - DELETE /api/roles/metadata/config
// Remove all configuration overlay (revert to base)
func (h *Handler) DeleteRoleMetadataConfig(c *gin.Context) {
	if h.metadataService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "metadata service not available"})
		return
	}

	if err := h.metadataService.DeleteConfig(c.Request.Context()); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete configuration", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "configuration deleted successfully (reverted to base)"})
}
