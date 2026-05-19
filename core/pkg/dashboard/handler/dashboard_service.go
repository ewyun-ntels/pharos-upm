package handler

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ory/fosite/token/jwt"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/internal/casbin"
	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/factory"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/shared/types/dashboard"
	"ntels.com/pharos/shared/types/role"
)

// DashboardService DashboardService는 dashboard repository를 사용하는 새로운 서비스 구조체입니다.
type DashboardService struct {
	repo        repositories.DashboardRepository
	folderRepo  repositories.DashboardFolderRepository
	historyRepo repositories.DashboardHistoryRepository
	enforcer    *casbin.Enforcer
}

// NewDashboardService creates a new dashboard service with repository pattern
func NewDashboardService(config common.Config, enforcer *casbin.Enforcer) *DashboardService {
	repo, err := factory.NewDashboardRepository(factory.RepositoryOptions{
		DatabaseConfig: config.Database,
	})
	if err != nil {
		slog.Error("Failed to create dashboard repository", "error", err)
		repo = nil
	}

	folderRepo, err := factory.NewDashboardFolderRepository(factory.RepositoryOptions{
		DatabaseConfig: config.Database,
	})
	if err != nil {
		slog.Error("Failed to create dashboard folder repository", "error", err)
		folderRepo = nil
	}

	historyRepo, err := factory.NewDashboardHistoryRepository(factory.RepositoryOptions{
		DatabaseConfig: config.Database,
	})
	if err != nil {
		slog.Error("Failed to create dashboard history repository", "error", err)
		historyRepo = nil
	}

	return &DashboardService{
		repo:        repo,
		folderRepo:  folderRepo,
		historyRepo: historyRepo,
		enforcer:    enforcer,
	}
}

const maxDashboardHistory = 500

func (s *DashboardService) recordHistory(dashboardID, title, action, changedBy string, configJSON *string) {
	if s.historyRepo == nil {
		return
	}
	record := repositories.DashboardHistoryRecord{
		ID:             uuid.New().String(),
		DashboardID:    dashboardID,
		DashboardTitle: title,
		Action:         action,
		ChangedBy:      changedBy,
		ChangedAt:      time.Now(),
		ConfigJSON:     configJSON,
	}
	go func() {
		if err := s.historyRepo.Record(context.Background(), record); err != nil {
			slog.Error("Failed to record dashboard history", "error", err)
			return
		}
		if err := s.historyRepo.Prune(context.Background(), maxDashboardHistory); err != nil {
			slog.Error("Failed to prune dashboard history", "error", err)
		}
	}()
}

// GetHandler handles GET requests for dashboards
func (s *DashboardService) GetHandler(c *gin.Context) (int, any) {
	id := c.Param("id")
	if id != "" {
		return s.getHandler(c, id)
	}
	return s.getsHandler(c)
}

// PostHandler handles POST requests to create a new dashboard
func (s *DashboardService) PostHandler(c *gin.Context) (int, any) {
	jwtClaims := getJWTClaims(c)
	if len(jwtClaims.Subject) == 0 {
		return http.StatusBadRequest, external.ErrorResponse{Message: "empty subject"}
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	dashboardConfig, err := dashboard.UnmarshalDashboardConfig(body)
	if err != nil {
		return http.StatusBadRequest, external.ErrorResponse{Message: err.Error()}
	}

	dashboardEntity := &repositories.DashboardEntity{
		ID:     uuid.New().String(),
		Config: dashboardConfig,
	}

	if err := s.repo.Create(context.Background(), dashboardEntity); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	policy := casbin.Policy{
		jwtClaims.Subject,
		dashboardEntity.ID,
		casbin.ActionOwner,
		PermissionKindUser,
	}
	if err := s.enforcer.AddPolicy(policy); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	s.recordHistory(dashboardEntity.ID, dashboardConfig.Title, "created", jwtClaims.Subject, ptrString(body))
	return http.StatusOK, map[string]string{"id": dashboardEntity.ID}
}

// PutHandler handles PUT requests to update a dashboard
func (s *DashboardService) PutHandler(c *gin.Context) (int, any) {
	id := c.Param("id")
	if id == "" {
		return http.StatusBadRequest, external.ErrorResponse{Message: "id is empty"}
	}

	jwtClaims := getJWTClaims(c)
	check := func(action string) bool { return action == casbin.ActionOwner || action == casbin.ActionEditor }
	if statusCode, _, err := s.checkPermission(id, jwtClaims, check); statusCode == http.StatusForbidden {
		return statusCode, nil
	} else if err != nil {
		return statusCode, external.ErrorResponse{Message: err.Error()}
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	dashboardConfig, err := dashboard.UnmarshalDashboardConfig(body)
	if err != nil {
		return http.StatusBadRequest, external.ErrorResponse{Message: err.Error()}
	}

	// Annotations are stored in the separate `annotations` column; strip from config before saving.
	dashboardConfig.Annotations = nil
	for i := range dashboardConfig.Panels {
		if dashboardConfig.Panels[i].Options != nil {
			delete(dashboardConfig.Panels[i].Options, "annotations")
		}
	}

	dashboardEntity := &repositories.DashboardEntity{
		ID:     id,
		Config: dashboardConfig,
	}

	if err := s.repo.Update(context.Background(), dashboardEntity); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	s.recordHistory(id, dashboardConfig.Title, "updated", jwtClaims.Subject, ptrString(body))
	return http.StatusOK, nil
}

// PutAnnotationsHandler handles PUT /dashboard/:id/annotations — replaces annotation list
func (s *DashboardService) PutAnnotationsHandler(c *gin.Context) (int, any) {
	id := c.Param("id")
	if id == "" {
		return http.StatusBadRequest, external.ErrorResponse{Message: "id is empty"}
	}

	jwtClaims := getJWTClaims(c)
	check := func(action string) bool { return action == casbin.ActionOwner || action == casbin.ActionEditor }
	if statusCode, _, err := s.checkPermission(id, jwtClaims, check); statusCode == http.StatusForbidden {
		return statusCode, nil
	} else if err != nil {
		return statusCode, external.ErrorResponse{Message: err.Error()}
	}

	body, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	var annotations []dashboard.Annotation
	if err := json.Unmarshal(body, &annotations); err != nil {
		return http.StatusBadRequest, external.ErrorResponse{Message: err.Error()}
	}

	if err := s.repo.UpdateAnnotations(context.Background(), id, annotations); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

// DeleteHandler handles DELETE requests to delete a dashboard
func (s *DashboardService) DeleteHandler(c *gin.Context) (int, any) {
	id := c.Param("id")
	if id == "" {
		return http.StatusBadRequest, external.ErrorResponse{Message: "id is empty"}
	}

	jwtClaims := getJWTClaims(c)
	check := func(action string) bool { return action == casbin.ActionOwner }
	if statusCode, _, err := s.checkPermission(id, jwtClaims, check); statusCode == http.StatusForbidden {
		return statusCode, nil
	} else if err != nil {
		return statusCode, external.ErrorResponse{Message: err.Error()}
	}

	if err := s.repo.DeleteByID(context.Background(), id); errors.Is(err, sql.ErrNoRows) {
		return http.StatusOK, nil
	} else if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	if err := s.enforcer.RemovePolicyFromField(1, id); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	if s.historyRepo != nil {
		if err := s.historyRepo.DeleteByDashboardID(context.Background(), id); err != nil {
			slog.Error("Failed to delete dashboard history", "dashboardID", id, "error", err)
		}
	}
	return http.StatusOK, nil
}

// HistoryHandler handles GET /dashboard/history
func (s *DashboardService) HistoryHandler(c *gin.Context) (int, any) {
	if s.historyRepo == nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: "history repository not available"}
	}
	limit := 20
	offset := 0
	if l := c.Query("limit"); l != "" {
		if v, err := strconv.Atoi(l); err == nil && v > 0 {
			limit = v
		}
	}
	if o := c.Query("offset"); o != "" {
		if v, err := strconv.Atoi(o); err == nil && v >= 0 {
			offset = v
		}
	}
	search := c.Query("search")

	jwtClaims := getJWTClaims(c)

	var records []repositories.DashboardHistoryRecord
	var total int

	if common.GetRole(jwtClaims, string(role.RoleSuperAdmin)) {
		var err error
		records, err = s.historyRepo.Query(context.Background(), limit, offset, search)
		if err != nil {
			return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
		}
		total, err = s.historyRepo.Count(context.Background(), search)
		if err != nil {
			return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
		}
	} else {
		allDashboards, err := s.gets()
		if err != nil {
			return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
		}
		groupName := getGroupName(jwtClaims)
		var permittedIDs []string
		for _, d := range allDashboards {
			policies := []casbin.Policy{
				{jwtClaims.Subject, d.ID, casbin.ActionOwner, PermissionKindUser},
				{jwtClaims.Subject, d.ID, casbin.ActionEditor, PermissionKindUser},
				{jwtClaims.Subject, d.ID, casbin.ActionViewer, PermissionKindUser},
				{groupName, d.ID, casbin.ActionOwner, PermissionKindGroup},
				{groupName, d.ID, casbin.ActionEditor, PermissionKindGroup},
				{groupName, d.ID, casbin.ActionViewer, PermissionKindGroup},
				{casbin.SubjectPublic, d.ID, casbin.ActionOwner, PermissionKindUser},
				{casbin.SubjectPublic, d.ID, casbin.ActionEditor, PermissionKindUser},
				{casbin.SubjectPublic, d.ID, casbin.ActionViewer, PermissionKindUser},
			}
			if _, enforceErr := s.enforcer.Enforce(policies...); enforceErr == nil {
				permittedIDs = append(permittedIDs, d.ID)
			} else if !errors.Is(enforceErr, external.ErrorNoSuchPolicy) {
				return http.StatusInternalServerError, external.ErrorResponse{Message: enforceErr.Error()}
			}
		}
		records, err = s.historyRepo.QueryByDashboardIDs(context.Background(), permittedIDs, limit, offset, search)
		if err != nil {
			return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
		}
		total, err = s.historyRepo.CountByDashboardIDs(context.Background(), permittedIDs, search)
		if err != nil {
			return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
		}
	}

	type historyResponse struct {
		ID             string  `json:"id"`
		DashboardID    string  `json:"dashboardId"`
		DashboardTitle string  `json:"dashboardTitle"`
		Action         string  `json:"action"`
		ChangedBy      string  `json:"changedBy"`
		ChangedAt      string  `json:"changedAt"`
		ConfigJSON     *string `json:"configJson,omitempty"`
	}
	items := make([]historyResponse, 0, len(records))
	for _, r := range records {
		items = append(items, historyResponse{
			ID:             r.ID,
			DashboardID:    r.DashboardID,
			DashboardTitle: r.DashboardTitle,
			Action:         r.Action,
			ChangedBy:      r.ChangedBy,
			ChangedAt:      r.ChangedAt.UTC().Format(time.RFC3339),
			ConfigJSON:     r.ConfigJSON,
		})
	}
	return http.StatusOK, map[string]any{"data": items, "total": total}
}

// PatchFolderHandler handles PATCH /:id/folder — moves a dashboard into (or out of) a folder
func (s *DashboardService) PatchFolderHandler(c *gin.Context) (int, any) {
	dashboardID := c.Param("id")
	if dashboardID == "" {
		return http.StatusBadRequest, external.ErrorResponse{Message: "id is empty"}
	}

	var body struct {
		FolderID *string `json:"folderId"`
	}
	if err := c.ShouldBindJSON(&body); err != nil {
		return http.StatusBadRequest, external.ErrorResponse{Message: err.Error()}
	}

	if s.folderRepo == nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: "folder repository not available"}
	}

	if err := s.folderRepo.MoveDashboard(context.Background(), dashboardID, body.FolderID); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

// getHandler retrieves a single dashboard
func (s *DashboardService) getHandler(c *gin.Context, id string) (int, any) {
	check := func(action string) bool { return true }
	if response, err := s.makeGetResponse(id, getJWTClaims(c), check, true); errors.Is(err, external.ErrorNoSuchPolicy) {
		return http.StatusForbidden, nil
	} else if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	} else {
		return http.StatusOK, response
	}
}

func (s *DashboardService) gets() ([]*repositories.DashboardEntity, error) {
	return s.repo.ListAll(context.Background())
}

// getsHandler retrieves all dashboards accessible to the user
func (s *DashboardService) getsHandler(c *gin.Context) (int, any) {
	dashboards, err := s.gets()
	if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	responseBody := map[string]any{}

	for _, d := range dashboards {
		check := func(action string) bool { return true }
		if response, err := s.makeGetResponseFromEntity(d, getJWTClaims(c), check, true); errors.Is(err, external.ErrorNoSuchPolicy) {
			continue
		} else if err != nil {
			return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
		} else {
			responseBody[d.ID] = response
		}
	}

	return http.StatusOK, responseBody
}

// makeGetResponseFromEntity builds a response from an already-loaded DashboardEntity
func (s *DashboardService) makeGetResponseFromEntity(entity *repositories.DashboardEntity, jwtClaims *jwt.JWTClaims, check func(string) bool, hide bool) (map[string]any, error) {
	_, policy, err := s.checkPermission(entity.ID, jwtClaims, check)
	if err != nil {
		return nil, err
	}
	action := policy[2]
	config := entity.Config
	if hide {
		s.hideConfig(&config)
	}
	resp := map[string]any{"permission": action, "config": config}
	if entity.FolderID != nil {
		resp["folderId"] = *entity.FolderID
	}
	if entity.Annotations != nil {
		resp["annotations"] = entity.Annotations
	} else {
		resp["annotations"] = []dashboard.Annotation{}
	}
	return resp, nil
}

// makeGetResponse creates a response with permission, config, and folderId
func (s *DashboardService) makeGetResponse(id string, jwtClaims *jwt.JWTClaims, check func(string) bool, hide bool) (map[string]any, error) {
	_, policy, err := s.checkPermission(id, jwtClaims, check)
	if err != nil {
		return nil, err
	}
	action := policy[2]

	entity, err := s.repo.GetByID(context.Background(), id)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	} else if err != nil {
		return nil, err
	}

	switch action {
	case casbin.ActionOwner, casbin.ActionEditor:
		// full access
	case casbin.ActionViewer:
		if hide {
			s.hideConfig(&entity.Config)
		}
	default:
		return nil, external.ErrorInvalidPermissionType
	}

	resp := map[string]any{"permission": action, "config": &entity.Config}
	if entity.FolderID != nil {
		resp["folderId"] = *entity.FolderID
	}
	if entity.Annotations != nil {
		resp["annotations"] = entity.Annotations
	} else {
		resp["annotations"] = []dashboard.Annotation{}
	}
	return resp, nil
}

// hideConfig hides sensitive information
func (s *DashboardService) hideConfig(config *dashboard.DashboardConfig) {
	for panelIndex := range config.Panels {
		if config.Panels[panelIndex].DataProvider == nil {
			continue
		}

		for chartQueryIndex := range config.Panels[panelIndex].DataProvider.ChartQuery {
			config.Panels[panelIndex].DataProvider.ChartQuery[chartQueryIndex].DatasourceName = ""
			config.Panels[panelIndex].DataProvider.ChartQuery[chartQueryIndex].Query = ""
		}
	}

	for index := range config.Filters {
		config.Filters[index].DatasourceName = nil
		config.Filters[index].Query = nil
	}
}

// checkPermission verifies user permission for the dashboard
func (s *DashboardService) checkPermission(id string, jwtClaims *jwt.JWTClaims, check func(string) bool) (int, casbin.Policy, error) {
	if common.GetRole(jwtClaims, string(role.RoleSuperAdmin)) {
		return http.StatusOK, casbin.Policy{jwtClaims.Subject, id, casbin.ActionOwner, PermissionKindUser}, nil
	}

	var groupName string
	groups := common.GetGroups(jwtClaims)
	if len(groups) > 0 {
		groupName = groups[0]
	}

	policies := []casbin.Policy{
		{jwtClaims.Subject, id, casbin.ActionOwner, PermissionKindUser},
		{jwtClaims.Subject, id, casbin.ActionEditor, PermissionKindUser},
		{jwtClaims.Subject, id, casbin.ActionViewer, PermissionKindUser},

		{groupName, id, casbin.ActionOwner, PermissionKindGroup},
		{groupName, id, casbin.ActionEditor, PermissionKindGroup},
		{groupName, id, casbin.ActionViewer, PermissionKindGroup},

		{casbin.SubjectPublic, id, casbin.ActionOwner, PermissionKindUser},
		{casbin.SubjectPublic, id, casbin.ActionEditor, PermissionKindUser},
		{casbin.SubjectPublic, id, casbin.ActionViewer, PermissionKindUser},
	}

	if policy, err := s.enforcer.Enforce(policies...); errors.Is(err, external.ErrorNoSuchPolicy) {
		return http.StatusForbidden, nil, err
	} else if err != nil {
		return http.StatusInternalServerError, nil, err
	} else if !check(policy[2]) {
		return http.StatusForbidden, policy, external.ErrorNoSuchPermission
	} else {
		return http.StatusOK, policy, nil
	}
}

func ptrString(b []byte) *string {
	s := string(b)
	return &s
}
