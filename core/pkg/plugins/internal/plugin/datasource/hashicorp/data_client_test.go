package hashicorp

import (
	"context"
	"testing"

	"github.com/hashicorp/go-plugin"
	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/model"
	"ntels.com/pharos/core/pkg/plugins/shared"
)

// Mock 구조체들
type mockDataServer struct {
	queryDataResponse   *model.QueryDataResponse
	removeDatasourceErr error
}

func (m *mockDataServer) QueryData(_ context.Context, request *model.QueryDataRequest) *model.QueryDataResponse {
	return m.queryDataResponse
}

func (m *mockDataServer) RemoveDatasource(request *model.RemoveDatasourceRequest) error {
	return m.removeDatasourceErr
}

func setupTestEnvironment() {
	shared.HashicorpClients = internal.NewMap[*plugin.Client]()
}

func TestDataClient_QueryData_PluginNotFound(t *testing.T) {
	setupTestEnvironment()

	dataClient := &DataClient{
		Config: common.Config{},
	}

	request := &model.QueryDataRequest{
		Datasource: model.Datasource{
			Type: "non-existent-plugin",
		},
		Queries: []model.Query{
			{ID: "query1", SQL: "SELECT 1"},
		},
	}

	response := dataClient.QueryData(context.Background(), request)

	// 검증
	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if len(response.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(response.Results))
	}

	result := response.Results["query1"]
	if result.Error != "not implemented" {
		t.Errorf("Expected 'not implemented' error, got %s", result.Error)
	}
}

func TestDataClient_RemoveDatasource_PluginNotFound(t *testing.T) {
	setupTestEnvironment()

	dataClient := &DataClient{
		Config: common.Config{},
	}

	request := &model.RemoveDatasourceRequest{
		Datasource: model.Datasource{
			Type: "non-existent-plugin",
		},
	}

	err := dataClient.RemoveDatasource(request)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err.Error() != "not implemented" {
		t.Errorf("Expected 'not implemented' error, got %s", err.Error())
	}
}

func TestDataClient_getDataServerRPCClient_PluginNotExist(t *testing.T) {
	setupTestEnvironment()

	dataClient := &DataClient{
		Config: common.Config{},
	}

	client, err := dataClient.getDataServerRPCClient("non-existent-plugin")

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if client != nil {
		t.Error("Expected nil client, got client")
	}

	if err.Error() != "not implemented" {
		t.Errorf("Expected 'not implemented' error, got %s", err.Error())
	}
}
