package rolehandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/pkg/plugins"
	pluginModel "ntels.com/pharos/core/pkg/plugins/model"
	"ntels.com/pharos/core/pkg/role"
	sharedRole "ntels.com/pharos/shared/types/role"
)

// Mock repository for testing
type mockConfigRepo struct {
	activeConfig    *repositories.RoleMetadataConfigEntity
	history         []repositories.RoleMetadataConfigEntity
	getActiveErr    error
	getHistoryErr   error
	saveErr         error
	deleteAllErr    error
	rollbackErr     error
	saveCalled      bool
	deleteAllCalled bool
	rollbackCalled  bool
	lastSavedConfig []sharedRole.RoleMetadata
	lastRollbackID  string
}

func (m *mockConfigRepo) GetActiveConfig(_ context.Context) (*repositories.RoleMetadataConfigEntity, error) {
	if m.getActiveErr != nil {
		return nil, m.getActiveErr
	}
	return m.activeConfig, nil
}

func (m *mockConfigRepo) GetHistory(_ context.Context, limit int) ([]repositories.RoleMetadataConfigEntity, error) {
	if m.getHistoryErr != nil {
		return nil, m.getHistoryErr
	}
	if limit > 0 && limit < len(m.history) {
		return m.history[:limit], nil
	}
	return m.history, nil
}

func (m *mockConfigRepo) SaveConfig(_ context.Context, config repositories.RoleMetadataConfigEntity, _ int) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saveCalled = true
	m.lastSavedConfig = config.Config
	return nil
}

func (m *mockConfigRepo) DeleteAll(_ context.Context) error {
	if m.deleteAllErr != nil {
		return m.deleteAllErr
	}
	m.deleteAllCalled = true
	return nil
}

func (m *mockConfigRepo) Rollback(_ context.Context, id string, _ *string, _ int) error {
	if m.rollbackErr != nil {
		return m.rollbackErr
	}
	m.rollbackCalled = true
	m.lastRollbackID = id
	return nil
}

func setupTestRouter(handler *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()
	router.GET("/api/roles", handler.GetAllRoles)
	router.GET("/api/roles/base", handler.GetBaseRoles)
	router.GET("/api/roles/metadata/config", handler.GetRoleMetadataConfig)
	router.PUT("/api/roles/metadata/config", handler.SaveRoleMetadataConfig)
	router.DELETE("/api/roles/metadata/config", handler.DeleteRoleMetadataConfig)
	router.GET("/api/roles/metadata/config/history", handler.GetRoleMetadataConfigHistory)
	router.POST("/api/roles/metadata/config/rollback", handler.RollbackRoleMetadataConfig)
	return router
}

func TestGetAllRoles(t *testing.T) {
	t.Run("without metadata service", func(t *testing.T) {
		handler := NewHandler(nil)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/roles", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var result []sharedRole.RoleGroup
		err := json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.NotEmpty(t, result) // Should have core roles at least
	})

	t.Run("with metadata service and overlay", func(t *testing.T) {
		hide := true
		overlay := []sharedRole.RoleMetadata{
			{
				Key:         "role:custom_role",
				DisplayName: "Custom Role",
				Description: "Test custom role",
				Group:       "custom",
			},
			{
				Key:         "role:super_admin",
				DisplayName: "Super Admin (Modified)",
				Description: "Modified description",
				Group:       "role",
				Hide:        &hide, // This should be filtered out
			},
		}
		mockRepo := &mockConfigRepo{
			activeConfig: &repositories.RoleMetadataConfigEntity{
				Config: overlay,
			},
		}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/roles", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var result []sharedRole.RoleGroup
		err := json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)

		// Find custom role
		foundCustom := false
		foundHidden := false
		for _, group := range result {
			for _, groupRole := range group.Roles {
				if groupRole.Key == "role:custom_role" {
					foundCustom = true
					assert.Equal(t, "Custom Role", groupRole.DisplayName)
				}
				if groupRole.Key == "role:super_admin" && groupRole.Hide != nil && *groupRole.Hide {
					foundHidden = true
				}
			}
		}
		assert.True(t, foundCustom, "Should include custom role from overlay")
		assert.False(t, foundHidden, "Should not include hidden roles")
	})

	t.Run("overlay error", func(t *testing.T) {
		mockRepo := &mockConfigRepo{
			getActiveErr: errors.New("db error"),
		}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/roles", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})

	t.Run("overlay patches extension role", func(t *testing.T) {
		overlay := []sharedRole.RoleMetadata{
			{Key: "test:ext_role", DisplayName: "Extension Role (patched)", Group: testExtName},
		}
		mockRepo := &mockConfigRepo{
			activeConfig: &repositories.RoleMetadataConfigEntity{Config: overlay},
		}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/roles", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		var result []sharedRole.RoleGroup
		err := json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)

		for _, g := range result {
			for _, r := range g.Roles {
				if r.Key == "test:ext_role" {
					assert.Equal(t, "Extension Role (patched)", r.DisplayName)
					return
				}
			}
		}
		t.Error("test:ext_role not found in GetAllRoles response")
	})
}

func TestGetRoleMetadataConfig(t *testing.T) {
	t.Run("no metadata service", func(t *testing.T) {
		handler := NewHandler(nil)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/roles/metadata/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})

	t.Run("no configuration", func(t *testing.T) {
		mockRepo := &mockConfigRepo{
			activeConfig: nil,
		}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/roles/metadata/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var result map[string]any
		err := json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result, "message")
	})

	t.Run("with configuration", func(t *testing.T) {
		config := []sharedRole.RoleMetadata{
			{
				Key:         "role:test",
				DisplayName: "Test Role",
				Description: "Test",
				Group:       "test",
			},
		}
		mockRepo := &mockConfigRepo{
			activeConfig: &repositories.RoleMetadataConfigEntity{
				Config: config,
			},
		}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/roles/metadata/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var result map[string]any
		err := json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result, "config")
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := &mockConfigRepo{
			getActiveErr: errors.New("db error"),
		}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/roles/metadata/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestSaveRoleMetadataConfig(t *testing.T) {
	t.Run("no metadata service", func(t *testing.T) {
		handler := NewHandler(nil)
		router := setupTestRouter(handler)

		body, _ := json.Marshal(map[string]any{
			"config": []sharedRole.RoleMetadata{},
		})
		req := httptest.NewRequest(http.MethodPut, "/api/roles/metadata/config", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockRepo := &mockConfigRepo{}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodPut, "/api/roles/metadata/config", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("missing config field", func(t *testing.T) {
		mockRepo := &mockConfigRepo{}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		body, _ := json.Marshal(map[string]any{})
		req := httptest.NewRequest(http.MethodPut, "/api/roles/metadata/config", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("invalid configuration", func(t *testing.T) {
		mockRepo := &mockConfigRepo{}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		// Missing required field (key)
		body, _ := json.Marshal(map[string]any{
			"config": []map[string]any{
				{
					"displayName": "Test",
				},
			},
		})
		req := httptest.NewRequest(http.MethodPut, "/api/roles/metadata/config", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockConfigRepo{}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		config := []sharedRole.RoleMetadata{
			{
				Key:         "role:test",
				DisplayName: "Test Role",
				Description: "Test",
				Group:       "test",
			},
		}
		body, _ := json.Marshal(map[string]any{
			"config":      config,
			"description": "Test save",
		})
		req := httptest.NewRequest(http.MethodPut, "/api/roles/metadata/config", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, mockRepo.saveCalled)
		assert.Len(t, mockRepo.lastSavedConfig, 1)
		assert.Equal(t, "role:test", mockRepo.lastSavedConfig[0].Key)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := &mockConfigRepo{
			saveErr: errors.New("db error"),
		}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		config := []sharedRole.RoleMetadata{
			{
				Key:         "role:test",
				DisplayName: "Test",
				Group:       "test",
			},
		}
		body, _ := json.Marshal(map[string]any{
			"config": config,
		})
		req := httptest.NewRequest(http.MethodPut, "/api/roles/metadata/config", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestDeleteRoleMetadataConfig(t *testing.T) {
	t.Run("no metadata service", func(t *testing.T) {
		handler := NewHandler(nil)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodDelete, "/api/roles/metadata/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockConfigRepo{}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodDelete, "/api/roles/metadata/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, mockRepo.deleteAllCalled)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := &mockConfigRepo{
			deleteAllErr: errors.New("db error"),
		}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodDelete, "/api/roles/metadata/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestGetRoleMetadataConfigHistory(t *testing.T) {
	t.Run("no metadata service", func(t *testing.T) {
		handler := NewHandler(nil)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/roles/metadata/config/history", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})

	t.Run("default limit", func(t *testing.T) {
		history := make([]repositories.RoleMetadataConfigEntity, 20)
		for i := range 20 {
			history[i] = repositories.RoleMetadataConfigEntity{
				ID: string(rune('A' + i)),
			}
		}
		mockRepo := &mockConfigRepo{
			history: history,
		}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/roles/metadata/config/history", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var result map[string]any
		err := json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result, "history")
	})

	t.Run("custom limit", func(t *testing.T) {
		history := make([]repositories.RoleMetadataConfigEntity, 20)
		mockRepo := &mockConfigRepo{
			history: history,
		}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/roles/metadata/config/history?limit=5", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := &mockConfigRepo{
			getHistoryErr: errors.New("db error"),
		}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/roles/metadata/config/history", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestRollbackRoleMetadataConfig(t *testing.T) {
	t.Run("no metadata service", func(t *testing.T) {
		handler := NewHandler(nil)
		router := setupTestRouter(handler)

		body, _ := json.Marshal(map[string]string{"id": "test-id"})
		req := httptest.NewRequest(http.MethodPost, "/api/roles/metadata/config/rollback", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})

	t.Run("invalid JSON", func(t *testing.T) {
		mockRepo := &mockConfigRepo{}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		req := httptest.NewRequest(http.MethodPost, "/api/roles/metadata/config/rollback", bytes.NewBufferString("{invalid"))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("missing id", func(t *testing.T) {
		mockRepo := &mockConfigRepo{}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		body, _ := json.Marshal(map[string]string{})
		req := httptest.NewRequest(http.MethodPost, "/api/roles/metadata/config/rollback", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("success", func(t *testing.T) {
		mockRepo := &mockConfigRepo{}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		body, _ := json.Marshal(map[string]string{"id": "test-id-123"})
		req := httptest.NewRequest(http.MethodPost, "/api/roles/metadata/config/rollback", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, mockRepo.rollbackCalled)
		assert.Equal(t, "test-id-123", mockRepo.lastRollbackID)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := &mockConfigRepo{
			rollbackErr: errors.New("db error"),
		}
		metadataService := role.NewMetadataService(mockRepo)
		handler := NewHandler(metadataService)
		router := setupTestRouter(handler)

		body, _ := json.Marshal(map[string]string{"id": "test-id"})
		req := httptest.NewRequest(http.MethodPost, "/api/roles/metadata/config/rollback", bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

// mockExtension implements plugins.Extension for testing.
type mockExtension struct {
	name  string
	roles []sharedRole.RoleMetadata
}

func (m *mockExtension) GetPlugin() (*pluginModel.Plugin, error) { return nil, nil }
func (m *mockExtension) RegisterRoutes(_ *gin.RouterGroup)       {}
func (m *mockExtension) GetName() string                         { return m.name }
func (m *mockExtension) GetVersion() string                      { return "0.0.1-test" }
func (m *mockExtension) GetType() string                         { return "feature" }
func (m *mockExtension) GetRoles() []sharedRole.RoleMetadata     { return m.roles }

// testExtName is the registry key used for the mock extension registered in TestMain.
const testExtName = "test-ext"

// TestMain registers a stub extension so that collectBase, GetBaseRoles, and overlay tests
// can verify extension-role behaviour without mocking the global registry.
func TestMain(m *testing.M) {
	gin.SetMode(gin.TestMode)
	plugins.RegisterExtension(&mockExtension{
		name: testExtName,
		roles: []sharedRole.RoleMetadata{
			{Key: "test:ext_role", DisplayName: "Extension Role", Description: "Provided by test extension"},
		},
	})
	os.Exit(m.Run())
}

func TestGroupByField(t *testing.T) {
	t.Run("groups by Group field", func(t *testing.T) {
		roles := []sharedRole.RoleMetadata{
			{Key: "role:a", DisplayName: "A", Group: "role"},
			{Key: "role:b", DisplayName: "B", Group: "role"},
			{Key: "attr:x", DisplayName: "X", Group: "attribute"},
		}
		result := groupByField(roles)
		assert.Len(t, result, 2)
		counts := map[string]int{}
		for _, g := range result {
			counts[g.Group] = len(g.Roles)
		}
		assert.Equal(t, 2, counts["role"])
		assert.Equal(t, 1, counts["attribute"])
	})

	t.Run("empty group defaults to custom", func(t *testing.T) {
		roles := []sharedRole.RoleMetadata{
			{Key: "x:orphan", DisplayName: "Orphan"},
		}
		result := groupByField(roles)
		require.Len(t, result, 1)
		assert.Equal(t, "custom", result[0].Group)
	})

	t.Run("unknown group key uses key as display name", func(t *testing.T) {
		roles := []sharedRole.RoleMetadata{
			{Key: "ext:role1", DisplayName: "Role1", Group: "my-extension"},
		}
		result := groupByField(roles)
		require.Len(t, result, 1)
		assert.Equal(t, "my-extension", result[0].DisplayName)
	})

	t.Run("known group key uses mapped display name", func(t *testing.T) {
		roles := []sharedRole.RoleMetadata{
			{Key: "attr:x", DisplayName: "X", Group: "attribute"},
		}
		result := groupByField(roles)
		require.Len(t, result, 1)
		assert.Equal(t, "User Attributes", result[0].DisplayName)
	})
}

func TestCollectBase(t *testing.T) {
	base := collectBase()

	t.Run("includes core roles", func(t *testing.T) {
		assert.NotEmpty(t, base, "collectBase must return at least core roles")
	})

	t.Run("includes extension role with registry name as Group", func(t *testing.T) {
		found := false
		for _, r := range base {
			if r.Key == "test:ext_role" {
				found = true
				assert.Equal(t, testExtName, r.Group, "extension role Group must equal the registry name")
			}
		}
		assert.True(t, found, "extension role 'test:ext_role' must appear in collectBase output")
	})

	t.Run("no duplicate keys", func(t *testing.T) {
		seen := map[string]int{}
		for _, r := range base {
			seen[r.Key]++
		}
		for k, count := range seen {
			assert.Equal(t, 1, count, "key %q appears %d times", k, count)
		}
	})
}

func TestGetBaseRoles(t *testing.T) {
	handler := NewHandler(nil)
	router := setupTestRouter(handler)

	req := httptest.NewRequest(http.MethodGet, "/api/roles/base", nil)
	rec := httptest.NewRecorder()
	router.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)
	var result []sharedRole.RoleGroup
	err := json.Unmarshal(rec.Body.Bytes(), &result)
	require.NoError(t, err)

	foundGroup := false
	foundRole := false
	for _, g := range result {
		if g.Group == testExtName {
			foundGroup = true
			for _, r := range g.Roles {
				if r.Key == "test:ext_role" {
					foundRole = true
				}
			}
		}
	}
	assert.True(t, foundGroup, "extension group %q must appear in /api/roles/base response", testExtName)
	assert.True(t, foundRole, "extension role test:ext_role must appear in /api/roles/base response")
}
