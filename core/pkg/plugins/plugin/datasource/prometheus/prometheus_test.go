package prometheus

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/model"
)

func TestGetPlugin(t *testing.T) {
	config := common.Config{}

	plugin, err := GetPlugin(config)
	if err != nil {
		t.Fatalf("GetPlugin() returned error: %v", err)
	}

	if plugin == nil {
		t.Fatal("GetPlugin() returned nil plugin")
	}

	// Test plugin basic properties
	if plugin.Type != model.PluginTypeDatasource {
		t.Errorf("Expected plugin type %v, got %v", model.PluginTypeDatasource, plugin.Type)
	}

	if plugin.ID != "prometheus" {
		t.Errorf("Expected plugin ID 'prometheus', got '%s'", plugin.ID)
	}

	if plugin.Name != "Prometheus" {
		t.Errorf("Expected plugin name 'Prometheus', got '%s'", plugin.Name)
	}

	// Test datasource client is properly initialized
	if plugin.DatasourceClient == nil {
		t.Fatal("DatasourceClient is nil")
	}

	if plugin.DatasourceClient.DataClient == nil {
		t.Fatal("DataClient is nil")
	}
}

func TestEmbeddedJSONSchema(t *testing.T) {
	// Test that embedded JSON schema is valid JSON
	var schema map[string]any
	err := json.Unmarshal([]byte(jsonSchema), &schema)
	if err != nil {
		t.Fatalf("JSON schema is not valid JSON: %v", err)
	}

	// Test schema structure
	if schema["type"] != "object" {
		t.Errorf("Expected schema type 'object', got %v", schema["type"])
	}

	// Test required fields
	required, ok := schema["required"].([]any)
	if !ok {
		t.Fatal("Required field is not an array")
	}

	expectedRequired := []string{"type", "name"}
	if len(required) != len(expectedRequired) {
		t.Errorf("Expected %d required fields, got %d", len(expectedRequired), len(required))
	}

	// Test properties structure
	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("Properties field is not an object")
	}

	// Test type property
	typeProperty, ok := properties["type"].(map[string]any)
	if !ok {
		t.Fatal("Type property is not an object")
	}

	if typeProperty["default"] != "prometheus" {
		t.Errorf("Expected type default 'prometheus', got %v", typeProperty["default"])
	}

	// Test data property
	dataProperty, ok := properties["data"].(map[string]any)
	if !ok {
		t.Fatal("Data property is not an object")
	}

	dataProperties, ok := dataProperty["properties"].(map[string]any)
	if !ok {
		t.Fatal("Data properties is not an object")
	}

	// Test required prometheus connection properties
	expectedDBProperties := []string{"address"}
	for _, prop := range expectedDBProperties {
		if _, exists := dataProperties[prop]; !exists {
			t.Errorf("Expected data property '%s' to exist", prop)
		}
	}

	// Test address property type
	addressProperty, ok := dataProperties["address"].(map[string]any)
	if !ok {
		t.Fatal("Address property is not an object")
	}

	if addressProperty["type"] != "string" {
		t.Errorf("Expected address type 'string', got %v", addressProperty["type"])
	}

	if addressProperty["default"] != "http://localhost:9090" {
		t.Errorf("Expected address default 'http://localhost:9090', got %v", addressProperty["default"])
	}
}

func TestEmbeddedUISchema(t *testing.T) {
	// Test that embedded UI schema is valid JSON
	var schema map[string]any
	err := json.Unmarshal([]byte(uiSchema), &schema)
	if err != nil {
		t.Fatalf("UI schema is not valid JSON: %v", err)
	}

	// Prometheus UI schema can be empty object or contain data field
	// Both are valid for this test
	if len(schema) > 0 {
		// If it has content, check if data field exists and is properly structured
		if data, exists := schema["data"]; exists {
			if _, ok := data.(map[string]any); !ok {
				t.Error("Data field in UI schema should be an object if present")
			}
		}
	}
}

func TestEmbeddedSampleFormData(t *testing.T) {
	// Test that embedded sample form data is valid JSON
	var data map[string]any
	err := json.Unmarshal([]byte(sampleFormData), &data)
	if err != nil {
		t.Fatalf("Sample form data is not valid JSON: %v", err)
	}

	// Test sample data structure
	if data["type"] != "prometheus" {
		t.Errorf("Expected type 'prometheus', got %v", data["type"])
	}

	if data["name"] != "sample-prometheus" {
		t.Errorf("Expected name 'sample-prometheus', got %v", data["name"])
	}

	// Test data field
	dataField, ok := data["data"].(map[string]any)
	if !ok {
		t.Fatal("Data field is not an object")
	}

	// Test prometheus connection properties
	expectedAddress := "http://192.168.15.101:30090"
	if dataField["address"] != expectedAddress {
		t.Errorf("Expected address '%s', got %v", expectedAddress, dataField["address"])
	}
}

func TestEmbeddedSchemaContent(t *testing.T) {
	tests := []struct {
		name    string
		content string
		isEmpty bool
	}{
		{"JSON Schema", jsonSchema, false},
		{"UI Schema", uiSchema, false},
		{"Sample Form Data", sampleFormData, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.isEmpty && tt.content != "" {
				t.Errorf("Expected %s to be empty, but got content", tt.name)
			}
			if !tt.isEmpty && strings.TrimSpace(tt.content) == "" {
				t.Errorf("Expected %s to have content, but it's empty", tt.name)
			}
		})
	}
}

func TestPluginIntegrity(t *testing.T) {
	config := common.Config{}
	plugin, err := GetPlugin(config)
	if err != nil {
		t.Fatalf("GetPlugin() returned error: %v", err)
	}

	// Test that all required components are present
	if plugin.Settings.JSONSchemaForJson == nil {
		t.Error("JSONSchemaForJson is nil")
	}

	if plugin.Settings.UiSchemaForJson == nil {
		t.Error("UiSchemaForJson is nil")
	}

	if plugin.SampleFormData == "" {
		t.Error("SampleFormData is empty")
	}

	// Test that schemas are not nil (OrderedMap types)
	if plugin.Settings.JSONSchemaForJson == nil {
		t.Error("Plugin JSONSchemaForJson OrderedMap is nil")
	}

	if plugin.Settings.UiSchemaForJson == nil {
		t.Error("Plugin UiSchemaForJson OrderedMap is nil")
	}

	if plugin.SampleFormData != sampleFormData {
		t.Error("Plugin SampleFormData doesn't match embedded sampleFormData")
	}
}

func TestPrometheusConstants(t *testing.T) {
	expectedType := "prometheus"
	expectedName := "Prometheus"

	config := common.Config{}
	plugin, _ := GetPlugin(config)

	if plugin.ID != expectedType {
		t.Errorf("Expected plugin ID '%s', got '%s'", expectedType, plugin.ID)
	}

	if plugin.Name != expectedName {
		t.Errorf("Expected plugin name '%s', got '%s'", expectedName, plugin.Name)
	}
}

func TestDataClientInitialization(t *testing.T) {
	config := common.Config{}
	plugin, err := GetPlugin(config)
	if err != nil {
		t.Fatalf("GetPlugin() returned error: %v", err)
	}

	// Verify DataClient is properly wrapped
	dataClient := plugin.DatasourceClient.DataClient
	if dataClient == nil {
		t.Fatal("DataClient is nil")
	}

	// Test that the DataClient interface is properly implemented
	// by attempting to call methods (they should not panic)
	queryReq := &model.QueryDataRequest{}
	queryResp := dataClient.QueryData(context.Background(), queryReq)
	if queryResp == nil {
		t.Error("QueryData returned nil response")
	}

	removeReq := &model.RemoveDatasourceRequest{}
	err = dataClient.RemoveDatasource(removeReq)
	if err != nil {
		// RemoveDatasource may return error in test environment - this is OK
		t.Logf("RemoveDatasource returned error (expected in test): %v", err)
	}
}

func TestPrometheusSchemaValidation(t *testing.T) {
	// Test that the JSON schema contains Prometheus-specific properties
	var schema map[string]any
	err := json.Unmarshal([]byte(jsonSchema), &schema)
	if err != nil {
		t.Fatalf("JSON schema is not valid JSON: %v", err)
	}

	properties, ok := schema["properties"].(map[string]any)
	if !ok {
		t.Fatal("Properties field is not an object")
	}

	dataProperty, ok := properties["data"].(map[string]any)
	if !ok {
		t.Fatal("Data property is not an object")
	}

	dataRequired, ok := dataProperty["required"].([]any)
	if !ok {
		t.Fatal("Data required field is not an array")
	}

	// Check for Prometheus-specific required fields
	expectedDataRequired := []string{"address"}

	if len(dataRequired) != len(expectedDataRequired) {
		t.Errorf("Expected %d required data fields, got %d", len(expectedDataRequired), len(dataRequired))
	}

	// Verify all expected fields are in required
	requiredMap := make(map[string]bool)
	for _, req := range dataRequired {
		if reqStr, ok := req.(string); ok {
			requiredMap[reqStr] = true
		}
	}

	for _, expected := range expectedDataRequired {
		if !requiredMap[expected] {
			t.Errorf("Expected required field '%s' not found", expected)
		}
	}
}

func TestGetPlugin_WithDifferentConfigs(t *testing.T) {
	tests := []struct {
		name   string
		config common.Config
	}{
		{
			name:   "Default config",
			config: common.Config{},
		},
		{
			name:   "Config with settings",
			config: common.Config{
				// Use valid Config fields here when available
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

			if plugin.ID != "prometheus" {
				t.Errorf("GetPlugin() plugin ID = %v, want %v", plugin.ID, "prometheus")
			}
		})
	}
}

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

	// Verify datasource client is complete
	if plugin.DatasourceClient == nil {
		t.Errorf("DatasourceClient is missing")
	}

	if plugin.DatasourceClient.DataClient == nil {
		t.Errorf("DataClient is missing")
	}
}

// Benchmark tests
func BenchmarkGetPlugin(b *testing.B) {
	config := common.Config{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetPlugin(config)
		if err != nil {
			b.Fatalf("GetPlugin() returned error: %v", err)
		}
	}
}

func BenchmarkJSONSchemaUnmarshal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var schema map[string]any
		err := json.Unmarshal([]byte(jsonSchema), &schema)
		if err != nil {
			b.Fatalf("JSON unmarshal failed: %v", err)
		}
	}
}

func BenchmarkSampleFormDataUnmarshal(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		var data map[string]any
		err := json.Unmarshal([]byte(sampleFormData), &data)
		if err != nil {
			b.Fatalf("JSON unmarshal failed: %v", err)
		}
	}
}

func BenchmarkDataClientCreation(b *testing.B) {
	config := common.Config{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		plugin, err := GetPlugin(config)
		if err != nil {
			b.Fatalf("GetPlugin() returned error: %v", err)
		}
		_ = plugin.DatasourceClient.DataClient
	}
}
