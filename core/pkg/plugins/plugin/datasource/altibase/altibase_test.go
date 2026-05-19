package altibase

import (
	"encoding/json"
	"testing"

	"ntels.com/pharos/core/pkg/common"
)

// TestEmbeddedContent tests that embedded content is available and valid
func TestEmbeddedContent(t *testing.T) {
	// Test that embedded variables are not empty
	if jsonSchema == "" {
		t.Error("Expected jsonSchema to be embedded and not empty")
	}

	if uiSchema == "" {
		t.Error("Expected uiSchema to be embedded and not empty")
	}

	if sampleFormData == "" {
		t.Error("Expected sampleFormData to be embedded and not empty")
	}

	// Verify each embedded content is valid JSON
	var jsonData, uiData, sampleData any

	if err := json.Unmarshal([]byte(jsonSchema), &jsonData); err != nil {
		t.Errorf("jsonSchema should be valid JSON, got error: %v", err)
	}

	if err := json.Unmarshal([]byte(uiSchema), &uiData); err != nil {
		t.Errorf("uiSchema should be valid JSON, got error: %v", err)
	}

	if err := json.Unmarshal([]byte(sampleFormData), &sampleData); err != nil {
		t.Errorf("sampleFormData should be valid JSON, got error: %v", err)
	}
}

// TestSampleFormDataStructure tests the structure of embedded sample form data
func TestSampleFormDataStructure(t *testing.T) {
	if sampleFormData == "" {
		t.Skip("No sample form data to test")
	}

	// Parse sample form data to verify it's valid JSON
	var sampleData map[string]any
	if err := json.Unmarshal([]byte(sampleFormData), &sampleData); err != nil {
		t.Errorf("Sample form data should be valid JSON, got error: %v", err)
	}

	// Check for expected fields in altibase configuration
	expectedFields := []string{"type", "name", "data"}
	for _, field := range expectedFields {
		if _, exists := sampleData[field]; !exists {
			t.Errorf("Expected field '%s' in sample form data", field)
		}
	}

	// Check data section structure if it exists
	if data, exists := sampleData["data"]; exists {
		if dataMap, ok := data.(map[string]any); ok {
			expectedDataFields := []string{"version", "dsn", "port", "uid", "database", "password", "max_open_connection", "max_lifetime"}
			for _, field := range expectedDataFields {
				if _, exists := dataMap[field]; !exists {
					t.Errorf("Expected field '%s' in data section of sample form data", field)
				}
			}
		}
	}
}

// TestJSONSchemaStructure tests the structure of the JSON schema
func TestJSONSchemaStructure(t *testing.T) {
	var schema map[string]any
	if err := json.Unmarshal([]byte(jsonSchema), &schema); err != nil {
		t.Fatalf("Failed to parse JSON schema: %v", err)
	}

	// Check for basic schema properties
	if schemaType, exists := schema["type"]; !exists || schemaType != "object" {
		t.Error("Expected schema to be of type 'object'")
	}

	// Check for required fields
	if required, exists := schema["required"]; exists {
		if requiredArray, ok := required.([]any); ok {
			expectedRequired := []string{"type", "name"}
			for _, field := range expectedRequired {
				found := false
				for _, req := range requiredArray {
					if req == field {
						found = true
						break
					}
				}
				if !found {
					t.Errorf("Expected '%s' to be in required fields", field)
				}
			}
		}
	}

	// Check for properties
	if properties, exists := schema["properties"]; exists {
		if propsMap, ok := properties.(map[string]any); ok {
			expectedProps := []string{"type", "name", "data"}
			for _, prop := range expectedProps {
				if _, exists := propsMap[prop]; !exists {
					t.Errorf("Expected property '%s' in schema", prop)
				}
			}
		}
	}
}

// TestUISchemaStructure tests the structure of the UI schema
func TestUISchemaStructure(t *testing.T) {
	var uiSchemaData map[string]any
	if err := json.Unmarshal([]byte(uiSchema), &uiSchemaData); err != nil {
		t.Fatalf("Failed to parse UI schema: %v", err)
	}

	// UI schema should be a valid object
	if len(uiSchemaData) == 0 {
		t.Log("UI schema is empty, which is acceptable")
	}
}

// TestGetPlugin_Basic tests basic plugin creation functionality
func TestGetPlugin_Basic(t *testing.T) {
	// Create a minimal config
	config := common.Config{}

	// Test plugin creation
	plugin, err := GetPlugin(config)

	// Basic assertions
	if err != nil {
		t.Fatalf("Expected no error, got %v", err)
	}

	if plugin == nil {
		t.Fatal("Expected plugin to be created, got nil")
	}

	// Verify basic plugin properties
	if plugin.ID != "altibase" {
		t.Errorf("Expected plugin ID 'altibase', got '%s'", plugin.ID)
	}

	if plugin.Name != "Altibase" {
		t.Errorf("Expected plugin name 'Altibase', got '%s'", plugin.Name)
	}
}

// TestGetPlugin_WithConfig tests plugin creation with specific configuration
func TestGetPlugin_WithConfig(t *testing.T) {
	// Create config with shared library settings
	config := common.Config{
		SharedLibrary: common.SharedLibraryConfig{
			Altibase: map[int]struct {
				Driver string `mapstructure:"driver"`
			}{
				7: {Driver: "/usr/lib/altibase7.so"},
			},
		},
	}

	plugin, err := GetPlugin(config)

	if err != nil {
		t.Fatalf("Expected no error with config, got %v", err)
	}

	if plugin == nil {
		t.Fatal("Expected plugin to be created with config")
	}

	// Verify datasource client is created
	if plugin.DatasourceClient == nil {
		t.Error("Expected datasource client to be set")
	}
}

// TestGetPlugin_DatasourceClient tests the datasource client setup
func TestGetPlugin_DatasourceClient(t *testing.T) {
	config := common.Config{}
	plugin, err := GetPlugin(config)

	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	// Verify datasource client exists
	if plugin.DatasourceClient == nil {
		t.Fatal("Expected datasource client to be set")
	}

	if plugin.DatasourceClient.DataClient == nil {
		t.Error("Expected data client to be configured")
	}
}

// TestGetPlugin_MultipleCreation tests creating multiple plugin instances
func TestGetPlugin_MultipleCreation(t *testing.T) {
	config := common.Config{}

	// Create multiple plugins
	plugin1, err1 := GetPlugin(config)
	plugin2, err2 := GetPlugin(config)

	if err1 != nil || err2 != nil {
		t.Fatalf("Failed to create plugins: %v, %v", err1, err2)
	}

	// Both should be valid but independent instances
	if plugin1 == plugin2 {
		t.Error("Expected different plugin instances, got same reference")
	}

	// Should have same basic properties
	if plugin1.ID != plugin2.ID {
		t.Error("Expected same plugin ID for both instances")
	}

	if plugin1.Name != plugin2.Name {
		t.Error("Expected same plugin name for both instances")
	}
}

// TestGetPlugin_EmptyConfigHandling tests plugin creation with completely empty config
func TestGetPlugin_EmptyConfigHandling(t *testing.T) {
	config := common.Config{}

	plugin, err := GetPlugin(config)

	// Should create plugin successfully even with empty config
	if err != nil {
		t.Fatalf("Expected no error with empty config, got %v", err)
	}

	if plugin == nil {
		t.Fatal("Expected plugin to be created even with empty config")
	}

	// Verify essential properties are set
	if plugin.ID == "" {
		t.Error("Expected plugin ID to be set")
	}

	if plugin.Name == "" {
		t.Error("Expected plugin name to be set")
	}

	if plugin.SampleFormData == "" {
		t.Error("Expected sample form data to be set")
	}
}

// BenchmarkGetPlugin benchmarks the plugin creation performance
func BenchmarkGetPlugin(b *testing.B) {
	config := common.Config{
		SharedLibrary: common.SharedLibraryConfig{
			Altibase: map[int]struct {
				Driver string `mapstructure:"driver"`
			}{
				7: {Driver: "/usr/lib/altibase7.so"},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := GetPlugin(config)
		if err != nil {
			b.Fatalf("Benchmark failed: %v", err)
		}
	}
}

// TestGetPlugin_PluginTypeDatasource tests that plugin type is consistently set to datasource
func TestGetPlugin_PluginTypeDatasource(t *testing.T) {
	config := common.Config{}
	plugin, err := GetPlugin(config)

	if err != nil {
		t.Fatalf("Failed to create plugin: %v", err)
	}

	// Use string comparison since we may not have access to the constant
	if string(plugin.Type) != "datasource" {
		t.Errorf("Expected plugin type to be 'datasource', got '%s'", plugin.Type)
	}
}
