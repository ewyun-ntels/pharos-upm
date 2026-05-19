package http_receiver

import (
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

	if plugin.ID != "http-receiver" {
		t.Errorf("Expected plugin ID 'http-receiver', got '%s'", plugin.ID)
	}

	if plugin.Name != "HTTP Receiver" {
		t.Errorf("Expected plugin name 'HTTP Receiver', got '%s'", plugin.Name)
	}

	// Test datasource client is properly initialized
	if plugin.DatasourceClient == nil {
		t.Fatal("DatasourceClient is nil")
	}

	if plugin.DatasourceClient.StreamClient == nil {
		t.Fatal("StreamClient is nil")
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

	if typeProperty["default"] != "http-receiver" {
		t.Errorf("Expected type default 'http-receiver', got %v", typeProperty["default"])
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

	// Test port property
	portProperty, ok := dataProperties["port"].(map[string]any)
	if !ok {
		t.Fatal("Port property is not an object")
	}

	if portProperty["type"] != "integer" {
		t.Errorf("Expected port type 'integer', got %v", portProperty["type"])
	}

	// Test uris property
	urisProperty, ok := dataProperties["uris"].(map[string]any)
	if !ok {
		t.Fatal("URIs property is not an object")
	}

	if urisProperty["type"] != "array" {
		t.Errorf("Expected uris type 'array', got %v", urisProperty["type"])
	}
}

func TestEmbeddedUISchema(t *testing.T) {
	// Test that embedded UI schema is valid JSON
	var schema map[string]any
	err := json.Unmarshal([]byte(uiSchema), &schema)
	if err != nil {
		t.Fatalf("UI schema is not valid JSON: %v", err)
	}

	// Test UI schema structure for checkboxes widget
	data, ok := schema["data"].(map[string]any)
	if !ok {
		t.Fatal("Data field is not an object in UI schema")
	}

	uris, ok := data["uris"].(map[string]any)
	if !ok {
		t.Fatal("URIs field is not an object in UI schema")
	}

	items, ok := uris["items"].(map[string]any)
	if !ok {
		t.Fatal("Items field is not an object in UI schema")
	}

	method, ok := items["method"].(map[string]any)
	if !ok {
		t.Fatal("Method field is not an object in UI schema")
	}

	widget, ok := method["ui:widget"].(string)
	if !ok {
		t.Fatal("ui:widget is not a string")
	}

	if widget != "checkboxes" {
		t.Errorf("Expected ui:widget 'checkboxes', got '%s'", widget)
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
	if data["type"] != "http-receiver" {
		t.Errorf("Expected type 'http-receiver', got %v", data["type"])
	}

	if data["name"] != "sample-http-receiver" {
		t.Errorf("Expected name 'sample-http-receiver', got %v", data["name"])
	}

	// Test data field
	dataField, ok := data["data"].(map[string]any)
	if !ok {
		t.Fatal("Data field is not an object")
	}

	// Test port
	port, ok := dataField["port"].(float64)
	if !ok {
		t.Fatal("Port is not a number")
	}

	if port != 10000 {
		t.Errorf("Expected port 10000, got %v", port)
	}

	// Test uris
	uris, ok := dataField["uris"].([]any)
	if !ok {
		t.Fatal("URIs is not an array")
	}

	if len(uris) != 1 {
		t.Errorf("Expected 1 URI, got %d", len(uris))
	}

	// Test first URI
	firstURI, ok := uris[0].(map[string]any)
	if !ok {
		t.Fatal("First URI is not an object")
	}

	if firstURI["uri"] != "receive" {
		t.Errorf("Expected URI 'receive', got %v", firstURI["uri"])
	}

	methods, ok := firstURI["method"].([]any)
	if !ok {
		t.Fatal("Methods is not an array")
	}

	expectedMethods := []string{"GET", "POST"}
	if len(methods) != len(expectedMethods) {
		t.Errorf("Expected %d methods, got %d", len(expectedMethods), len(methods))
	}

	for i, method := range methods {
		if method != expectedMethods[i] {
			t.Errorf("Expected method '%s', got '%v'", expectedMethods[i], method)
		}
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

	// Test that schemas match embedded content
	// Note: JSONSchemaForJson and UiSchemaForJson are OrderedMap types
	// We verify they exist and are not nil
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

func TestHTTPReceiverConstants(t *testing.T) {
	expectedType := "http-receiver"
	expectedName := "HTTP Receiver"

	config := common.Config{}
	plugin, _ := GetPlugin(config)

	if plugin.ID != expectedType {
		t.Errorf("Expected plugin ID '%s', got '%s'", expectedType, plugin.ID)
	}

	if plugin.Name != expectedName {
		t.Errorf("Expected plugin name '%s', got '%s'", expectedName, plugin.Name)
	}
}

func TestStreamClientInitialization(t *testing.T) {
	config := common.Config{}
	plugin, err := GetPlugin(config)
	if err != nil {
		t.Fatalf("GetPlugin() returned error: %v", err)
	}

	// Verify StreamClient is properly wrapped
	streamClient := plugin.DatasourceClient.StreamClient
	if streamClient == nil {
		t.Fatal("StreamClient is nil")
	}

	// Test that the StreamClient interface is properly implemented
	// by attempting to call methods (they should not panic)
	subscribeReq := &model.SubscribeStreamRequest{}
	subscribeResp := streamClient.SubscribeStream(subscribeReq)
	if subscribeResp == nil {
		t.Error("SubscribeStream returned nil response")
	}

	unsubscribeReq := &model.UnsubscribeStreamRequest{}
	err = streamClient.UnsubscribeStream(unsubscribeReq)
	if err != nil {
		t.Errorf("UnsubscribeStream returned error: %v", err)
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
