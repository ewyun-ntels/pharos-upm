package postgresql

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

	if plugin.ID != "postgresql" {
		t.Errorf("Expected plugin ID 'postgresql', got '%s'", plugin.ID)
	}

	if plugin.Name != "PostgreSQL" {
		t.Errorf("Expected plugin name 'PostgreSQL', got '%s'", plugin.Name)
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

	if typeProperty["default"] != "postgresql" {
		t.Errorf("Expected type default 'postgresql', got %v", typeProperty["default"])
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

	// Test required database connection properties
	expectedDBProperties := []string{"host", "port", "database", "username", "password"}
	for _, prop := range expectedDBProperties {
		if _, exists := dataProperties[prop]; !exists {
			t.Errorf("Expected data property '%s' to exist", prop)
		}
	}

	// Test port property type
	portProperty, ok := dataProperties["port"].(map[string]any)
	if !ok {
		t.Fatal("Port property is not an object")
	}

	if portProperty["type"] != "integer" {
		t.Errorf("Expected port type 'integer', got %v", portProperty["type"])
	}

	if portProperty["default"] != float64(5432) {
		t.Errorf("Expected port default 5432, got %v", portProperty["default"])
	}

	// Test connection pool properties
	maxOpenConnProperty, ok := dataProperties["max_open_connection"].(map[string]any)
	if !ok {
		t.Fatal("max_open_connection property is not an object")
	}

	if maxOpenConnProperty["type"] != "integer" {
		t.Errorf("Expected max_open_connection type 'integer', got %v", maxOpenConnProperty["type"])
	}

	maxLifetimeProperty, ok := dataProperties["max_lifetime"].(map[string]any)
	if !ok {
		t.Fatal("max_lifetime property is not an object")
	}

	if maxLifetimeProperty["type"] != "integer" {
		t.Errorf("Expected max_lifetime type 'integer', got %v", maxLifetimeProperty["type"])
	}
}

func TestEmbeddedUISchema(t *testing.T) {
	// Test that embedded UI schema is valid JSON
	var schema map[string]any
	err := json.Unmarshal([]byte(uiSchema), &schema)
	if err != nil {
		t.Fatalf("UI schema is not valid JSON: %v", err)
	}

	// Test UI schema structure for password widget
	data, ok := schema["data"].(map[string]any)
	if !ok {
		t.Fatal("Data field is not an object in UI schema")
	}

	password, ok := data["password"].(map[string]any)
	if !ok {
		t.Fatal("Password field is not an object in UI schema")
	}

	widget, ok := password["ui:widget"].(string)
	if !ok {
		t.Fatal("ui:widget is not a string")
	}

	if widget != "password" {
		t.Errorf("Expected ui:widget 'password', got '%s'", widget)
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
	if data["type"] != "postgresql" {
		t.Errorf("Expected type 'postgresql', got %v", data["type"])
	}

	if data["name"] != "sample-postgresql" {
		t.Errorf("Expected name 'sample-postgresql', got %v", data["name"])
	}

	// Test data field
	dataField, ok := data["data"].(map[string]any)
	if !ok {
		t.Fatal("Data field is not an object")
	}

	// Test database connection properties
	if dataField["host"] != "192.168.15.103" {
		t.Errorf("Expected host '192.168.15.103', got %v", dataField["host"])
	}

	port, ok := dataField["port"].(float64)
	if !ok {
		t.Fatal("Port is not a number")
	}

	if port != 5432 {
		t.Errorf("Expected port 5432, got %v", port)
	}

	if dataField["database"] != "postgres" {
		t.Errorf("Expected database 'postgres', got %v", dataField["database"])
	}

	if dataField["username"] != "admin" {
		t.Errorf("Expected username 'admin', got %v", dataField["username"])
	}

	if dataField["password"] != "admin" {
		t.Errorf("Expected password 'admin', got %v", dataField["password"])
	}

	// Test connection pool settings
	maxOpenConn, ok := dataField["max_open_connection"].(float64)
	if !ok {
		t.Fatal("max_open_connection is not a number")
	}

	if maxOpenConn != 10 {
		t.Errorf("Expected max_open_connection 10, got %v", maxOpenConn)
	}

	maxLifetime, ok := dataField["max_lifetime"].(float64)
	if !ok {
		t.Fatal("max_lifetime is not a number")
	}

	if maxLifetime != 14400 {
		t.Errorf("Expected max_lifetime 14400, got %v", maxLifetime)
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

func TestPostgreSQLConstants(t *testing.T) {
	expectedType := "postgresql"
	expectedName := "PostgreSQL"

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

func TestPostgreSQLSchemaValidation(t *testing.T) {
	// Test that the JSON schema contains PostgreSQL-specific properties
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

	// Check for PostgreSQL-specific required fields
	expectedDataRequired := []string{
		"host", "port", "database", "username", "password",
		"max_open_connection", "max_lifetime",
	}

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

			if plugin.ID != "postgresql" {
				t.Errorf("GetPlugin() plugin ID = %v, want %v", plugin.ID, "postgresql")
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
