package userhandler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/user"
	coreUser "ntels.com/pharos/core/pkg/user"
	sharedUser "ntels.com/pharos/shared/types/user"
)

// Mock repository for testing
type mockUserMetadataConfigRepo struct {
	activeConfig    *repositories.UserMetadataConfigEntity
	history         []repositories.UserMetadataConfigEntity
	getActiveErr    error
	getHistoryErr   error
	saveErr         error
	getByIDErr      error
	saveCalled      bool
	lastSavedConfig *sharedUser.UserMetadataConfig
}

func (m *mockUserMetadataConfigRepo) GetActiveConfig(_ context.Context) (*repositories.UserMetadataConfigEntity, error) {
	if m.getActiveErr != nil {
		return nil, m.getActiveErr
	}
	return m.activeConfig, nil
}

func (m *mockUserMetadataConfigRepo) GetHistory(_ context.Context, limit int) ([]repositories.UserMetadataConfigEntity, error) {
	if m.getHistoryErr != nil {
		return nil, m.getHistoryErr
	}
	if limit > 0 && limit < len(m.history) {
		return m.history[:limit], nil
	}
	return m.history, nil
}

func (m *mockUserMetadataConfigRepo) SaveConfig(_ context.Context, config repositories.UserMetadataConfigEntity, _ int) error {
	if m.saveErr != nil {
		return m.saveErr
	}
	m.saveCalled = true
	// Parse ConfigJSON to Config for verification
	var parsedConfig sharedUser.UserMetadataConfig
	if err := json.Unmarshal([]byte(config.ConfigJSON), &parsedConfig); err == nil {
		m.lastSavedConfig = &parsedConfig
	}
	return nil
}

func (m *mockUserMetadataConfigRepo) GetByID(_ context.Context, id string) (*repositories.UserMetadataConfigEntity, error) {
	if m.getByIDErr != nil {
		return nil, m.getByIDErr
	}
	for _, h := range m.history {
		if h.ID == id {
			return &h, nil
		}
	}
	return nil, errors.New("not found")
}

// Mock user store for testing
type mockUserStore struct{}

func (m *mockUserStore) ListUsers() ([]user.User, error)            { return nil, nil }
func (m *mockUserStore) FindByUsername(_ string) (user.User, error) { return nil, nil }
func (m *mockUserStore) CreateUser(_ string, _ string, _ map[string]any, _ *[]string) error {
	return nil
}
func (m *mockUserStore) DeleteUser(_ string) error                      { return nil }
func (m *mockUserStore) ChangePassword(_, _ string) error               { return nil }
func (m *mockUserStore) SetAttributes(_ string, _ map[string]any) error { return nil }
func (m *mockUserStore) DecodeExtra(_ map[string]any) (map[string]any, error) {
	return nil, nil
}
func (m *mockUserStore) SetBlock(_ string, _ bool) error                       { return nil }
func (m *mockUserStore) Authenticate(_, _ string) error                        { return nil }
func (m *mockUserStore) InsertPasswordHistory(_, _ string) error               { return nil }
func (m *mockUserStore) DeletePasswordHistory(_ string, _ int32) error         { return nil }
func (m *mockUserStore) CheckPasswordHistory(_, _ string) error                { return nil }
func (m *mockUserStore) UpdatePasswordExpiration(_ string, _ *time.Time) error { return nil }
func (m *mockUserStore) UpdatePrepare(_ string, _ []string) error              { return nil }
func (m *mockUserStore) Close() error                                          { return nil }
func (m *mockUserStore) ValidatePasswordComplexity(_ string) error             { return nil }
func (m *mockUserStore) GetLoginRetryLimit() int                               { return 0 }
func (m *mockUserStore) GetLoginBlockDuration() time.Duration                  { return 0 }
func (m *mockUserStore) GetPrepare(_ user.User) ([]string, error)              { return nil, nil }

func setupTestRouterForMetadata(handler Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Add simple middleware to set username in context
	router.Use(func(c *gin.Context) {
		c.Set("username", "testuser")
		c.Next()
	})

	router.GET("/api/users/metadata/config", handler.GetUserMetadataConfig)
	router.GET("/api/users/metadata/config/history", handler.GetUserMetadataConfigHistory)
	router.PUT("/api/users/metadata/config", handler.SaveUserMetadataConfig)
	router.POST("/api/users/metadata/config/rollback", handler.RollbackUserMetadataConfig)

	return router
}

func TestGetUserMetadataConfig(t *testing.T) {
	t.Run("no metadata service", func(t *testing.T) {
		handler := NewHandler(&mockUserStore{}, nil, nil)
		router := setupTestRouterForMetadata(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/users/metadata/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})

	t.Run("no configuration exists - returns empty schema", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{
			activeConfig: nil,
		}
		metadataService := coreUser.NewMetadataService(mockRepo)
		handler := NewHandler(&mockUserStore{}, nil, metadataService)
		router := setupTestRouterForMetadata(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/users/metadata/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var result map[string]any
		err := json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result, "data")
	})

	t.Run("with configuration", func(t *testing.T) {
		schema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"department": map[string]any{
					"type": "string",
				},
			},
		}
		mockRepo := &mockUserMetadataConfigRepo{
			activeConfig: &repositories.UserMetadataConfigEntity{
				ID: "test-id",
				Config: &sharedUser.UserMetadataConfig{
					Schema: schema,
				},
			},
		}
		metadataService := coreUser.NewMetadataService(mockRepo)
		handler := NewHandler(&mockUserStore{}, nil, metadataService)
		router := setupTestRouterForMetadata(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/users/metadata/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var result map[string]any
		err := json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result, "data")

		data := result["data"].(map[string]any)
		assert.Contains(t, data, "schema")
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{
			getActiveErr: errors.New("db error"),
		}
		metadataService := coreUser.NewMetadataService(mockRepo)
		handler := NewHandler(&mockUserStore{}, nil, metadataService)
		router := setupTestRouterForMetadata(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/users/metadata/config", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestGetUserMetadataConfigHistory(t *testing.T) {
	t.Run("no metadata service", func(t *testing.T) {
		handler := NewHandler(&mockUserStore{}, nil, nil)
		router := setupTestRouterForMetadata(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/users/metadata/config/history", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})

	t.Run("get history with default limit", func(t *testing.T) {
		history := []repositories.UserMetadataConfigEntity{
			{
				ID: "id1",
				Config: &sharedUser.UserMetadataConfig{
					Schema: map[string]any{"type": "object"},
				},
			},
		}
		mockRepo := &mockUserMetadataConfigRepo{
			history: history,
		}
		metadataService := coreUser.NewMetadataService(mockRepo)
		handler := NewHandler(&mockUserStore{}, nil, metadataService)
		router := setupTestRouterForMetadata(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/users/metadata/config/history", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)

		var result map[string]any
		err := json.Unmarshal(rec.Body.Bytes(), &result)
		require.NoError(t, err)
		assert.Contains(t, result, "history")
	})

	t.Run("get history with custom limit", func(t *testing.T) {
		history := []repositories.UserMetadataConfigEntity{
			{ID: "id1", Config: &sharedUser.UserMetadataConfig{Schema: map[string]any{"type": "object"}}},
			{ID: "id2", Config: &sharedUser.UserMetadataConfig{Schema: map[string]any{"type": "object"}}},
			{ID: "id3", Config: &sharedUser.UserMetadataConfig{Schema: map[string]any{"type": "object"}}},
		}
		mockRepo := &mockUserMetadataConfigRepo{
			history: history,
		}
		metadataService := coreUser.NewMetadataService(mockRepo)
		handler := NewHandler(&mockUserStore{}, nil, metadataService)
		router := setupTestRouterForMetadata(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/users/metadata/config/history?limit=2", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{
			getHistoryErr: errors.New("db error"),
		}
		metadataService := coreUser.NewMetadataService(mockRepo)
		handler := NewHandler(&mockUserStore{}, nil, metadataService)
		router := setupTestRouterForMetadata(handler)

		req := httptest.NewRequest(http.MethodGet, "/api/users/metadata/config/history", nil)
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}

func TestSaveUserMetadataConfig(t *testing.T) {
	t.Run("no metadata service", func(t *testing.T) {
		handler := NewHandler(&mockUserStore{}, nil, nil)
		router := setupTestRouterForMetadata(handler)

		body := map[string]any{
			"config": map[string]any{
				"schema": map[string]any{
					"type": "object",
				},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/users/metadata/config", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})

	t.Run("valid configuration", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{}
		metadataService := coreUser.NewMetadataService(mockRepo)
		handler := NewHandler(&mockUserStore{}, nil, metadataService)
		router := setupTestRouterForMetadata(handler)

		body := map[string]any{
			"config": map[string]any{
				"schema": map[string]any{
					"type": "object",
					"properties": map[string]any{
						"department": map[string]any{
							"type": "string",
						},
					},
				},
			},
			"description": "Initial configuration",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/users/metadata/config", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, mockRepo.saveCalled)
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{}
		metadataService := coreUser.NewMetadataService(mockRepo)
		handler := NewHandler(&mockUserStore{}, nil, metadataService)
		router := setupTestRouterForMetadata(handler)

		req := httptest.NewRequest(http.MethodPut, "/api/users/metadata/config", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, mockRepo.saveCalled)
	})

	t.Run("validation error", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{}
		metadataService := coreUser.NewMetadataService(mockRepo)
		handler := NewHandler(&mockUserStore{}, nil, metadataService)
		router := setupTestRouterForMetadata(handler)

		// Missing required schema field
		body := map[string]any{
			"config": map[string]any{
				"schema": nil,
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/users/metadata/config", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
		assert.False(t, mockRepo.saveCalled)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{
			saveErr: errors.New("db error"),
		}
		metadataService := coreUser.NewMetadataService(mockRepo)
		handler := NewHandler(&mockUserStore{}, nil, metadataService)
		router := setupTestRouterForMetadata(handler)

		body := map[string]any{
			"config": map[string]any{
				"schema": map[string]any{
					"type": "object",
				},
			},
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPut, "/api/users/metadata/config", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})
}

func TestRollbackUserMetadataConfig(t *testing.T) {
	t.Run("no metadata service", func(t *testing.T) {
		handler := NewHandler(&mockUserStore{}, nil, nil)
		router := setupTestRouterForMetadata(handler)

		body := map[string]any{
			"id": "rollback-id",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/users/metadata/config/rollback", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusServiceUnavailable, rec.Code)
	})

	t.Run("successful rollback", func(t *testing.T) {
		oldConfig := &sharedUser.UserMetadataConfig{
			Schema: map[string]any{
				"type": "object",
			},
		}
		mockRepo := &mockUserMetadataConfigRepo{
			history: []repositories.UserMetadataConfigEntity{
				{
					ID:     "rollback-id",
					Config: oldConfig,
				},
			},
		}
		metadataService := coreUser.NewMetadataService(mockRepo)
		handler := NewHandler(&mockUserStore{}, nil, metadataService)
		router := setupTestRouterForMetadata(handler)

		body := map[string]any{
			"id": "rollback-id",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/users/metadata/config/rollback", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusOK, rec.Code)
		assert.True(t, mockRepo.saveCalled)
	})

	t.Run("invalid request body", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{}
		metadataService := coreUser.NewMetadataService(mockRepo)
		handler := NewHandler(&mockUserStore{}, nil, metadataService)
		router := setupTestRouterForMetadata(handler)

		req := httptest.NewRequest(http.MethodPost, "/api/users/metadata/config/rollback", bytes.NewReader([]byte("invalid json")))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("missing required id field", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{}
		metadataService := coreUser.NewMetadataService(mockRepo)
		handler := NewHandler(&mockUserStore{}, nil, metadataService)
		router := setupTestRouterForMetadata(handler)

		body := map[string]any{}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/users/metadata/config/rollback", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusBadRequest, rec.Code)
	})

	t.Run("rollback error", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{
			getByIDErr: errors.New("not found"),
		}
		metadataService := coreUser.NewMetadataService(mockRepo)
		handler := NewHandler(&mockUserStore{}, nil, metadataService)
		router := setupTestRouterForMetadata(handler)

		body := map[string]any{
			"id": "nonexistent-id",
		}
		bodyBytes, _ := json.Marshal(body)

		req := httptest.NewRequest(http.MethodPost, "/api/users/metadata/config/rollback", bytes.NewReader(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		router.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
	})
}
