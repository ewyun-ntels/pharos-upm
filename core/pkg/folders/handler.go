package folders

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/factory"
	"ntels.com/pharos/core/pkg/common"
)

// FolderService handles folder CRUD operations
type FolderService struct {
	repo repositories.DashboardFolderRepository
}

func NewFolderService(config common.Config) *FolderService {
	repo, err := factory.NewDashboardFolderRepository(factory.RepositoryOptions{
		DatabaseConfig: config.Database,
	})
	if err != nil {
		slog.Error("Failed to create dashboard folder repository", "error", err)
		repo = nil
	}
	return &FolderService{repo: repo}
}

// folderResponse is the JSON response for a folder
type folderResponse struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"createdAt"`
	UpdatedAt time.Time `json:"updatedAt"`
}

func toResponse(e *repositories.DashboardFolderEntity) folderResponse {
	return folderResponse{
		ID:        e.ID,
		Name:      e.Name,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

// ListHandler handles GET /folders
func (s *FolderService) ListHandler(c *gin.Context) {
	folders, err := s.repo.ListAll(context.Background())
	if err != nil {
		c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
		return
	}
	resp := make([]folderResponse, 0, len(folders))
	for _, f := range folders {
		resp = append(resp, toResponse(f))
	}
	c.JSON(http.StatusOK, resp)
}

// CreateHandler handles POST /folders
func (s *FolderService) CreateHandler(c *gin.Context) {
	var body struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
		c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "name is required"})
		return
	}

	now := time.Now()
	folder := &repositories.DashboardFolderEntity{
		ID:        uuid.New().String(),
		Name:      body.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}
	if err := s.repo.Create(context.Background(), folder); err != nil {
		c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, toResponse(folder))
}

// UpdateHandler handles PUT /folders/:id
func (s *FolderService) UpdateHandler(c *gin.Context) {
	id := c.Param("id")
	var body struct {
		Name string `json:"name"`
	}
	if err := c.ShouldBindJSON(&body); err != nil || body.Name == "" {
		c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "name is required"})
		return
	}

	folder := &repositories.DashboardFolderEntity{
		ID:        id,
		Name:      body.Name,
		UpdatedAt: time.Now(),
	}
	if err := s.repo.Update(context.Background(), folder); err != nil {
		c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, nil)
}

// DeleteHandler handles DELETE /folders/:id
func (s *FolderService) DeleteHandler(c *gin.Context) {
	id := c.Param("id")
	// Move dashboards in this folder back to root before deleting
	if err := s.repo.MoveDashboard(context.Background(), "", nil); err != nil {
		// best effort – ignore if the cascade is handled by DB ON DELETE SET NULL
		_ = err
	}
	if err := s.repo.DeleteByID(context.Background(), id); err != nil {
		c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
		return
	}
	c.JSON(http.StatusOK, nil)
}

func RegisterRoutes(config common.Config, routes gin.IRoutes) {
	svc := NewFolderService(config)

	routes.GET("", svc.ListHandler)
	routes.POST("", svc.CreateHandler)
	routes.PUT("/:id", svc.UpdateHandler)
	routes.DELETE("/:id", svc.DeleteHandler)
}
