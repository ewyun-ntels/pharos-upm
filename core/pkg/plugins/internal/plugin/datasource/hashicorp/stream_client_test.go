package hashicorp

import (
	"testing"

	"github.com/hashicorp/go-plugin"
	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/model"
	"ntels.com/pharos/core/pkg/plugins/shared"
)

// Mock StreamServer
type mockStreamServer struct {
	runStreamResponse       *model.RunStreamResponse
	subscribeStreamResponse *model.SubscribeStreamResponse
	unsubscribeStreamErr    error
	publishStreamResponse   *model.PublishStreamResponse
	removeDatasourceErr     error
}

func (m *mockStreamServer) RunStream(request *model.RunStreamRequest) *model.RunStreamResponse {
	return m.runStreamResponse
}

func (m *mockStreamServer) SubscribeStream(request *model.SubscribeStreamRequest) *model.SubscribeStreamResponse {
	return m.subscribeStreamResponse
}

func (m *mockStreamServer) UnsubscribeStream(request *model.UnsubscribeStreamRequest) error {
	return m.unsubscribeStreamErr
}

func (m *mockStreamServer) PublishStream(request *model.PublishStreamRequest) *model.PublishStreamResponse {
	return m.publishStreamResponse
}

func (m *mockStreamServer) RemoveDatasource(request *model.RemoveDatasourceRequest) error {
	return m.removeDatasourceErr
}

func TestStreamClient_RunStream_PluginNotFound(t *testing.T) {
	shared.HashicorpClients = internal.NewMap[*plugin.Client]()

	streamClient := &StreamClient{
		Config: common.Config{},
	}

	request := &model.RunStreamRequest{
		Datasource: model.Datasource{
			Type: "non-existent-plugin",
		},
		Path: "/test",
		Data: "test",
	}

	response := streamClient.RunStream(request)

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Error != "not implemented" {
		t.Errorf("Expected 'not implemented' error, got %s", response.Error)
	}
}

func TestStreamClient_SubscribeStream_PluginNotFound(t *testing.T) {
	shared.HashicorpClients = internal.NewMap[*plugin.Client]()

	streamClient := &StreamClient{
		Config: common.Config{},
	}

	request := &model.SubscribeStreamRequest{
		Datasource: model.Datasource{
			Type: "non-existent-plugin",
		},
		Path: "/subscribe",
		Data: "test",
	}

	response := streamClient.SubscribeStream(request)

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Error != "not implemented" {
		t.Errorf("Expected 'not implemented' error, got %s", response.Error)
	}
}

func TestStreamClient_UnsubscribeStream_PluginNotFound(t *testing.T) {
	shared.HashicorpClients = internal.NewMap[*plugin.Client]()

	streamClient := &StreamClient{
		Config: common.Config{},
	}

	request := &model.UnsubscribeStreamRequest{
		Datasource: model.Datasource{
			Type: "non-existent-plugin",
		},
		Path: "/unsubscribe",
		Data: "test",
	}

	err := streamClient.UnsubscribeStream(request)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err.Error() != "not implemented" {
		t.Errorf("Expected 'not implemented' error, got %s", err.Error())
	}
}

func TestStreamClient_PublishStream_PluginNotFound(t *testing.T) {
	shared.HashicorpClients = internal.NewMap[*plugin.Client]()

	streamClient := &StreamClient{
		Config: common.Config{},
	}

	request := &model.PublishStreamRequest{
		Datasource: model.Datasource{
			Type: "non-existent-plugin",
		},
		Path: "/publish",
		Data: "test",
	}

	response := streamClient.PublishStream(request)

	if response == nil {
		t.Fatal("Expected response, got nil")
	}

	if response.Error != "not implemented" {
		t.Errorf("Expected 'not implemented' error, got %s", response.Error)
	}
}

func TestStreamClient_RemoveDatasource_PluginNotFound(t *testing.T) {
	shared.HashicorpClients = internal.NewMap[*plugin.Client]()

	streamClient := &StreamClient{
		Config: common.Config{},
	}

	request := &model.RemoveDatasourceRequest{
		Datasource: model.Datasource{
			Type: "non-existent-plugin",
		},
	}

	err := streamClient.RemoveDatasource(request)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	if err.Error() != "not implemented" {
		t.Errorf("Expected 'not implemented' error, got %s", err.Error())
	}
}

func TestStreamClient_getStreamServerRPCClient_PluginNotExist(t *testing.T) {
	shared.HashicorpClients = internal.NewMap[*plugin.Client]()

	streamClient := &StreamClient{
		Config: common.Config{},
	}

	client, err := streamClient.getStreamServerRPCClient("non-existent-plugin")

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
