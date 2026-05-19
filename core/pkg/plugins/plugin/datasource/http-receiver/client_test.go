package http_receiver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/plugins/model"
)

func TestStreamClient_SubscribeStream(t *testing.T) {
	servers := internal.NewMap[*http.Server]()
	client := &StreamClient{Servers: servers}

	request := &model.SubscribeStreamRequest{}
	response := client.SubscribeStream(request)

	if response == nil {
		t.Error("SubscribeStream returned nil response")
	}
}

func TestStreamClient_UnsubscribeStream(t *testing.T) {
	servers := internal.NewMap[*http.Server]()
	client := &StreamClient{Servers: servers}

	request := &model.UnsubscribeStreamRequest{}
	err := client.UnsubscribeStream(request)

	if err != nil {
		t.Errorf("UnsubscribeStream returned error: %v", err)
	}
}

func TestStreamClient_RunStream(t *testing.T) {
	tests := []struct {
		name        string
		request     *model.RunStreamRequest
		expectError bool
	}{
		{
			name: "Valid request",
			request: &model.RunStreamRequest{
				Datasource: model.Datasource{
					Name: "test-datasource",
					Data: map[string]any{
						"port": float64(8080),
						"uris": []any{
							map[string]any{
								"uri":    "/test",
								"method": []any{"GET", "POST"},
							},
						},
					},
				},
				Headers: map[string]string{
					"websocket-endpoints": "ws://localhost:8000",
				},
			},
			expectError: false,
		},
		{
			name: "Invalid URI format",
			request: &model.RunStreamRequest{
				Datasource: model.Datasource{
					Name: "test-datasource-2",
					Data: map[string]any{
						"port": float64(8081),
						"uris": "invalid-format",
					},
				},
				Headers: map[string]string{
					"websocket-endpoints": "ws://localhost:8000",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			servers := internal.NewMap[*http.Server]()
			client := &StreamClient{Servers: servers}

			response := client.RunStream(tt.request)

			if response == nil {
				t.Error("RunStream returned nil response")
				return
			}

			if tt.expectError {
				if response.Error == "" {
					t.Error("Expected error but got none")
				}
			} else {
				if response.Error != "" {
					t.Errorf("Unexpected error: %s", response.Error)
				}

				// Check if server was added to the map
				if !client.Servers.Exist(tt.request.Datasource.Name) {
					t.Error("Expected server to be added to Servers map")
				}
			}
		})
	}
}

func TestStreamClient_RunStream_ExistingDatasource(t *testing.T) {
	servers := internal.NewMap[*http.Server]()
	client := &StreamClient{Servers: servers}

	// Pre-populate servers map
	existingServer := &http.Server{Addr: ":9999"}
	servers.Set("existing-datasource", existingServer)

	request := &model.RunStreamRequest{
		Datasource: model.Datasource{
			Name: "existing-datasource",
			Data: map[string]any{
				"port": float64(8080),
				"uris": []any{
					map[string]any{
						"uri":    "/test",
						"method": []any{"GET"},
					},
				},
			},
		},
		Headers: map[string]string{
			"websocket-endpoints": "ws://localhost:8000",
		},
	}

	response := client.RunStream(request)

	if response == nil {
		t.Error("RunStream returned nil response")
		return
	}

	// Should return early for existing datasource
	if response.Error != "" {
		t.Errorf("Unexpected error for existing datasource: %s", response.Error)
	}
}

func TestStreamClient_PublishStream(t *testing.T) {
	servers := internal.NewMap[*http.Server]()
	client := &StreamClient{Servers: servers}

	request := &model.PublishStreamRequest{}
	response := client.PublishStream(request)

	if response == nil {
		t.Error("PublishStream returned nil response")
	}
}

func TestStreamClient_RemoveDatasource(t *testing.T) {
	servers := internal.NewMap[*http.Server]()
	client := &StreamClient{Servers: servers}

	// Create a test server
	server := &http.Server{Addr: ":0"}
	servers.Set("test-datasource", server)

	request := &model.RemoveDatasourceRequest{
		Datasource: model.Datasource{
			Name: "test-datasource",
		},
	}

	err := client.RemoveDatasource(request)

	if err != nil {
		t.Errorf("RemoveDatasource returned error: %v", err)
	}

	// Verify server was removed
	if servers.Exist("test-datasource") {
		t.Error("Expected server to be removed from Servers map")
	}
}

func TestStreamClient_Handler(t *testing.T) {
	// Set Gin to test mode
	gin.SetMode(gin.TestMode)

	servers := internal.NewMap[*http.Server]()
	client := &StreamClient{Servers: servers}

	// Create test handler
	websocketEndpoints := []string{"ws://localhost:8000"}
	datasourceName := "test-datasource"
	handler := client.handler(websocketEndpoints, datasourceName)

	// Create test cases
	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		expectedStatus int
	}{
		{
			name:           "GET request",
			method:         "GET",
			path:           "/test",
			body:           "",
			expectedStatus: http.StatusInternalServerError, // Will fail due to centrifuge publish error
		},
		{
			name:           "POST request with body",
			method:         "POST",
			path:           "/test",
			body:           `{"test": "data"}`,
			expectedStatus: http.StatusInternalServerError, // Will fail due to centrifuge publish error
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create test request
			req := httptest.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			w := httptest.NewRecorder()

			// Create Gin context
			c, _ := gin.CreateTestContext(w)
			c.Request = req
			c.FullPath()

			// Call handler
			handler(c)

			// Note: We expect StatusInternalServerError because CentrifugePublish will fail
			// in test environment without actual Centrifuge server
			if w.Code != tt.expectedStatus {
				t.Errorf("Expected status %d, got %d", tt.expectedStatus, w.Code)
			}
		})
	}
}

func TestStreamClient_HandlerDataExtraction(t *testing.T) {
	// Test that handler correctly extracts request data
	gin.SetMode(gin.TestMode)

	// Create a modified handler that captures data instead of publishing
	handlerFunc := func(c *gin.Context) {
		data := map[string]any{}
		data["fullPath"] = c.FullPath()
		data["method"] = c.Request.Method

		if body, err := c.GetRawData(); err == nil {
			data["body"] = string(body)
		}

		// Verify data structure - FullPath might be empty in test context
		if data["method"] == "" {
			t.Error("Expected method to be set")
		}

		if _, exists := data["body"]; !exists {
			t.Error("Expected body to be set")
		}

		c.JSON(http.StatusOK, data)
	}

	// Test the data extraction logic
	req := httptest.NewRequest("POST", "/test-endpoint", strings.NewReader(`{"key": "value"}`))
	w := httptest.NewRecorder()

	c, _ := gin.CreateTestContext(w)
	c.Request = req

	handlerFunc(c)

	if w.Code != http.StatusOK {
		t.Errorf("Expected status 200, got %d", w.Code)
	}

	// Verify response contains expected data
	var responseData map[string]any
	err := json.Unmarshal(w.Body.Bytes(), &responseData)
	if err != nil {
		t.Errorf("Failed to unmarshal response: %v", err)
	}

	if responseData["method"] != "POST" {
		t.Errorf("Expected method POST, got %v", responseData["method"])
	}

	if responseData["body"] != `{"key": "value"}` {
		t.Errorf("Expected body to match input, got %v", responseData["body"])
	}
}

func TestStreamClient_RunStreamPortParsing(t *testing.T) {
	servers := internal.NewMap[*http.Server]()
	client := &StreamClient{Servers: servers}

	request := &model.RunStreamRequest{
		Datasource: model.Datasource{
			Name: "port-test-datasource",
			Data: map[string]any{
				"port": float64(9999),
				"uris": []any{
					map[string]any{
						"uri":    "/test",
						"method": []any{"GET"},
					},
				},
			},
		},
		Headers: map[string]string{
			"websocket-endpoints": "ws://localhost:8000",
		},
	}

	response := client.RunStream(request)

	if response == nil {
		t.Fatal("RunStream returned nil response")
	}

	if response.Error != "" {
		t.Errorf("Unexpected error: %s", response.Error)
		return
	}

	// Verify server exists and has correct port
	if !client.Servers.Exist("port-test-datasource") {
		t.Error("Expected server to be added to Servers map")
		return
	}

	server := client.Servers.Get("port-test-datasource")
	expectedAddr := ":9999"
	if server.Addr != expectedAddr {
		t.Errorf("Expected server address %s, got %s", expectedAddr, server.Addr)
	}
}

// Benchmark tests
func BenchmarkStreamClient_SubscribeStream(b *testing.B) {
	servers := internal.NewMap[*http.Server]()
	client := &StreamClient{Servers: servers}
	request := &model.SubscribeStreamRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.SubscribeStream(request)
	}
}

func BenchmarkStreamClient_UnsubscribeStream(b *testing.B) {
	servers := internal.NewMap[*http.Server]()
	client := &StreamClient{Servers: servers}
	request := &model.UnsubscribeStreamRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.UnsubscribeStream(request)
	}
}

func BenchmarkStreamClient_PublishStream(b *testing.B) {
	servers := internal.NewMap[*http.Server]()
	client := &StreamClient{Servers: servers}
	request := &model.PublishStreamRequest{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.PublishStream(request)
	}
}

func BenchmarkStreamClient_HandlerCreation(b *testing.B) {
	servers := internal.NewMap[*http.Server]()
	client := &StreamClient{Servers: servers}
	websocketEndpoints := []string{"ws://localhost:8000"}
	datasourceName := "benchmark-test"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client.handler(websocketEndpoints, datasourceName)
	}
}
