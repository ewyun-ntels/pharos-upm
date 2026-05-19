package model

import (
	"testing"

	"ntels.com/pharos/core/external"
)

// Helper functions for assertions
func assertStringField(t *testing.T, fieldName, actual, expected string) {
	t.Helper()
	if actual != expected {
		t.Errorf("Expected %s to be '%s', got '%s'", fieldName, expected, actual)
	}
}

func assertIntField(t *testing.T, fieldName string, actual, expected int) {
	t.Helper()
	if actual != expected {
		t.Errorf("Expected %s to be %d, got %d", fieldName, expected, actual)
	}
}

func TestPluginTypeConstants(t *testing.T) {
	tests := []struct {
		name     string
		actual   PluginType
		expected string
	}{
		{
			name:     "PluginTypeDatasource constant",
			actual:   PluginTypeDatasource,
			expected: "datasource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.actual != PluginType(tt.expected) {
				t.Errorf("Expected %s to be '%s', got '%s'", tt.name, tt.expected, tt.actual)
			}
		})
	}
}

func TestErrorConstants(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected string
	}{
		{
			name:     "ErrorNotSupportedDatasourceType",
			err:      external.ErrorNotSupportedDatasourceType,
			expected: "not supported datasource type",
		},
		{
			name:     "ErrorNotSupportedDataServer",
			err:      external.ErrorNotSupportedDataServer,
			expected: "not supported data server",
		},
		{
			name:     "ErrorNotSupportedStreamServer",
			err:      external.ErrorNotSupportedStreamServer,
			expected: "not supported stream server",
		},
		{
			name:     "ErrorDatasourceNotExist",
			err:      external.ErrorNotExistDatasource,
			expected: "not exist datasource",
		},
		{
			name:     "ErrorNotSupported",
			err:      external.ErrorNotSupported,
			expected: "not supported",
		},
		{
			name:     "ErrorNotImplemented",
			err:      external.ErrorNotImplemented,
			expected: "not implemented",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.err.Error() != tt.expected {
				t.Errorf("Expected error message '%s', got '%s'", tt.expected, tt.err.Error())
			}
		})
	}
}

// Test data constants
var (
	testDatasource = Datasource{
		Name: "test-datasource",
		Type: "postgresql",
	}

	testStreamDatasource = Datasource{
		Name: "test-stream-datasource",
		Type: "websocket",
	}

	testCentrifugeDatasource = Datasource{
		Name: "test-centrifuge-datasource",
		Type: "centrifuge",
	}

	testHeaders = map[string]string{
		"Authorization": "Bearer token123",
		"Content-Type":  "application/json",
	}
)

func TestStructures(t *testing.T) {
	t.Run("Query", func(t *testing.T) {
		query := Query{
			ID:      "test-query-id",
			SQL:     "SELECT * FROM users WHERE id = ?",
			Timeout: 60,
		}

		assertStringField(t, "ID", query.ID, "test-query-id")
		assertStringField(t, "SQL", query.SQL, "SELECT * FROM users WHERE id = ?")
		assertIntField(t, "Timeout", query.Timeout, 60)
	})

	t.Run("QueryDataResult", func(t *testing.T) {
		result := QueryDataResult{
			Error: "test error message",
		}

		assertStringField(t, "Error", result.Error, "test error message")
	})

	t.Run("QueryDataRequest", func(t *testing.T) {
		queries := []Query{
			{ID: "query1", SQL: "SELECT 1"},
			{ID: "query2", SQL: "SELECT 2"},
		}

		request := QueryDataRequest{
			Datasource: testDatasource,
			Queries:    queries,
		}

		assertStringField(t, "Datasource.Name", request.Datasource.Name, "test-datasource")
		assertIntField(t, "Queries length", len(request.Queries), 2)
		assertStringField(t, "First query ID", request.Queries[0].ID, "query1")
	})

	t.Run("QueryDataResponse", func(t *testing.T) {
		response := QueryDataResponse{
			Results: map[string]QueryDataResult{
				"query1": {Error: ""},
				"query2": {Error: "connection error"},
			},
		}

		assertIntField(t, "Results length", len(response.Results), 2)

		result1, exists := response.Results["query1"]
		if !exists {
			t.Error("Expected query1 result to exist")
		}
		assertStringField(t, "Query1 error", result1.Error, "")

		result2, exists := response.Results["query2"]
		if !exists {
			t.Error("Expected query2 result to exist")
		}
		assertStringField(t, "Query2 error", result2.Error, "connection error")
	})

	t.Run("RemoveDatasourceRequest", func(t *testing.T) {
		request := RemoveDatasourceRequest{
			Datasource: testDatasource,
		}

		assertStringField(t, "Datasource.Name", request.Datasource.Name, "test-datasource")
		assertStringField(t, "Datasource.Type", request.Datasource.Type, "postgresql")
	})
}

func TestStreamStructures(t *testing.T) {
	t.Run("RunStreamRequest", func(t *testing.T) {
		request := RunStreamRequest{
			Datasource: testStreamDatasource,
			Path:       "/api/stream",
			Data:       `{"message": "hello"}`,
			Headers:    testHeaders,
		}

		assertStringField(t, "Datasource.Name", request.Datasource.Name, "test-stream-datasource")
		assertStringField(t, "Path", request.Path, "/api/stream")
		assertStringField(t, "Data", request.Data, `{"message": "hello"}`)
		assertIntField(t, "Headers length", len(request.Headers), 2)
		assertStringField(t, "Authorization header", request.Headers["Authorization"], "Bearer token123")
	})

	t.Run("RunStreamResponse", func(t *testing.T) {
		// Success case
		response := RunStreamResponse{
			Data:  `{"status": "connected"}`,
			Error: "",
		}

		assertStringField(t, "Data", response.Data, `{"status": "connected"}`)
		assertStringField(t, "Error", response.Error, "")

		// Error case
		errorResponse := RunStreamResponse{
			Data:  "",
			Error: "connection failed",
		}

		assertStringField(t, "Data", errorResponse.Data, "")
		assertStringField(t, "Error", errorResponse.Error, "connection failed")
	})

	t.Run("SubscribeStreamRequest", func(t *testing.T) {
		request := SubscribeStreamRequest{
			Datasource: testCentrifugeDatasource,
			Client:     nil, // centrifuge.Client는 테스트에서 nil로 설정
			Path:       "/subscribe/channel1",
			Data:       `{"channel": "notifications"}`,
			Headers: map[string]string{
				"X-Client-ID": "client123",
			},
		}

		assertStringField(t, "Datasource.Name", request.Datasource.Name, "test-centrifuge-datasource")
		assertStringField(t, "Path", request.Path, "/subscribe/channel1")
		assertStringField(t, "Data", request.Data, `{"channel": "notifications"}`)
		assertStringField(t, "X-Client-ID header", request.Headers["X-Client-ID"], "client123")

		// Verify Client field is nil as expected in test
		if request.Client != nil {
			t.Error("Expected Client to be nil in test environment")
		}
	})

	t.Run("SubscribeStreamResponse", func(t *testing.T) {
		response := SubscribeStreamResponse{
			Data:  `{"subscribed": true, "channel": "notifications"}`,
			Error: "",
		}

		assertStringField(t, "Data", response.Data, `{"subscribed": true, "channel": "notifications"}`)
		assertStringField(t, "Error", response.Error, "")
	})

	t.Run("UnsubscribeStreamRequest", func(t *testing.T) {
		request := UnsubscribeStreamRequest{
			Datasource: testCentrifugeDatasource,
			Path:       "/unsubscribe/channel1",
			Data:       `{"channel": "notifications"}`,
			Headers: map[string]string{
				"X-Session-ID": "session456",
			},
		}

		assertStringField(t, "Datasource.Name", request.Datasource.Name, "test-centrifuge-datasource")
		assertStringField(t, "Path", request.Path, "/unsubscribe/channel1")
		assertStringField(t, "Data", request.Data, `{"channel": "notifications"}`)
		assertStringField(t, "X-Session-ID header", request.Headers["X-Session-ID"], "session456")
	})

	t.Run("PublishStreamRequest", func(t *testing.T) {
		publishHeaders := map[string]string{
			"X-Publisher-ID": "publisher789",
			"Content-Type":   "application/json",
		}

		request := PublishStreamRequest{
			Datasource: testCentrifugeDatasource,
			Path:       "/publish/channel1",
			Data:       `{"message": "Hello World", "timestamp": "2023-01-01T00:00:00Z"}`,
			Headers:    publishHeaders,
		}

		assertStringField(t, "Datasource.Name", request.Datasource.Name, "test-centrifuge-datasource")
		assertStringField(t, "Path", request.Path, "/publish/channel1")

		expectedData := `{"message": "Hello World", "timestamp": "2023-01-01T00:00:00Z"}`
		assertStringField(t, "Data", request.Data, expectedData)

		assertIntField(t, "Headers length", len(request.Headers), 2)
		assertStringField(t, "X-Publisher-ID header", request.Headers["X-Publisher-ID"], "publisher789")
		assertStringField(t, "Content-Type header", request.Headers["Content-Type"], "application/json")
	})

	t.Run("PublishStreamResponse", func(t *testing.T) {
		// Success case
		response := PublishStreamResponse{
			Data:  `{"published": true, "messageId": "msg123"}`,
			Error: "",
		}

		assertStringField(t, "Data", response.Data, `{"published": true, "messageId": "msg123"}`)
		assertStringField(t, "Error", response.Error, "")

		// Error case
		errorResponse := PublishStreamResponse{
			Data:  "",
			Error: "publish failed: channel not found",
		}

		assertStringField(t, "Data", errorResponse.Data, "")
		assertStringField(t, "Error", errorResponse.Error, "publish failed: channel not found")
	})
}
