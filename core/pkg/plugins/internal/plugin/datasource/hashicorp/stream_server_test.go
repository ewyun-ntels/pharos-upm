package hashicorp

import (
	"errors"
	"net/rpc"
	"testing"

	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/pkg/plugins/model"
)

// Tests for StreamServerPlugin
func TestStreamServerPlugin_Server(t *testing.T) {
	mockStreamServer := &mockStreamServer{
		runStreamResponse: &model.RunStreamResponse{},
	}

	plugin := &StreamServerPlugin{
		StreamServer: mockStreamServer,
	}

	server, err := plugin.Server(nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if server == nil {
		t.Error("Expected server, got nil")
	}

	if rpcServer, ok := server.(*StreamServerRPCServer); !ok {
		t.Error("Expected StreamServerRPCServer type")
	} else if rpcServer.streamServer != mockStreamServer {
		t.Error("StreamServer not properly set")
	}
}

func TestStreamServerPlugin_Client(t *testing.T) {
	plugin := &StreamServerPlugin{}

	client, err := plugin.Client(nil, &rpc.Client{})
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if client == nil {
		t.Error("Expected client, got nil")
	}

	if _, ok := client.(*StreamServerRPCClient); !ok {
		t.Error("Expected StreamServerRPCClient type")
	}
}

// Tests for StreamServerRPCServer
func TestStreamServerRPCServer_RunStream_Success(t *testing.T) {
	expectedResponse := &model.RunStreamResponse{
		Data:  "test-data",
		Error: "",
	}

	mockStreamServer := &mockStreamServer{
		runStreamResponse: expectedResponse,
	}

	server := &StreamServerRPCServer{
		streamServer: mockStreamServer,
	}

	request := &model.RunStreamRequest{
		Path: "/test",
	}
	var response model.RunStreamResponse

	err := server.RunStream(request, &response)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response.Data != expectedResponse.Data {
		t.Errorf("Expected %s, got %s", expectedResponse.Data, response.Data)
	}
}

func TestStreamServerRPCServer_RunStream_NoStreamServer(t *testing.T) {
	server := &StreamServerRPCServer{
		streamServer: nil,
	}

	request := &model.RunStreamRequest{}
	var response model.RunStreamResponse

	err := server.RunStream(request, &response)
	if err != external.ErrorNotSupportedStreamServer {
		t.Errorf("Expected ErrorNotSupportedStreamServer, got %v", err)
	}
}

func TestStreamServerRPCServer_SubscribeStream_Success(t *testing.T) {
	expectedResponse := &model.SubscribeStreamResponse{
		Data:  "subscribe-data",
		Error: "",
	}

	mockStreamServer := &mockStreamServer{
		subscribeStreamResponse: expectedResponse,
	}

	server := &StreamServerRPCServer{
		streamServer: mockStreamServer,
	}

	request := &model.SubscribeStreamRequest{
		Path: "/subscribe",
	}
	var response model.SubscribeStreamResponse

	err := server.SubscribeStream(request, &response)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response.Data != expectedResponse.Data {
		t.Errorf("Expected %s, got %s", expectedResponse.Data, response.Data)
	}
}

func TestStreamServerRPCServer_SubscribeStream_NoStreamServer(t *testing.T) {
	server := &StreamServerRPCServer{
		streamServer: nil,
	}

	request := &model.SubscribeStreamRequest{}
	var response model.SubscribeStreamResponse

	err := server.SubscribeStream(request, &response)
	if err != external.ErrorNotSupportedStreamServer {
		t.Errorf("Expected ErrorNotSupportedStreamServer, got %v", err)
	}
}

func TestStreamServerRPCServer_UnsubscribeStream_Success(t *testing.T) {
	mockStreamServer := &mockStreamServer{
		unsubscribeStreamErr: nil,
	}

	server := &StreamServerRPCServer{
		streamServer: mockStreamServer,
	}

	request := &model.UnsubscribeStreamRequest{
		Path: "/unsubscribe",
	}
	var response error

	err := server.UnsubscribeStream(request, &response)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response != nil {
		t.Errorf("Expected nil error in response, got %v", response)
	}
}

func TestStreamServerRPCServer_UnsubscribeStream_WithError(t *testing.T) {
	expectedErr := errors.New("unsubscribe error")

	mockStreamServer := &mockStreamServer{
		unsubscribeStreamErr: expectedErr,
	}

	server := &StreamServerRPCServer{
		streamServer: mockStreamServer,
	}

	request := &model.UnsubscribeStreamRequest{
		Path: "/unsubscribe",
	}
	var response error

	err := server.UnsubscribeStream(request, &response)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response != expectedErr {
		t.Errorf("Expected %v in response, got %v", expectedErr, response)
	}
}

func TestStreamServerRPCServer_UnsubscribeStream_NoStreamServer(t *testing.T) {
	server := &StreamServerRPCServer{
		streamServer: nil,
	}

	request := &model.UnsubscribeStreamRequest{}
	var response error

	err := server.UnsubscribeStream(request, &response)
	if err != external.ErrorNotSupportedStreamServer {
		t.Errorf("Expected ErrorNotSupportedStreamServer, got %v", err)
	}
}

func TestStreamServerRPCServer_PublishStream_Success(t *testing.T) {
	expectedResponse := &model.PublishStreamResponse{
		Data:  "publish-data",
		Error: "",
	}

	mockStreamServer := &mockStreamServer{
		publishStreamResponse: expectedResponse,
	}

	server := &StreamServerRPCServer{
		streamServer: mockStreamServer,
	}

	request := &model.PublishStreamRequest{
		Path: "/publish",
	}
	var response model.PublishStreamResponse

	err := server.PublishStream(request, &response)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if response.Data != expectedResponse.Data {
		t.Errorf("Expected %s, got %s", expectedResponse.Data, response.Data)
	}
}

func TestStreamServerRPCServer_PublishStream_NoStreamServer(t *testing.T) {
	server := &StreamServerRPCServer{
		streamServer: nil,
	}

	request := &model.PublishStreamRequest{}
	var response model.PublishStreamResponse

	err := server.PublishStream(request, &response)
	if err != external.ErrorNotSupportedStreamServer {
		t.Errorf("Expected ErrorNotSupportedStreamServer, got %v", err)
	}
}

func TestStreamServerRPCServer_RemoveDatasource_Success(t *testing.T) {
	mockStreamServer := &mockStreamServer{
		removeDatasourceErr: nil,
	}

	server := &StreamServerRPCServer{
		streamServer: mockStreamServer,
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

func TestStreamServerRPCServer_RemoveDatasource_NoStreamServer(t *testing.T) {
	server := &StreamServerRPCServer{
		streamServer: nil,
	}

	request := &model.RemoveDatasourceRequest{}
	var response error

	err := server.RemoveDatasource(request, &response)
	if err != external.ErrorNotSupportedStreamServer {
		t.Errorf("Expected ErrorNotSupportedStreamServer, got %v", err)
	}
}
