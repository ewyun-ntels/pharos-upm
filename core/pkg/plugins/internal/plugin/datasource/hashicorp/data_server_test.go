package hashicorp

import (
	"errors"
	"net/rpc"
	"testing"

	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/pkg/plugins/model"
)

// Tests for DataServerPlugin
func TestDataServerPlugin_Server(t *testing.T) {
	mockDataServer := &mockDataServer{
		queryDataResponse: &model.QueryDataResponse{},
	}

	plugin := &DataServerPlugin{
		DataServer: mockDataServer,
	}

	server, err := plugin.Server(nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Error("Expected server, got nil")
	}

	if rpcServer, ok := server.(*DataServerRPCServer); !ok {
		t.Error("Expected DataServerRPCServer type")
	} else if rpcServer.dataServer != mockDataServer {
		t.Error("DataServer not properly set")
	}
}

func TestDataServerPlugin_Client(t *testing.T) {
	plugin := &DataServerPlugin{}

	client, err := plugin.Client(nil, &rpc.Client{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Error("Expected client, got nil")
	}

	if _, ok := client.(*DataServerRPCClient); !ok {
		t.Error("Expected DataServerRPCClient type")
	}
}

// Tests for DataServerRPCServer
func TestDataServerRPCServer_QueryData_Success(t *testing.T) {
	expectedResponse := &model.QueryDataResponse{
		Results: map[string]model.QueryDataResult{
			"test": {Error: ""},
		},
	}

	mockDataServer := &mockDataServer{
		queryDataResponse: expectedResponse,
	}

	server := &DataServerRPCServer{
		dataServer: mockDataServer,
	}

	request := &model.QueryDataRequest{
		Queries: []model.Query{{ID: "test"}},
	}
	var response model.QueryDataResponse

	err := server.QueryData(request, &response)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(response.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(response.Results))
	}
}

func TestDataServerRPCServer_QueryData_NoDataServer(t *testing.T) {
	server := &DataServerRPCServer{
		dataServer: nil,
	}

	request := &model.QueryDataRequest{}
	var response model.QueryDataResponse

	err := server.QueryData(request, &response)
	if err != external.ErrorNotSupportedDataServer {
		t.Errorf("Expected ErrorNotSupportedDataServer, got %v", err)
	}
}

func TestDataServerRPCServer_RemoveDatasource_Success(t *testing.T) {
	mockDataServer := &mockDataServer{
		removeDatasourceErr: nil,
	}

	server := &DataServerRPCServer{
		dataServer: mockDataServer,
	}

	request := &model.RemoveDatasourceRequest{}
	var response error

	err := server.RemoveDatasource(request, &response)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response != nil {
		t.Errorf("Expected nil error in response, got %v", response)
	}
}

func TestDataServerRPCServer_RemoveDatasource_WithError(t *testing.T) {
	expectedErr := errors.New("remove error")

	mockDataServer := &mockDataServer{
		removeDatasourceErr: expectedErr,
	}

	server := &DataServerRPCServer{
		dataServer: mockDataServer,
	}

	request := &model.RemoveDatasourceRequest{}
	var response error

	err := server.RemoveDatasource(request, &response)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response != expectedErr {
		t.Errorf("Expected %v in response, got %v", expectedErr, response)
	}
}

func TestDataServerRPCServer_RemoveDatasource_NoDataServer(t *testing.T) {
	server := &DataServerRPCServer{
		dataServer: nil,
	}

	request := &model.RemoveDatasourceRequest{}
	var response error

	err := server.RemoveDatasource(request, &response)
	if err != external.ErrorNotSupportedDataServer {
		t.Errorf("Expected ErrorNotSupportedDataServer, got %v", err)
	}
}
