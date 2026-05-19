package model

import (
	"context"
	"encoding/json"
	"testing"

	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/common"
)

func TestDatasource_ToDatabaseConfig_ClickHouse(t *testing.T) {
	datasource := &Datasource{
		Name: "test-clickhouse",
		Type: orm.DriverClickHouse,
		Data: map[string]any{
			"host":     "localhost",
			"port":     9000,
			"database": "default",
			"username": "default",
			"password": "password",
		},
	}

	config := common.Config{}
	databaseConfig, err := datasource.ToDatabaseConfig(config)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if databaseConfig.Driver != orm.DriverClickHouse {
		t.Errorf("Expected driver %s, got %s", orm.DriverClickHouse, databaseConfig.Driver)
	}

	// ClickHouse 특정 필드 검증은 orm.DatabaseConfig 구조에 따라 달라질 수 있음
}

func TestDatasource_ToDatabaseConfig_PostgreSQL(t *testing.T) {
	datasource := &Datasource{
		Name: "test-postgresql",
		Type: orm.DriverPostgreSQL,
		Data: map[string]any{
			"host":     "localhost",
			"port":     5432,
			"database": "testdb",
			"username": "postgres",
			"password": "password",
		},
	}

	config := common.Config{}
	databaseConfig, err := datasource.ToDatabaseConfig(config)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if databaseConfig.Driver != orm.DriverPostgreSQL {
		t.Errorf("Expected driver %s, got %s", orm.DriverPostgreSQL, databaseConfig.Driver)
	}
}

func TestDatasource_ToDatabaseConfig_Altibase(t *testing.T) {
	// Altibase 버전별 드라이버 설정을 위한 mock config
	config := common.Config{
		// SharedLibrary 필드 구조를 정확히 파악할 수 없으므로 기본 구조로 테스트
	}

	datasource := &Datasource{
		Name: "test-altibase",
		Type: orm.DriverAltibase,
		Data: map[string]any{
			"host":     "localhost",
			"port":     20300,
			"database": "mydb",
			"username": "sys",
			"password": "manager",
			"version":  float64(7),
		},
	}

	databaseConfig, err := datasource.ToDatabaseConfig(config)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if databaseConfig.Driver != orm.DriverAltibase {
		t.Errorf("Expected driver %s, got %s", orm.DriverAltibase, databaseConfig.Driver)
	}
}

func TestDatasource_ToDatabaseConfig_Altibase_InvalidVersion(t *testing.T) {
	config := common.Config{}
	datasource := &Datasource{
		Name: "test-altibase",
		Type: orm.DriverAltibase,
		Data: map[string]any{
			"host":     "localhost",
			"port":     20300,
			"database": "mydb",
			"username": "sys",
			"password": "manager",
			// version 필드가 없거나 잘못된 타입
		},
	}

	_, err := datasource.ToDatabaseConfig(config)

	if err != external.ErrorNotSupported {
		t.Errorf("Expected ErrorNotSupported, got %v", err)
	}
}

func TestDatasource_ToDatabaseConfig_UnsupportedDriver(t *testing.T) {
	datasource := &Datasource{
		Name: "test-unsupported",
		Type: "unsupported-driver",
		Data: map[string]any{
			"host": "localhost",
		},
	}

	config := common.Config{}
	_, err := datasource.ToDatabaseConfig(config)

	if err != external.ErrorNotImplemented {
		t.Errorf("Expected ErrorNotImplemented, got %v", err)
	}
}

func TestDatasource_ToDatabaseConfig_InvalidJSON(t *testing.T) {
	datasource := &Datasource{
		Name: "test-invalid-json",
		Type: orm.DriverClickHouse,
		Data: map[string]any{
			"invalid": func() {}, // 함수는 JSON으로 마샬링 불가
		},
	}

	config := common.Config{}
	_, err := datasource.ToDatabaseConfig(config)

	if err == nil {
		t.Error("Expected JSON marshaling error, got nil")
	}
}

func TestDatasourceClients_QueryData_NotSupportedType(t *testing.T) {
	datasourceClients := &DatasourceClients{
		Map: internal.NewMap[*DatasourceClient](),
	}

	request := &QueryDataRequest{
		Datasource: Datasource{
			Type: "unsupported-type",
		},
		Queries: []Query{
			{ID: "query1", SQL: "SELECT 1"},
		},
	}

	response := datasourceClients.QueryData(context.Background(), request)

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Results == nil {
		t.Fatal("Expected Results map to be initialized")
	}

	if len(response.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(response.Results))
	}

	result, exists := response.Results["query1"]
	if !exists {
		t.Error("Expected query1 result to exist")
	}

	if result.Error != external.ErrorNotSupportedDatasourceType.Error() {
		t.Errorf("Expected error %s, got %s", external.ErrorNotSupportedDatasourceType.Error(), result.Error)
	}
}

func TestDatasourceClients_QueryData_NoDataClient(t *testing.T) {
	datasourceClients := &DatasourceClients{
		Map: internal.NewMap[*DatasourceClient](),
	}

	// DataClient가 nil인 클라이언트 등록
	datasourceClients.Set("test-type", &DatasourceClient{
		DataClient:   nil,
		StreamClient: nil,
	})

	request := &QueryDataRequest{
		Datasource: Datasource{
			Type: "test-type",
		},
		Queries: []Query{
			{ID: "query1", SQL: "SELECT 1"},
		},
	}

	response := datasourceClients.QueryData(context.Background(), request)

	if len(response.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(response.Results))
	}

	result, exists := response.Results["query1"]
	if !exists {
		t.Error("Expected query1 result to exist")
	}

	if result.Error != external.ErrorNotSupportedDataServer.Error() {
		t.Errorf("Expected error %s, got %s", external.ErrorNotSupportedDataServer.Error(), result.Error)
	}
}

// Mock DataClient for testing
type mockDataClient struct {
	queryDataFunc func(*QueryDataRequest) *QueryDataResponse
}

func (m *mockDataClient) QueryData(_ context.Context, request *QueryDataRequest) *QueryDataResponse {
	if m.queryDataFunc != nil {
		return m.queryDataFunc(request)
	}
	return &QueryDataResponse{
		Results: map[string]QueryDataResult{
			"test": {Error: ""},
		},
	}
}

func (m *mockDataClient) RemoveDatasource(request *RemoveDatasourceRequest) error {
	return nil
}

func TestDatasourceClients_QueryData_WithDataClient(t *testing.T) {
	datasourceClients := &DatasourceClients{
		Map: internal.NewMap[*DatasourceClient](),
	}

	expectedResponse := &QueryDataResponse{
		Results: map[string]QueryDataResult{
			"query1": {Error: ""},
		},
	}

	mockClient := &mockDataClient{
		queryDataFunc: func(request *QueryDataRequest) *QueryDataResponse {
			return expectedResponse
		},
	}

	datasourceClients.Set("test-type", &DatasourceClient{
		DataClient:   mockClient,
		StreamClient: nil,
	})

	request := &QueryDataRequest{
		Datasource: Datasource{
			Type: "test-type",
		},
		Queries: []Query{
			{ID: "query1", SQL: "SELECT 1"},
		},
	}

	response := datasourceClients.QueryData(context.Background(), request)

	if response != expectedResponse {
		t.Error("Expected response from DataClient")
	}
}

func TestDatasourceClients_RunStream_NotSupportedType(t *testing.T) {
	datasourceClients := &DatasourceClients{
		Map: internal.NewMap[*DatasourceClient](),
	}

	request := &RunStreamRequest{
		Datasource: Datasource{
			Type: "unsupported-type",
		},
	}

	response := datasourceClients.RunStream(request)

	if response.Error != external.ErrorNotSupportedDatasourceType.Error() {
		t.Errorf("Expected error %s, got %s", external.ErrorNotSupportedDatasourceType.Error(), response.Error)
	}
}

func TestDatasourceClients_RunStream_NoStreamClient(t *testing.T) {
	datasourceClients := &DatasourceClients{
		Map: internal.NewMap[*DatasourceClient](),
	}

	datasourceClients.Set("test-type", &DatasourceClient{
		DataClient:   nil,
		StreamClient: nil,
	})

	request := &RunStreamRequest{
		Datasource: Datasource{
			Type: "test-type",
		},
	}

	response := datasourceClients.RunStream(request)

	if response.Error != external.ErrorNotSupportedStreamServer.Error() {
		t.Errorf("Expected error %s, got %s", external.ErrorNotSupportedStreamServer.Error(), response.Error)
	}
}

// Mock StreamClient for testing
type mockStreamClient struct{}

func (m *mockStreamClient) RunStream(request *RunStreamRequest) *RunStreamResponse {
	return &RunStreamResponse{
		Data:  "test-data",
		Error: "",
	}
}

func (m *mockStreamClient) SubscribeStream(request *SubscribeStreamRequest) *SubscribeStreamResponse {
	return &SubscribeStreamResponse{
		Data:  "subscribed",
		Error: "",
	}
}

func (m *mockStreamClient) UnsubscribeStream(request *UnsubscribeStreamRequest) error {
	return nil
}

func (m *mockStreamClient) PublishStream(request *PublishStreamRequest) *PublishStreamResponse {
	return &PublishStreamResponse{
		Data:  "published",
		Error: "",
	}
}

func (m *mockStreamClient) RemoveDatasource(request *RemoveDatasourceRequest) error {
	return nil
}

func TestDatasourceClients_RunStream_WithStreamClient(t *testing.T) {
	datasourceClients := &DatasourceClients{
		Map: internal.NewMap[*DatasourceClient](),
	}

	mockClient := &mockStreamClient{}

	datasourceClients.Set("test-type", &DatasourceClient{
		DataClient:   nil,
		StreamClient: mockClient,
	})

	request := &RunStreamRequest{
		Datasource: Datasource{
			Type: "test-type",
		},
	}

	response := datasourceClients.RunStream(request)

	if response.Data != "test-data" {
		t.Errorf("Expected 'test-data', got %s", response.Data)
	}

	if response.Error != "" {
		t.Errorf("Expected no error, got %s", response.Error)
	}
}

func TestDatasourceClients_SubscribeStream(t *testing.T) {
	datasourceClients := &DatasourceClients{
		Map: internal.NewMap[*DatasourceClient](),
	}

	mockClient := &mockStreamClient{}

	datasourceClients.Set("test-type", &DatasourceClient{
		DataClient:   nil,
		StreamClient: mockClient,
	})

	request := &SubscribeStreamRequest{
		Datasource: Datasource{
			Type: "test-type",
		},
	}

	response := datasourceClients.SubscribeStream(request)

	if response.Data != "subscribed" {
		t.Errorf("Expected 'subscribed', got %s", response.Data)
	}
}

func TestDatasourceClients_UnsubscribeStream(t *testing.T) {
	datasourceClients := &DatasourceClients{
		Map: internal.NewMap[*DatasourceClient](),
	}

	mockClient := &mockStreamClient{}

	datasourceClients.Set("test-type", &DatasourceClient{
		DataClient:   nil,
		StreamClient: mockClient,
	})

	request := &UnsubscribeStreamRequest{
		Datasource: Datasource{
			Type: "test-type",
		},
	}

	err := datasourceClients.UnsubscribeStream(request)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestDatasourceClients_UnsubscribeStream_NotSupportedType(t *testing.T) {
	datasourceClients := &DatasourceClients{
		Map: internal.NewMap[*DatasourceClient](),
	}

	request := &UnsubscribeStreamRequest{
		Datasource: Datasource{
			Type: "unsupported-type",
		},
	}

	err := datasourceClients.UnsubscribeStream(request)

	if err != external.ErrorNotSupportedDatasourceType {
		t.Errorf("Expected ErrorNotSupportedDatasourceType, got %v", err)
	}
}

func TestDatasourceClients_PublishStream(t *testing.T) {
	datasourceClients := &DatasourceClients{
		Map: internal.NewMap[*DatasourceClient](),
	}

	mockClient := &mockStreamClient{}

	datasourceClients.Set("test-type", &DatasourceClient{
		DataClient:   nil,
		StreamClient: mockClient,
	})

	request := &PublishStreamRequest{
		Datasource: Datasource{
			Type: "test-type",
		},
	}

	response := datasourceClients.PublishStream(request)

	if response.Data != "published" {
		t.Errorf("Expected 'published', got %s", response.Data)
	}
}

func TestDatasourceClients_RemoveDatasource(t *testing.T) {
	datasourceClients := &DatasourceClients{
		Map: internal.NewMap[*DatasourceClient](),
	}

	mockDataClient := &mockDataClient{}
	mockStreamClient := &mockStreamClient{}

	datasourceClients.Set("test-type", &DatasourceClient{
		DataClient:   mockDataClient,
		StreamClient: mockStreamClient,
	})

	request := &RemoveDatasourceRequest{
		Datasource: Datasource{
			Type: "test-type",
		},
	}

	err := datasourceClients.RemoveDatasource(request)

	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
}

func TestDatasourceClients_RemoveDatasource_NotSupportedType(t *testing.T) {
	datasourceClients := &DatasourceClients{
		Map: internal.NewMap[*DatasourceClient](),
	}

	request := &RemoveDatasourceRequest{
		Datasource: Datasource{
			Type: "unsupported-type",
		},
	}

	err := datasourceClients.RemoveDatasource(request)

	if err != external.ErrorNotSupportedDatasourceType {
		t.Errorf("Expected ErrorNotSupportedDatasourceType, got %v", err)
	}
}

func TestDatasource_JSONSerialization(t *testing.T) {
	original := &Datasource{
		Name: "test-datasource",
		Type: "test-type",
		Data: map[string]any{
			"host":   "localhost",
			"port":   5432,
			"secure": true,
		},
	}

	// JSON 마샬링 테스트
	jsonData, err := json.Marshal(original)
	if err != nil {
		t.Errorf("Failed to marshal datasource: %v", err)
	}

	// JSON 언마샬링 테스트
	var unmarshaled Datasource
	err = json.Unmarshal(jsonData, &unmarshaled)
	if err != nil {
		t.Errorf("Failed to unmarshal datasource: %v", err)
	}

	// 값 검증
	if unmarshaled.Name != original.Name {
		t.Errorf("Expected name %s, got %s", original.Name, unmarshaled.Name)
	}

	if unmarshaled.Type != original.Type {
		t.Errorf("Expected type %s, got %s", original.Type, unmarshaled.Type)
	}

	// Data 맵 검증
	if host, ok := unmarshaled.Data["host"].(string); !ok || host != "localhost" {
		t.Errorf("Expected host localhost, got %v", unmarshaled.Data["host"])
	}

	if port, ok := unmarshaled.Data["port"].(float64); !ok || port != 5432 {
		t.Errorf("Expected port 5432, got %v", unmarshaled.Data["port"])
	}

	if secure, ok := unmarshaled.Data["secure"].(bool); !ok || !secure {
		t.Errorf("Expected secure true, got %v", unmarshaled.Data["secure"])
	}
}

// Benchmark tests
func BenchmarkDatasource_ToDatabaseConfig(b *testing.B) {
	datasource := &Datasource{
		Name: "benchmark-test",
		Type: orm.DriverPostgreSQL,
		Data: map[string]any{
			"host":     "localhost",
			"port":     5432,
			"database": "testdb",
			"username": "postgres",
			"password": "password",
		},
	}

	config := common.Config{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = datasource.ToDatabaseConfig(config)
	}
}

func BenchmarkDatasourceClients_QueryData(b *testing.B) {
	datasourceClients := &DatasourceClients{
		Map: internal.NewMap[*DatasourceClient](),
	}

	mockClient := &mockDataClient{}

	datasourceClients.Set("test-type", &DatasourceClient{
		DataClient:   mockClient,
		StreamClient: nil,
	})

	request := &QueryDataRequest{
		Datasource: Datasource{
			Type: "test-type",
		},
		Queries: []Query{
			{ID: "query1", SQL: "SELECT 1"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = datasourceClients.QueryData(context.Background(), request)
	}
}
