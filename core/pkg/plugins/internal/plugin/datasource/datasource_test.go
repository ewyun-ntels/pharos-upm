package datasource

import (
	"context"
	"testing"

	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/internal/plugin/datasource/hashicorp"
	"ntels.com/pharos/core/pkg/plugins/model"
)

// Mock 구현체들
type mockBuiltInDataClient struct {
	queryDataResponse   *model.QueryDataResponse
	removeDatasourceErr error
}

func (m *mockBuiltInDataClient) QueryData(_ context.Context, request *model.QueryDataRequest) *model.QueryDataResponse {
	return m.queryDataResponse
}

func (m *mockBuiltInDataClient) RemoveDatasource(request *model.RemoveDatasourceRequest) error {
	return m.removeDatasourceErr
}

type mockBuiltInStreamClient struct {
	runStreamResponse       *model.RunStreamResponse
	subscribeStreamResponse *model.SubscribeStreamResponse
	unsubscribeStreamErr    error
	publishStreamResponse   *model.PublishStreamResponse
	removeDatasourceErr     error
}

func (m *mockBuiltInStreamClient) RunStream(request *model.RunStreamRequest) *model.RunStreamResponse {
	return m.runStreamResponse
}

func (m *mockBuiltInStreamClient) SubscribeStream(request *model.SubscribeStreamRequest) *model.SubscribeStreamResponse {
	return m.subscribeStreamResponse
}

func (m *mockBuiltInStreamClient) UnsubscribeStream(request *model.UnsubscribeStreamRequest) error {
	return m.unsubscribeStreamErr
}

func (m *mockBuiltInStreamClient) PublishStream(request *model.PublishStreamRequest) *model.PublishStreamResponse {
	return m.publishStreamResponse
}

func (m *mockBuiltInStreamClient) RemoveDatasource(request *model.RemoveDatasourceRequest) error {
	return m.removeDatasourceErr
}

func TestGetDataClient_BuiltIn(t *testing.T) {
	// Arrange
	mockBuiltIn := &mockBuiltInDataClient{
		queryDataResponse: &model.QueryDataResponse{
			Results: map[string]model.QueryDataResult{
				"test": {Error: "built-in response"},
			},
		},
	}
	config := common.Config{}

	// Act
	client := GetDataClient(PluginLocationBuiltIn, mockBuiltIn, config)

	// Assert
	if client != mockBuiltIn {
		t.Errorf("Expected built-in client, got different instance")
	}

	// 실제 동작 테스트
	request := &model.QueryDataRequest{
		Queries: []model.Query{{ID: "test", SQL: "SELECT 1"}},
	}
	response := client.QueryData(context.Background(), request)

	if response.Results["test"].Error != "built-in response" {
		t.Errorf("Expected 'built-in response', got %s", response.Results["test"].Error)
	}
}

func TestGetDataClient_External(t *testing.T) {
	// Arrange
	mockBuiltIn := &mockBuiltInDataClient{}
	config := common.Config{}

	// Act
	client := GetDataClient(PluginLocationExternal, mockBuiltIn, config)

	// Assert
	hashicorpClient, ok := client.(*hashicorp.DataClient)
	if !ok {
		t.Errorf("Expected hashicorp.DataClient, got %T", client)
	}

	// Config는 비교할 수 없는 필드가 있으므로 타입만 확인
	_ = hashicorpClient.Config
}

func TestGetDataClient_Default(t *testing.T) {
	// Arrange
	mockBuiltIn := &mockBuiltInDataClient{
		queryDataResponse: &model.QueryDataResponse{
			Results: map[string]model.QueryDataResult{
				"default": {Error: "default response"},
			},
		},
	}
	config := common.Config{}
	invalidLocation := PluginLocation("invalid")

	// Act
	client := GetDataClient(invalidLocation, mockBuiltIn, config)

	// Assert
	if client != mockBuiltIn {
		t.Errorf("Expected built-in client as default, got different instance")
	}

	// 실제 동작 테스트
	request := &model.QueryDataRequest{
		Queries: []model.Query{{ID: "default", SQL: "SELECT 1"}},
	}
	response := client.QueryData(context.Background(), request)

	if response.Results["default"].Error != "default response" {
		t.Errorf("Expected 'default response', got %s", response.Results["default"].Error)
	}
}

func TestGetStreamClient_BuiltIn(t *testing.T) {
	// Arrange
	mockBuiltIn := &mockBuiltInStreamClient{
		runStreamResponse: &model.RunStreamResponse{
			Data:  "built-in stream response",
			Error: "",
		},
	}
	config := common.Config{}

	// Act
	client := GetStreamClient(PluginLocationBuiltIn, mockBuiltIn, config)

	// Assert
	if client != mockBuiltIn {
		t.Errorf("Expected built-in stream client, got different instance")
	}

	// 실제 동작 테스트
	request := &model.RunStreamRequest{
		Path: "/test",
		Data: "test data",
	}
	response := client.RunStream(request)

	if response.Data != "built-in stream response" {
		t.Errorf("Expected 'built-in stream response', got %s", response.Data)
	}
}

func TestGetStreamClient_External(t *testing.T) {
	// Arrange
	mockBuiltIn := &mockBuiltInStreamClient{}
	config := common.Config{}

	// Act
	client := GetStreamClient(PluginLocationExternal, mockBuiltIn, config)

	// Assert
	hashicorpClient, ok := client.(*hashicorp.StreamClient)
	if !ok {
		t.Errorf("Expected hashicorp.StreamClient, got %T", client)
	}

	// Config는 비교할 수 없는 필드가 있으므로 타입만 확인
	_ = hashicorpClient.Config
}

func TestGetStreamClient_Default(t *testing.T) {
	// Arrange
	mockBuiltIn := &mockBuiltInStreamClient{
		subscribeStreamResponse: &model.SubscribeStreamResponse{
			Data:  "default stream response",
			Error: "",
		},
	}
	config := common.Config{}
	invalidLocation := PluginLocation("invalid")

	// Act
	client := GetStreamClient(invalidLocation, mockBuiltIn, config)

	// Assert
	if client != mockBuiltIn {
		t.Errorf("Expected built-in stream client as default, got different instance")
	}

	// 실제 동작 테스트
	request := &model.SubscribeStreamRequest{
		Path: "/subscribe",
		Data: "test data",
	}
	response := client.SubscribeStream(request)

	if response.Data != "default stream response" {
		t.Errorf("Expected 'default stream response', got %s", response.Data)
	}
}

func TestPluginLocation_Constants(t *testing.T) {
	// 상수 값들이 올바르게 정의되어 있는지 테스트
	if PluginLocationBuiltIn != "built-in" {
		t.Errorf("Expected PluginLocationBuiltIn to be 'built-in', got %s", PluginLocationBuiltIn)
	}

	if PluginLocationExternal != "external" {
		t.Errorf("Expected PluginLocationExternal to be 'external', got %s", PluginLocationExternal)
	}
}

func TestGetDataClient_NilBuiltInClient(t *testing.T) {
	// Arrange
	config := common.Config{}

	// Act
	client := GetDataClient(PluginLocationBuiltIn, nil, config)

	// Assert
	if client != nil {
		t.Errorf("Expected nil when built-in client is nil, got %T", client)
	}
}

func TestGetStreamClient_NilBuiltInClient(t *testing.T) {
	// Arrange
	config := common.Config{}

	// Act
	client := GetStreamClient(PluginLocationBuiltIn, nil, config)

	// Assert
	if client != nil {
		t.Errorf("Expected nil when built-in stream client is nil, got %T", client)
	}
}

// 테이블 주도 테스트 예제
func TestGetDataClient_AllLocations(t *testing.T) {
	tests := []struct {
		name             string
		location         PluginLocation
		expectedType     string
		shouldUseBuiltIn bool
	}{
		{
			name:             "BuiltIn Location",
			location:         PluginLocationBuiltIn,
			expectedType:     "*datasource.mockBuiltInDataClient",
			shouldUseBuiltIn: true,
		},
		{
			name:             "External Location",
			location:         PluginLocationExternal,
			expectedType:     "*hashicorp.DataClient",
			shouldUseBuiltIn: false,
		},
		{
			name:             "Invalid Location",
			location:         PluginLocation("invalid"),
			expectedType:     "*datasource.mockBuiltInDataClient",
			shouldUseBuiltIn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBuiltIn := &mockBuiltInDataClient{}
			config := common.Config{}

			client := GetDataClient(tt.location, mockBuiltIn, config)

			if tt.shouldUseBuiltIn {
				if client != mockBuiltIn {
					t.Errorf("Expected built-in client for location %s", tt.location)
				}
			} else {
				if _, ok := client.(*hashicorp.DataClient); !ok {
					t.Errorf("Expected hashicorp.DataClient for location %s, got %T", tt.location, client)
				}
			}
		})
	}
}

func TestGetStreamClient_AllLocations(t *testing.T) {
	tests := []struct {
		name             string
		location         PluginLocation
		expectedType     string
		shouldUseBuiltIn bool
	}{
		{
			name:             "BuiltIn Location",
			location:         PluginLocationBuiltIn,
			expectedType:     "*datasource.mockBuiltInStreamClient",
			shouldUseBuiltIn: true,
		},
		{
			name:             "External Location",
			location:         PluginLocationExternal,
			expectedType:     "*hashicorp.StreamClient",
			shouldUseBuiltIn: false,
		},
		{
			name:             "Invalid Location",
			location:         PluginLocation("invalid"),
			expectedType:     "*datasource.mockBuiltInStreamClient",
			shouldUseBuiltIn: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockBuiltIn := &mockBuiltInStreamClient{}
			config := common.Config{}

			client := GetStreamClient(tt.location, mockBuiltIn, config)

			if tt.shouldUseBuiltIn {
				if client != mockBuiltIn {
					t.Errorf("Expected built-in stream client for location %s", tt.location)
				}
			} else {
				if _, ok := client.(*hashicorp.StreamClient); !ok {
					t.Errorf("Expected hashicorp.StreamClient for location %s, got %T", tt.location, client)
				}
			}
		})
	}
}
