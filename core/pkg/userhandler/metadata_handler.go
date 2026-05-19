package userhandler

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite/token/jwt"
	"ntels.com/pharos/core/pkg/common"
	sharedUser "ntels.com/pharos/shared/types/user"
)

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

// Metadata handlers - similar to rolehandler pattern

// GetUserMetadataConfig - GET /api/users/metadata/config
// Get active user metadata configuration from DB
func (h *handler) GetUserMetadataConfig(c *gin.Context) {
	if h.userMetadataService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "user metadata service not available"})
		return
	}

	config, err := h.userMetadataService.GetActiveConfig(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get configuration", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": config})
}

// GetUserMetadataConfigHistory - GET /api/users/metadata/config/history
// Get configuration history
func (h *handler) GetUserMetadataConfigHistory(c *gin.Context) {
	if h.userMetadataService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "user metadata service not available"})
		return
	}

	limit := queryLimit(c)

	history, err := h.userMetadataService.GetHistory(c.Request.Context(), limit)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to get history", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"history": history})
}

// SaveUserMetadataConfig - PUT /api/users/metadata/config
// Save entire configuration overlay (replaces all)
func (h *handler) SaveUserMetadataConfig(c *gin.Context) {
	if h.userMetadataService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "user metadata service not available"})
		return
	}

	var req struct {
		Config      sharedUser.UserMetadataConfig `json:"config" binding:"required"`
		Description *string                       `json:"description"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request body", "details": err.Error()})
		return
	}

	updatedBy := extractUpdatedBy(c)

	if err := h.userMetadataService.SaveConfig(c.Request.Context(), &req.Config, updatedBy, req.Description); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to save configuration", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "configuration saved successfully"})
}

// RollbackUserMetadataConfig - POST /api/users/metadata/config/rollback
// Rollback to a specific configuration version by id
func (h *handler) RollbackUserMetadataConfig(c *gin.Context) {
	if h.userMetadataService == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "user metadata service not available"})
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

	if err := h.userMetadataService.Rollback(c.Request.Context(), req.ID, updatedBy); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to rollback", "details": err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "rollback successful"})
}
