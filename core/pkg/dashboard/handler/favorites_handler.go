package handler

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/internal/casbin"
)

// FavoriteItem represents a favorite dashboard item (id + displayName only)
type FavoriteItem struct {
	ID          string `json:"id"`
	DisplayName string `json:"displayName"`
}

// GetFavoritesHandler handles GET /dashboard/favorites
// Returns the list of dashboards marked as favorite that the user has access to.
func (s *DashboardService) GetFavoritesHandler(c *gin.Context) (int, any) {
	dashboards, err := s.gets()
	if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	jwtClaims := getJWTClaims(c)
	var result []FavoriteItem

	for _, d := range dashboards {
		if !d.Config.Favorite {
			continue
		}

		check := func(action string) bool { return true }
		statusCode, _, err := s.checkPermission(d.ID, jwtClaims, check)
		if statusCode == http.StatusForbidden {
			continue
		} else if err != nil {
			return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
		}

		displayName := d.Config.Title
		if d.Config.DisplayName != "" {
			displayName = d.Config.DisplayName
		}
		result = append(result, FavoriteItem{ID: d.ID, DisplayName: displayName})
	}

	return http.StatusOK, result
}

// PatchFavoriteHandler handles PATCH /dashboard/:id/favorite
// Toggles the favorite field (true ↔ false) for a given dashboard.
func (s *DashboardService) PatchFavoriteHandler(c *gin.Context) (int, any) {
	id := c.Param("id")
	if id == "" {
		return http.StatusBadRequest, external.ErrorResponse{Message: "id is empty"}
	}

	jwtClaims := getJWTClaims(c)
	check := func(action string) bool {
		return action == casbin.ActionOwner || action == casbin.ActionEditor
	}

	statusCode, _, err := s.checkPermission(id, jwtClaims, check)
	if statusCode == http.StatusForbidden {
		return http.StatusForbidden, nil
	} else if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	dashboardEntity, err := s.repo.GetByID(context.Background(), id)
	if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}
	if dashboardEntity == nil {
		return http.StatusNotFound, nil
	}

	dashboardEntity.Config.Favorite = !dashboardEntity.Config.Favorite

	if err := s.repo.Update(context.Background(), dashboardEntity); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, map[string]bool{"favorite": dashboardEntity.Config.Favorite}
}

func getFavoritesHandler(c *gin.Context) {
	c.JSON(svc.GetFavoritesHandler(c))
}

func patchFavoriteHandler(c *gin.Context) {
	c.JSON(svc.PatchFavoriteHandler(c))
}
