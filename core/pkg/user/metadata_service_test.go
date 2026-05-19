package user

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/internal/repositories"
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

func TestMetadataService_GetActiveConfig(t *testing.T) {
	t.Run("no configuration exists - returns empty schema", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{
			activeConfig: nil,
		}
		service := NewMetadataService(mockRepo)

		config, err := service.GetActiveConfig(context.Background())
		require.NoError(t, err)
		require.NotNil(t, config)
		assert.NotNil(t, config.Schema)
		assert.Equal(t, "object", config.Schema["type"])
	})

	t.Run("configuration exists", func(t *testing.T) {
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
		service := NewMetadataService(mockRepo)

		config, err := service.GetActiveConfig(context.Background())
		require.NoError(t, err)
		require.NotNil(t, config)
		assert.Equal(t, schema, config.Schema)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{
			getActiveErr: errors.New("db error"),
		}
		service := NewMetadataService(mockRepo)

		config, err := service.GetActiveConfig(context.Background())
		assert.Error(t, err)
		assert.Nil(t, config)
	})
}

func TestMetadataService_GetHistory(t *testing.T) {
	t.Run("get history with limit", func(t *testing.T) {
		history := []repositories.UserMetadataConfigEntity{
			{
				ID: "id1",
				Config: &sharedUser.UserMetadataConfig{
					Schema: map[string]any{"type": "object"},
				},
			},
			{
				ID: "id2",
				Config: &sharedUser.UserMetadataConfig{
					Schema: map[string]any{"type": "object"},
				},
			},
			{
				ID: "id3",
				Config: &sharedUser.UserMetadataConfig{
					Schema: map[string]any{"type": "object"},
				},
			},
		}
		mockRepo := &mockUserMetadataConfigRepo{
			history: history,
		}
		service := NewMetadataService(mockRepo)

		result, err := service.GetHistory(context.Background(), 2)
		require.NoError(t, err)
		assert.Len(t, result, 2)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{
			getHistoryErr: errors.New("db error"),
		}
		service := NewMetadataService(mockRepo)

		result, err := service.GetHistory(context.Background(), 10)
		assert.Error(t, err)
		assert.Nil(t, result)
	})
}

func TestMetadataService_SaveConfig(t *testing.T) {
	t.Run("valid configuration", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{}
		service := NewMetadataService(mockRepo)

		schema := map[string]any{
			"type": "object",
			"properties": map[string]any{
				"department": map[string]any{
					"type": "string",
				},
			},
		}
		config := &sharedUser.UserMetadataConfig{
			Schema: schema,
		}
		username := "admin"
		desc := "Initial configuration"

		err := service.SaveConfig(context.Background(), config, &username, &desc)
		require.NoError(t, err)
		assert.True(t, mockRepo.saveCalled)
		require.NotNil(t, mockRepo.lastSavedConfig)
		assert.Equal(t, schema, mockRepo.lastSavedConfig.Schema)
	})

	t.Run("invalid configuration - missing schema", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{}
		service := NewMetadataService(mockRepo)

		config := &sharedUser.UserMetadataConfig{
			Schema: nil,
		}

		err := service.SaveConfig(context.Background(), config, nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "schema is required")
		assert.False(t, mockRepo.saveCalled)
	})

	t.Run("invalid configuration - schema not object", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{}
		service := NewMetadataService(mockRepo)

		config := &sharedUser.UserMetadataConfig{
			Schema: map[string]any{
				"type": "string", // Should be object
			},
		}

		err := service.SaveConfig(context.Background(), config, nil, nil)
		assert.Error(t, err)
		assert.Contains(t, err.Error(), "must be 'object'")
		assert.False(t, mockRepo.saveCalled)
	})

	t.Run("database error", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{
			saveErr: errors.New("db error"),
		}
		service := NewMetadataService(mockRepo)

		config := &sharedUser.UserMetadataConfig{
			Schema: map[string]any{
				"type": "object",
			},
		}

		err := service.SaveConfig(context.Background(), config, nil, nil)
		assert.Error(t, err)
	})
}

func TestMetadataService_Rollback(t *testing.T) {
	t.Run("successful rollback", func(t *testing.T) {
		oldConfig := &sharedUser.UserMetadataConfig{
			Schema: map[string]any{
				"type": "object",
				"properties": map[string]any{
					"oldField": map[string]any{
						"type": "string",
					},
				},
			},
		}
		// Marshal config to JSON for the mock
		configJSON, _ := json.Marshal(oldConfig)
		mockRepo := &mockUserMetadataConfigRepo{
			history: []repositories.UserMetadataConfigEntity{
				{
					ID:         "rollback-id",
					Config:     oldConfig,
					ConfigJSON: string(configJSON),
				},
			},
		}
		service := NewMetadataService(mockRepo)

		username := "admin"
		err := service.Rollback(context.Background(), "rollback-id", &username)
		require.NoError(t, err)
		assert.True(t, mockRepo.saveCalled)
		require.NotNil(t, mockRepo.lastSavedConfig)
		assert.Equal(t, oldConfig.Schema, mockRepo.lastSavedConfig.Schema)
	})

	t.Run("config not found", func(t *testing.T) {
		mockRepo := &mockUserMetadataConfigRepo{
			history:    []repositories.UserMetadataConfigEntity{},
			getByIDErr: errors.New("not found"),
		}
		service := NewMetadataService(mockRepo)

		err := service.Rollback(context.Background(), "nonexistent-id", nil)
		assert.Error(t, err)
		assert.False(t, mockRepo.saveCalled)
	})

	t.Run("database error on save", func(t *testing.T) {
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
			saveErr: errors.New("db error"),
		}
		service := NewMetadataService(mockRepo)

		err := service.Rollback(context.Background(), "rollback-id", nil)
		assert.Error(t, err)
	})
}

func TestValidateUserMetadataConfig(t *testing.T) {
	tests := []struct {
		name    string
		config  *sharedUser.UserMetadataConfig
		wantErr bool
		errMsg  string
	}{
		{
			name:    "nil config",
			config:  nil,
			wantErr: true,
			errMsg:  "config cannot be nil",
		},
		{
			name: "nil schema",
			config: &sharedUser.UserMetadataConfig{
				Schema: nil,
			},
			wantErr: true,
			errMsg:  "schema is required",
		},
		{
			name: "schema not an object",
			config: &sharedUser.UserMetadataConfig{
				Schema: map[string]any{
					"type": "string",
				},
			},
			wantErr: true,
			errMsg:  "must be 'object'",
		},
		{
			name: "valid schema without uiSchema",
			config: &sharedUser.UserMetadataConfig{
				Schema: map[string]any{
					"type": "object",
					"properties": map[string]any{
						"field1": map[string]any{
							"type": "string",
						},
					},
				},
			},
			wantErr: false,
		},
		{
			name: "valid schema with uiSchema",
			config: &sharedUser.UserMetadataConfig{
				Schema: map[string]any{
					"type": "object",
				},
				UISchema: map[string]any{
					"field1": map[string]any{
						"ui:widget": "textarea",
					},
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateUserMetadataConfig(tt.config)
			if tt.wantErr {
				assert.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
