package elasticsearch

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/model"
)

func TestGetPlugin(t *testing.T) {
	tests := []struct {
		name        string
		config      common.Config
		expectError bool
	}{
		{
			name:        "Empty config",
			config:      common.Config{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin, err := GetPlugin(tt.config)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, plugin)
			} else {
				assert.NoError(t, err)
				require.NotNil(t, plugin)

				// Verify plugin type
				assert.Equal(t, model.PluginTypeDatasource, plugin.Type)

				// Verify plugin ID and name
				assert.Equal(t, "elasticsearch", plugin.ID)
				assert.Equal(t, "Elasticsearch", plugin.Name)

				// Verify datasource client is not nil
				require.NotNil(t, plugin.DatasourceClient)
				require.NotNil(t, plugin.DatasourceClient.DataClient)
			}
		})
	}
}

func TestGetPlugin_Schemas(t *testing.T) {
	config := common.Config{}
	plugin, err := GetPlugin(config)
	require.NoError(t, err)
	require.NotNil(t, plugin)

	t.Run("JSON Schema validation", func(t *testing.T) {
		// Verify JSON schema is not nil
		assert.NotNil(t, plugin.Settings.JSONSchemaForJson)

		// Verify JSONSchema compiled successfully
		assert.NotNil(t, plugin.Settings.JSONSchema)
	})

	t.Run("UI Schema validation", func(t *testing.T) {
		// Verify UI schema is not nil
		assert.NotNil(t, plugin.Settings.UiSchemaForJson)
	})

	t.Run("Sample Form Data validation", func(t *testing.T) {
		// Verify sample form data is not empty
		assert.NotEmpty(t, plugin.SampleFormData)

		// Verify it's valid JSON
		var sampleData map[string]any
		err := json.Unmarshal([]byte(plugin.SampleFormData), &sampleData)
		assert.NoError(t, err)
		assert.NotEmpty(t, sampleData)
	})
}

func TestGetPlugin_DatasourceClient(t *testing.T) {
	config := common.Config{}

	plugin, err := GetPlugin(config)
	require.NoError(t, err)
	require.NotNil(t, plugin)

	t.Run("Datasource client structure", func(t *testing.T) {
		require.NotNil(t, plugin.DatasourceClient)
		assert.NotNil(t, plugin.DatasourceClient.DataClient)
	})

	t.Run("Data client configuration", func(t *testing.T) {
		// Verify the data client is properly configured
		dataClient := plugin.DatasourceClient.DataClient
		assert.NotNil(t, dataClient)
	})
}

func TestGetPlugin_EmbeddedFiles(t *testing.T) {
	tests := []struct {
		name     string
		schema   string
		validate func(*testing.T, string)
	}{
		{
			name:   "JSON Schema",
			schema: jsonSchema,
			validate: func(t *testing.T, content string) {
				assert.NotEmpty(t, content)
				var result map[string]any
				err := json.Unmarshal([]byte(content), &result)
				assert.NoError(t, err, "JSON Schema should be valid JSON")
			},
		},
		{
			name:   "UI Schema",
			schema: uiSchema,
			validate: func(t *testing.T, content string) {
				assert.NotEmpty(t, content)
				var result map[string]any
				err := json.Unmarshal([]byte(content), &result)
				assert.NoError(t, err, "UI Schema should be valid JSON")
			},
		},
		{
			name:   "Sample Form Data",
			schema: sampleFormData,
			validate: func(t *testing.T, content string) {
				assert.NotEmpty(t, content)
				var result map[string]any
				err := json.Unmarshal([]byte(content), &result)
				assert.NoError(t, err, "Sample Form Data should be valid JSON")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tt.validate(t, tt.schema)
		})
	}
}

func TestGetPlugin_PluginType(t *testing.T) {
	config := common.Config{}
	plugin, err := GetPlugin(config)
	require.NoError(t, err)
	require.NotNil(t, plugin)

	// Verify the plugin type is datasource
	assert.Equal(t, model.PluginTypeDatasource, plugin.Type)
	assert.NotEqual(t, "", plugin.Type)
}

func TestGetPlugin_PluginIdentity(t *testing.T) {
	config := common.Config{}
	plugin, err := GetPlugin(config)
	require.NoError(t, err)
	require.NotNil(t, plugin)

	// Test plugin identity fields
	tests := []struct {
		name     string
		field    string
		expected string
	}{
		{
			name:     "Plugin ID",
			field:    plugin.ID,
			expected: "elasticsearch",
		},
		{
			name:     "Plugin Name",
			field:    plugin.Name,
			expected: "Elasticsearch",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.expected, tt.field)
			assert.NotEmpty(t, tt.field)
		})
	}
}

func TestGetPlugin_MultipleInstances(t *testing.T) {
	// Test creating multiple plugin instances
	numInstances := 3

	plugins := make([]*model.Plugin, 0, numInstances)

	for range numInstances {
		plugin, err := GetPlugin(common.Config{})
		require.NoError(t, err)
		require.NotNil(t, plugin)

		plugins = append(plugins, plugin)

		// Verify basic properties
		assert.Equal(t, "elasticsearch", plugin.ID)
		assert.Equal(t, "Elasticsearch", plugin.Name)
		assert.NotNil(t, plugin.DatasourceClient)
	}

	// Verify all instances have the same sample form data
	if len(plugins) > 1 {
		for i := 1; i < len(plugins); i++ {
			assert.Equal(t, plugins[0].SampleFormData, plugins[i].SampleFormData)
		}
	}
}

func TestGetPlugin_PluginCreation(t *testing.T) {
	config := common.Config{}

	plugin, err := GetPlugin(config)
	require.NoError(t, err)
	require.NotNil(t, plugin)

	// Verify plugin was created successfully
	assert.NotNil(t, plugin.DatasourceClient)
	assert.Equal(t, model.PluginTypeDatasource, plugin.Type)
	assert.Equal(t, "elasticsearch", plugin.ID)
	assert.Equal(t, "Elasticsearch", plugin.Name)
}

func TestGetPlugin_SchemaConsistency(t *testing.T) {
	// Create multiple plugins and verify schemas are consistent
	plugin1, err := GetPlugin(common.Config{})
	require.NoError(t, err)

	plugin2, err := GetPlugin(common.Config{})
	require.NoError(t, err)

	// All plugins should have identical sample form data
	assert.Equal(t, plugin1.SampleFormData, plugin2.SampleFormData, "Sample form data should be identical")

	// Verify embedded variables are being used
	assert.Equal(t, sampleFormData, plugin1.SampleFormData)
}
