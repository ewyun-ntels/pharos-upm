package hashicorp

import (
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/model"
	"ntels.com/pharos/core/pkg/plugins/shared"
)

type StreamClient struct {
	Config common.Config
}

func (streamClient *StreamClient) RunStream(request *model.RunStreamRequest) *model.RunStreamResponse {
	response := &model.RunStreamResponse{}

	streamServerRPCClient, err := streamClient.getStreamServerRPCClient(request.Datasource.Type)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	return streamServerRPCClient.RunStream(request)
}

func (streamClient *StreamClient) SubscribeStream(request *model.SubscribeStreamRequest) *model.SubscribeStreamResponse {
	response := &model.SubscribeStreamResponse{}

	streamServerRPCClient, err := streamClient.getStreamServerRPCClient(request.Datasource.Type)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	return streamServerRPCClient.SubscribeStream(request)
}

func (streamClient *StreamClient) UnsubscribeStream(request *model.UnsubscribeStreamRequest) error {
	streamServerRPCClient, err := streamClient.getStreamServerRPCClient(request.Datasource.Type)
	if err != nil {
		return err
	}

	if err := streamServerRPCClient.UnsubscribeStream(request); err != nil {
		return err
	}

	return nil
}

func (streamClient *StreamClient) PublishStream(request *model.PublishStreamRequest) *model.PublishStreamResponse {
	response := &model.PublishStreamResponse{}

	streamServerRPCClient, err := streamClient.getStreamServerRPCClient(request.Datasource.Type)
	if err != nil {
		response.Error = err.Error()
		return response
	}

	return streamServerRPCClient.PublishStream(request)
}

func (streamClient *StreamClient) RemoveDatasource(request *model.RemoveDatasourceRequest) error {
	streamServerRPCClient, err := streamClient.getStreamServerRPCClient(request.Datasource.Type)
	if err != nil {
		return err
	}

	return streamServerRPCClient.RemoveDatasource(request)
}

func (streamClient *StreamClient) getStreamServerRPCClient(pluginName string) (*StreamServerRPCClient, error) {
	if !shared.HashicorpClients.Exist(pluginName) {
		return nil, external.ErrorNotImplemented
	}

	client := shared.HashicorpClients.Get(pluginName)
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense(PluginTypeStreamServer)
	if err != nil {
		return nil, err
	}

	clientTyped, ok := raw.(*StreamServerRPCClient)
	if !ok {
		return nil, external.ErrorInvalidType
	}
	return clientTyped, nil
}
