package model

import (
	"strings"
	"testing"

	"github.com/iancoleman/orderedmap"
)

// Test data constants
const (
	basicJSONSchema = `{
		"type": "object",
		"properties": {
			"name": {
				"type": "string"
			},
			"age": {
				"type": "number"
			}
		},
		"required": ["name"]
	}`

	complexJSONSchema = `{
		"type": "object",
		"properties": {
			"connection": {
				"type": "object",
				"properties": {
					"host": {
						"type": "string",
						"pattern": "^[a-zA-Z0-9.-]+$"
					},
					"port": {
						"type": "integer",
						"minimum": 1,
						"maximum": 65535
					},
					"ssl": {
						"type": "boolean"
					}
				},
				"required": ["host", "port"]
			},
			"authentication": {
				"type": "object",
				"properties": {
					"username": {
						"type": "string",
						"minLength": 1
					},
					"password": {
						"type": "string",
						"minLength": 1
					}
				},
				"required": ["username", "password"]
			}
		},
		"required": ["connection", "authentication"]
	}`

	postgresJSONSchema = `{
		"type": "object",
		"properties": {
			"host": {
				"type": "string",
				"description": "Database host"
			},
			"port": {
				"type": "integer",
				"minimum": 1,
				"maximum": 65535
			}
		},
		"required": ["host"]
	}`

	postgresUISchema = `{
		"host": {
			"ui:placeholder": "localhost"
		},
		"port": {
			"ui:widget": "updown"
		}
	}`

	htmlUISchema = `{
		"description": {
			"ui:help": "<b>Bold text</b>"
		}
	}`

	validBasicData    = `{"name": "Alice", "age": 25}`
	validBasicData2   = `{"name": "John", "age": 30}`
	invalidBasicData  = `{"age": 25}`
	malformedJSON     = `{"name": "John",}`
	validPostgresData = `{"host": "localhost", "port": 5432}`
	invalidSampleData = `{"port": 5432}`
	validComplexData  = `{
		"connection": {
			"host": "db.example.com",
			"port": 3306,
			"ssl": false
		},
		"authentication": {
			"username": "user123",
			"password": "password123"
		}
	}`
	invalidComplexData = `{
		"connection": {
			"host": "db.example.com",
			"port": 70000,
			"ssl": false
		},
		"authentication": {
			"username": "user123",
			"password": "password123"
		}
	}`
	complexSampleData = `{
		"connection": {
			"host": "localhost",
			"port": 5432,
			"ssl": true
		},
		"authentication": {
			"username": "admin",
			"password": "secret"
		}
	}`
)

// Helper functions
func createBasicPlugin(t *testing.T) *Plugin {
	t.Helper()
	plugin, err := MakePlugin(
		PluginTypeDatasource,
		"test-plugin",
		"Test Plugin",
		basicJSONSchema,
		"{}",
		nil,
		validBasicData2,
	)
	if err != nil {
		t.Fatalf("Failed to create basic plugin: %v", err)
	}
	return plugin
}

func createPluginWithSchema(t *testing.T, id, name, jsonSchema, uiSchema, sampleData string) *Plugin {
	t.Helper()
	plugin, err := MakePlugin(
		PluginTypeDatasource,
		id,
		name,
		jsonSchema,
		uiSchema,
		nil,
		sampleData,
	)
	if err != nil {
		t.Fatalf("Failed to create plugin %s: %v", id, err)
	}
	return plugin
}

func assertPluginFields(t *testing.T, plugin *Plugin, expectedType PluginType, expectedID, expectedName string) {
	t.Helper()
	if plugin.Type != expectedType {
		t.Errorf("Expected type %s, got %s", expectedType, plugin.Type)
	}
	if plugin.ID != expectedID {
		t.Errorf("Expected ID '%s', got '%s'", expectedID, plugin.ID)
	}
	if plugin.Name != expectedName {
		t.Errorf("Expected name '%s', got '%s'", expectedName, plugin.Name)
	}
}

func TestPlugin_ValidateJSONSchemaInstance(t *testing.T) {
	plugin := createBasicPlugin(t)

	tests := []struct {
		name        string
		instance    string
		expectError bool
		description string
	}{
		{
			name:        "Valid instance",
			instance:    validBasicData,
			expectError: false,
			description: "Should pass validation with valid data",
		},
		{
			name:        "Invalid instance - missing required field",
			instance:    invalidBasicData,
			expectError: true,
			description: "Should fail validation when required field is missing",
		},
		{
			name:        "Invalid JSON format",
			instance:    malformedJSON,
			expectError: true,
			description: "Should fail validation with malformed JSON",
		},
		{
			name:        "Empty object",
			instance:    "{}",
			expectError: true,
			description: "Should fail validation when required fields are missing",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := plugin.ValidateJSONSchemaInstance(tt.instance)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s, but validation passed", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected validation to pass for %s, got error: %v", tt.description, err)
			}
		})
	}
}

func TestMakePlugin(t *testing.T) {
	t.Run("Success", func(t *testing.T) {
		plugin := createPluginWithSchema(t,
			"postgresql-plugin",
			"PostgreSQL Plugin",
			postgresJSONSchema,
			postgresUISchema,
			validPostgresData,
		)

		// 기본 속성 검증
		assertPluginFields(t, plugin, PluginTypeDatasource, "postgresql-plugin", "PostgreSQL Plugin")

		if plugin.SampleFormData != validPostgresData {
			t.Errorf("Expected sample form data %s, got %s", validPostgresData, plugin.SampleFormData)
		}

		// JSON Schema 검증
		if plugin.Settings.JSONSchema == nil {
			t.Error("Expected JSON schema to be compiled")
		}

		if plugin.Settings.JSONSchemaForJson == nil {
			t.Error("Expected JSON schema for JSON to be set")
		}

		if plugin.Settings.UiSchemaForJson == nil {
			t.Error("Expected UI schema for JSON to be set")
		}
	})

	t.Run("Error cases", func(t *testing.T) {
		errorTests := []struct {
			name           string
			jsonSchema     string
			uiSchema       string
			sampleData     string
			expectedErrMsg string
		}{
			{
				name: "Invalid JSON Schema",
				jsonSchema: `{
					"type": "object",
					"properties": {
						"host": {
							"type": "invalid-type"
						}
					}
				}`,
				uiSchema:   "{}",
				sampleData: `{"host": "localhost"}`,
			},
			{
				name:       "Invalid UI Schema",
				jsonSchema: postgresJSONSchema,
				uiSchema: `{
					"host": {
						"ui:widget": "invalid
					}
				}`,
				sampleData: `{"host": "localhost"}`,
			},
			{
				name:       "Invalid sample form data",
				jsonSchema: postgresJSONSchema,
				uiSchema:   "{}",
				sampleData: invalidSampleData,
			},
		}

		for _, tt := range errorTests {
			t.Run(tt.name, func(t *testing.T) {
				_, err := MakePlugin(
					PluginTypeDatasource,
					"test-plugin",
					"Test Plugin",
					tt.jsonSchema,
					tt.uiSchema,
					nil,
					tt.sampleData,
				)

				if err == nil {
					t.Errorf("Expected error for %s", tt.name)
				}
			})
		}
	})

	t.Run("HTML Escape Disabled", func(t *testing.T) {
		jsonSchemaWithDesc := `{
			"type": "object",
			"properties": {
				"description": {
					"type": "string"
				}
			}
		}`

		plugin := createPluginWithSchema(t,
			"test-plugin",
			"Test Plugin",
			jsonSchemaWithDesc,
			htmlUISchema,
			`{"description": "test"}`,
		)

		// HTML 이스케이프가 비활성화되어 있는지 확인
		jsonBytes, err := plugin.Settings.UiSchemaForJson.MarshalJSON()
		if err != nil {
			t.Fatalf("Failed to marshal UI schema: %v", err)
		}

		jsonStr := string(jsonBytes)
		if !strings.Contains(jsonStr, "<b>Bold text</b>") {
			t.Error("Expected HTML tags to be preserved (not escaped)")
		}
	})

	t.Run("Empty schemas", func(t *testing.T) {
		plugin := createPluginWithSchema(t,
			"empty-plugin",
			"Empty Plugin",
			`{}`,
			`{}`,
			`{}`,
		)

		// 빈 스키마로도 유효한 플러그인이 생성되어야 함
		if plugin.Settings.JSONSchema == nil {
			t.Error("Expected JSON schema to be compiled even if empty")
		}

		// 빈 객체는 빈 스키마에 대해 유효해야 함
		err := plugin.ValidateJSONSchemaInstance(`{}`)
		if err != nil {
			t.Errorf("Expected empty object to be valid against empty schema, got error: %v", err)
		}
	})
}

func TestPlugin_ComplexJSONSchema(t *testing.T) {
	plugin := createPluginWithSchema(t,
		"complex-plugin",
		"Complex Plugin",
		complexJSONSchema,
		"{}",
		complexSampleData,
	)

	validationTests := []struct {
		name        string
		data        string
		expectError bool
		description string
	}{
		{
			name:        "Valid complex data",
			data:        validComplexData,
			expectError: false,
			description: "Should pass validation with valid complex data",
		},
		{
			name:        "Invalid complex data - port out of range",
			data:        invalidComplexData,
			expectError: true,
			description: "Should fail validation when port is out of valid range",
		},
		{
			name: "Missing authentication",
			data: `{
				"connection": {
					"host": "localhost",
					"port": 5432,
					"ssl": true
				}
			}`,
			expectError: true,
			description: "Should fail validation when required authentication is missing",
		},
		{
			name: "Invalid host pattern",
			data: `{
				"connection": {
					"host": "invalid_host!",
					"port": 5432,
					"ssl": true
				},
				"authentication": {
					"username": "user",
					"password": "pass"
				}
			}`,
			expectError: true,
			description: "Should fail validation when host doesn't match pattern",
		},
	}

	for _, tt := range validationTests {
		t.Run(tt.name, func(t *testing.T) {
			err := plugin.ValidateJSONSchemaInstance(tt.data)

			if tt.expectError && err == nil {
				t.Errorf("Expected error for %s, but validation passed", tt.description)
			}
			if !tt.expectError && err != nil {
				t.Errorf("Expected validation to pass for %s, got error: %v", tt.description, err)
			}
		})
	}
}

func TestSettings_JSONSchemaCompilation(t *testing.T) {
	settings := Settings{
		JSONSchemaForJson: orderedmap.New(),
		UiSchemaForJson:   orderedmap.New(),
	}

	// OrderedMap이 제대로 초기화되었는지 확인
	if settings.JSONSchemaForJson == nil {
		t.Error("Expected JSONSchemaForJson to be initialized")
	}

	if settings.UiSchemaForJson == nil {
		t.Error("Expected UiSchemaForJson to be initialized")
	}
}

func TestPluginType_String(t *testing.T) {
	tests := []struct {
		name        string
		pluginType  PluginType
		expectedStr string
	}{
		{
			name:        "Datasource plugin type",
			pluginType:  PluginTypeDatasource,
			expectedStr: "datasource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.pluginType.String() != tt.expectedStr {
				t.Errorf("Expected %s, got %s", tt.expectedStr, tt.pluginType.String())
			}
		})
	}
}

// Benchmark tests
func BenchmarkPlugin_ValidateJSONSchemaInstance(b *testing.B) {
	plugin, err := MakePlugin(
		PluginTypeDatasource,
		"benchmark-plugin",
		"Benchmark Plugin",
		basicJSONSchema,
		"{}",
		nil,
		validBasicData2,
	)

	if err != nil {
		b.Fatalf("Failed to create plugin: %v", err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = plugin.ValidateJSONSchemaInstance(validBasicData)
	}
}

func BenchmarkMakePlugin(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = MakePlugin(
			PluginTypeDatasource,
			"benchmark-plugin",
			"Benchmark Plugin",
			postgresJSONSchema,
			postgresUISchema,
			nil,
			validPostgresData,
		)
	}
}
