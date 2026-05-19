package hashicorp

import (
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/pkg/plugins/model"
)

type StreamServerPlugin struct {
	StreamServer model.Stream
}

func (streamServerPlugin *StreamServerPlugin) Server(*plugin.MuxBroker) (any, error) {
	return &StreamServerRPCServer{
		streamServer: streamServerPlugin.StreamServer,
	}, nil
}

func (streamServerPlugin *StreamServerPlugin) Client(_ *plugin.MuxBroker, client *rpc.Client) (any, error) {
	return &StreamServerRPCClient{client: client}, nil
}

type StreamServerRPCServer struct {
	streamServer model.Stream
}

func (streamServerRPCServer *StreamServerRPCServer) RunStream(request *model.RunStreamRequest, response *model.RunStreamResponse) error {
	if streamServerRPCServer.streamServer == nil {
		return external.ErrorNotSupportedStreamServer
	}

	*response = *streamServerRPCServer.streamServer.RunStream(request)

	return nil
}

func (streamServerRPCServer *StreamServerRPCServer) SubscribeStream(request *model.SubscribeStreamRequest, response *model.SubscribeStreamResponse) error {
	if streamServerRPCServer.streamServer == nil {
		return external.ErrorNotSupportedStreamServer
	}

	*response = *streamServerRPCServer.streamServer.SubscribeStream(request)

	return nil
}

func (streamServerRPCServer *StreamServerRPCServer) UnsubscribeStream(request *model.UnsubscribeStreamRequest, response *error) error {
	if streamServerRPCServer.streamServer == nil {
		return external.ErrorNotSupportedStreamServer
	}

	*response = streamServerRPCServer.streamServer.UnsubscribeStream(request)

	return nil
}

func (streamServerRPCServer *StreamServerRPCServer) PublishStream(request *model.PublishStreamRequest, response *model.PublishStreamResponse) error {
	if streamServerRPCServer.streamServer == nil {
		return external.ErrorNotSupportedStreamServer
	}

	*response = *streamServerRPCServer.streamServer.PublishStream(request)

	return nil
}

func (streamServerRPCServer *StreamServerRPCServer) RemoveDatasource(request *model.RemoveDatasourceRequest, response *error) error {
	if streamServerRPCServer.streamServer == nil {
		return external.ErrorNotSupportedStreamServer
	}

	*response = streamServerRPCServer.streamServer.RemoveDatasource(request)

	return nil
}

type StreamServerRPCClient struct {
	client *rpc.Client
}

func (streamServerRPCClient *StreamServerRPCClient) RunStream(request *model.RunStreamRequest) *model.RunStreamResponse {
	response := &model.RunStreamResponse{}

	if err := streamServerRPCClient.client.Call("Plugin.RunStream", request, response); err != nil {
		response.Error = err.Error()
		return response
	}

	return response
}

func (streamServerRPCClient *StreamServerRPCClient) SubscribeStream(request *model.SubscribeStreamRequest) *model.SubscribeStreamResponse {
	response := &model.SubscribeStreamResponse{}

	if err := streamServerRPCClient.client.Call("Plugin.SubscribeStream", request, response); err != nil {
		response.Error = err.Error()
		return response
	}

	return response
}

func (streamServerRPCClient *StreamServerRPCClient) UnsubscribeStream(request *model.UnsubscribeStreamRequest) error {
	var response error

	if err := streamServerRPCClient.client.Call("Plugin.UnsubscribeStream", request, &response); err != nil {
		return err
	}

	return response
}

func (streamServerRPCClient *StreamServerRPCClient) PublishStream(request *model.PublishStreamRequest) *model.PublishStreamResponse {
	response := &model.PublishStreamResponse{}

	if err := streamServerRPCClient.client.Call("Plugin.PublishStream", request, response); err != nil {
		response.Error = err.Error()
		return response
	}

	return response
}

func (streamServerRPCClient *StreamServerRPCClient) RemoveDatasource(request *model.RemoveDatasourceRequest) error {
	var response error

	if err := streamServerRPCClient.client.Call("Plugin.RemoveDatasource", request, &response); err != nil {
		return err
	}

	return response
}
