package clickhouse

import (
	"strings"
	"testing"

	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/model"
)

// TestGetPlugin tests the GetPlugin function
func TestGetPlugin(t *testing.T) {
	config := common.Config{}

	plugin, err := GetPlugin(config)

	if err != nil {
		t.Errorf("GetPlugin() error = %v, want nil", err)
		return
	}

	if plugin == nil {
		t.Errorf("GetPlugin() returned nil plugin")
		return
	}

	// Verify plugin properties
	if plugin.Type != model.PluginTypeDatasource {
		t.Errorf("GetPlugin() plugin type = %v, want %v", plugin.Type, model.PluginTypeDatasource)
	}

	if plugin.ID != "clickhouse" {
		t.Errorf("GetPlugin() plugin ID = %v, want %v", plugin.ID, "clickhouse")
	}

	if plugin.Name != "Clickhouse" {
		t.Errorf("GetPlugin() plugin name = %v, want %v", plugin.Name, "Clickhouse")
	}

	// Verify schemas are embedded
	if plugin.Settings.JSONSchemaForJson == nil {
		t.Errorf("GetPlugin() plugin JSONSchemaForJson is nil")
	}

	if plugin.Settings.UiSchemaForJson == nil {
		t.Errorf("GetPlugin() plugin UiSchemaForJson is nil")
	}

	if plugin.SampleFormData == "" {
		t.Errorf("GetPlugin() plugin SampleFormData is empty")
	}

	// Verify datasource client is set
	if plugin.DatasourceClient == nil {
		t.Errorf("GetPlugin() plugin DatasourceClient is nil")
	}

	if plugin.DatasourceClient.DataClient == nil {
		t.Errorf("GetPlugin() plugin DatasourceClient.DataClient is nil")
	}
}

// TestGetPlugin_EmbeddedSchemas tests that embedded schemas are properly loaded
func TestGetPlugin_EmbeddedSchemas(t *testing.T) {
	config := common.Config{}

	plugin, err := GetPlugin(config)

	if err != nil {
		t.Errorf("GetPlugin() error = %v, want nil", err)
		return
	}

	// Test JSON Schema
	t.Run("JSON Schema", func(t *testing.T) {
		if plugin.Settings.JSONSchemaForJson == nil {
			t.Errorf("JSONSchemaForJson is nil")
			return
		}

		// Check that the embedded schema contains expected ClickHouse properties
		if !strings.Contains(jsonSchema, "clickhouse") {
			t.Errorf("JSON Schema missing clickhouse reference")
		}

		if !strings.Contains(jsonSchema, "type") {
			t.Errorf("JSON Schema missing type property")
		}

		if !strings.Contains(jsonSchema, "name") {
			t.Errorf("JSON Schema missing name property")
		}
	})

	// Test UI Schema
	t.Run("UI Schema", func(t *testing.T) {
		if plugin.Settings.UiSchemaForJson == nil {
			t.Errorf("UiSchemaForJson is nil")
			return
		}

		// Check that the embedded UI schema contains password widget configuration
		if !strings.Contains(uiSchema, "password") {
			t.Errorf("UI Schema missing password widget configuration")
		}
	})

	// Test Sample Form Data
	t.Run("Sample Form Data", func(t *testing.T) {
		if plugin.SampleFormData == "" {
			t.Errorf("Sample Form Data is empty")
			return
		}

		// Check that it contains expected sample data
		expectedFields := []string{
			"clickhouse",
			"sample-clickhouse",
			"host",
			"port",
			"database",
			"username",
		}

		for _, field := range expectedFields {
			if !strings.Contains(plugin.SampleFormData, field) {
				t.Errorf("Sample Form Data missing expected field: %s", field)
			}
		}
	})
}

// TestGetPlugin_SchemaValidation tests JSON schema validation
func TestGetPlugin_SchemaValidation(t *testing.T) {
	config := common.Config{}

	plugin, err := GetPlugin(config)
	if err != nil {
		t.Errorf("GetPlugin() error = %v, want nil", err)
		return
	}

	// Test valid JSON instance validation
	validInstance := `{
		"type": "clickhouse",
		"name": "test-clickhouse",
		"data": {
			"host": "localhost",
			"port": 9000,
			"database": "default",
			"username": "default",
			"password": "password",
			"max_open_connection": 10,
			"max_lifetime": 14400
		}
	}`

	err = plugin.ValidateJSONSchemaInstance(validInstance)
	if err != nil {
		t.Errorf("ValidateJSONSchemaInstance() with valid instance failed: %v", err)
	}

	// Test invalid JSON instance validation
	invalidInstance := `{
"type": "clickhouse"
}`

	err = plugin.ValidateJSONSchemaInstance(invalidInstance)
	if err == nil {
		t.Errorf("ValidateJSONSchemaInstance() with invalid instance should have failed")
	}
}

// TestGetPlugin_EmbeddedFileContent tests the content of embedded files
func TestGetPlugin_EmbeddedFileContent(t *testing.T) {
	// Test jsonSchema content
	t.Run("jsonSchema content", func(t *testing.T) {
		if jsonSchema == "" {
			t.Errorf("jsonSchema is empty")
			return
		}

		// Should contain ClickHouse-specific properties
		expectedContent := []string{
			"clickhouse",
			"properties",
			"type",
			"name",
			"data",
		}

		for _, content := range expectedContent {
			if !strings.Contains(jsonSchema, content) {
				t.Errorf("jsonSchema missing expected content: %s", content)
			}
		}
	})

	// Test uiSchema content
	t.Run("uiSchema content", func(t *testing.T) {
		if uiSchema == "" {
			t.Errorf("uiSchema is empty")
			return
		}

		// Should contain password widget
		if !strings.Contains(uiSchema, "password") {
			t.Errorf("uiSchema missing password widget")
		}
	})

	// Test sampleFormData content
	t.Run("sampleFormData content", func(t *testing.T) {
		if sampleFormData == "" {
			t.Errorf("sampleFormData is empty")
			return
		}

		// Should contain sample ClickHouse configuration
		expectedSampleData := []string{
			"clickhouse",
			"sample-clickhouse",
			"host",
			"port",
			"database",
			"username",
		}

		for _, data := range expectedSampleData {
			if !strings.Contains(sampleFormData, data) {
				t.Errorf("sampleFormData missing expected data: %s", data)
			}
		}
	})
}

// BenchmarkGetPlugin benchmarks the GetPlugin function
func BenchmarkGetPlugin(b *testing.B) {
	config := common.Config{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plugin, err := GetPlugin(config)
		if err != nil {
			b.Errorf("GetPlugin() error = %v", err)
		}
		if plugin == nil {
			b.Errorf("GetPlugin() returned nil")
		}
	}
}

// TestGetPlugin_WithDifferentConfigs tests GetPlugin with various configurations
func TestGetPlugin_WithDifferentConfigs(t *testing.T) {
	tests := []struct {
		name   string
		config common.Config
	}{
		{
			name:   "default config",
			config: common.Config{},
		},
		{
			name: "config with serve settings",
			config: common.Config{
				Serve: common.ServeConfig{},
			},
		},
		{
			name: "config with plugins settings",
			config: common.Config{
				Plugins: common.PluginsConfig{
					Path: "/tmp/plugins",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			plugin, err := GetPlugin(tt.config)

			if err != nil {
				t.Errorf("GetPlugin() error = %v, want nil", err)
				return
			}

			if plugin == nil {
				t.Errorf("GetPlugin() returned nil plugin")
				return
			}

			// Basic validation that plugin was created
			if plugin.Type != model.PluginTypeDatasource {
				t.Errorf("GetPlugin() plugin type = %v, want %v", plugin.Type, model.PluginTypeDatasource)
			}

			if plugin.ID != "clickhouse" {
				t.Errorf("GetPlugin() plugin ID = %v, want %v", plugin.ID, "clickhouse")
			}
		})
	}
}

// TestGetPlugin_MultipleInstances tests creating multiple plugin instances
func TestGetPlugin_MultipleInstances(t *testing.T) {
	config := common.Config{}

	// Create multiple instances
	plugin1, err1 := GetPlugin(config)
	plugin2, err2 := GetPlugin(config)

	if err1 != nil {
		t.Errorf("GetPlugin() first call error = %v, want nil", err1)
	}

	if err2 != nil {
		t.Errorf("GetPlugin() second call error = %v, want nil", err2)
	}

	if plugin1 == nil || plugin2 == nil {
		t.Errorf("GetPlugin() returned nil plugin")
		return
	}

	// Both plugins should have the same configuration
	if plugin1.ID != plugin2.ID {
		t.Errorf("Plugin IDs differ: %v vs %v", plugin1.ID, plugin2.ID)
	}

	if plugin1.Name != plugin2.Name {
		t.Errorf("Plugin Names differ: %v vs %v", plugin1.Name, plugin2.Name)
	}

	if plugin1.Type != plugin2.Type {
		t.Errorf("Plugin Types differ: %v vs %v", plugin1.Type, plugin2.Type)
	}

	// Sample Form Data should be identical
	if plugin1.SampleFormData != plugin2.SampleFormData {
		t.Errorf("Sample Form Data differs between instances")
	}
}

// TestGetPlugin_PluginValidation tests plugin validation
func TestGetPlugin_PluginValidation(t *testing.T) {
	config := common.Config{}

	plugin, err := GetPlugin(config)

	if err != nil {
		t.Errorf("GetPlugin() error = %v, want nil", err)
		return
	}

	// Test plugin validation method if available
	if plugin.Type != model.PluginTypeDatasource {
		t.Errorf("Invalid plugin type: %v", plugin.Type)
	}

	if plugin.ID == "" {
		t.Errorf("Plugin ID cannot be empty")
	}

	if plugin.Name == "" {
		t.Errorf("Plugin Name cannot be empty")
	}

	// Verify all required components are present
	if plugin.Settings.JSONSchemaForJson == nil {
		t.Errorf("JSONSchemaForJson is missing")
	}

	if plugin.Settings.UiSchemaForJson == nil {
		t.Errorf("UiSchemaForJson is missing")
	}

	if plugin.SampleFormData == "" {
		t.Errorf("SampleFormData is missing or empty")
	}

	if plugin.DatasourceClient == nil {
		t.Errorf("DatasourceClient is missing")
	}
}

// TestGetPlugin_InvalidJSONValidation tests invalid JSON validation scenarios
func TestGetPlugin_InvalidJSONValidation(t *testing.T) {
	config := common.Config{}

	plugin, err := GetPlugin(config)
	if err != nil {
		t.Errorf("GetPlugin() error = %v, want nil", err)
		return
	}

	invalidCases := []struct {
		name     string
		instance string
	}{
		{
			name:     "missing required fields",
			instance: `{"type": "clickhouse"}`,
		},
		{
			name:     "missing name field",
			instance: `{"type": "clickhouse", "data": {}}`,
		},
		{
			name: "missing required data fields when data present",
			instance: `{
				"type": "clickhouse",
				"name": "test",
				"data": {"host": "localhost"}
			}`,
		},
		{
			name: "invalid data types",
			instance: `{
				"type": "clickhouse",
				"name": "test",
				"data": {
					"host": "localhost",
					"port": "invalid_port",
					"database": "default",
					"username": "default",
					"password": "password",
					"max_open_connection": 10,
					"max_lifetime": 14400
				}
			}`,
		},
		{
			name:     "malformed JSON",
			instance: `{"type": "clickhouse", "name": "test"`,
		},
	}

	for _, tc := range invalidCases {
		t.Run(tc.name, func(t *testing.T) {
			err := plugin.ValidateJSONSchemaInstance(tc.instance)
			if err == nil {
				t.Errorf("ValidateJSONSchemaInstance() with invalid instance should have failed for case: %s", tc.name)
			}
		})
	}
}

// BenchmarkGetPlugin_SchemaAccess benchmarks schema access
func BenchmarkGetPlugin_SchemaAccess(b *testing.B) {
	config := common.Config{}

	plugin, err := GetPlugin(config)
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = plugin.Settings.JSONSchemaForJson
		_ = plugin.Settings.UiSchemaForJson
		_ = plugin.SampleFormData
	}
}

// BenchmarkGetPlugin_Validation benchmarks JSON schema validation
func BenchmarkGetPlugin_Validation(b *testing.B) {
	config := common.Config{}

	plugin, err := GetPlugin(config)
	if err != nil {
		b.Fatalf("Setup failed: %v", err)
	}

	validInstance := `{
		"type": "clickhouse",
		"name": "test-clickhouse",
		"data": {
			"host": "localhost",
			"port": 9000,
			"database": "default",
			"username": "default",
			"password": "password",
			"max_open_connection": 10,
			"max_lifetime": 14400
		}
	}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := plugin.ValidateJSONSchemaInstance(validInstance)
		if err != nil {
			b.Errorf("Validation failed: %v", err)
		}
	}
}
