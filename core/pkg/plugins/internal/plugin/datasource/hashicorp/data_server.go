package hashicorp

import (
	"context"
	"net/rpc"

	"github.com/hashicorp/go-plugin"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/pkg/plugins/model"
)

type DataServerPlugin struct {
	DataServer model.Data
}

func (dataServerPlugin *DataServerPlugin) Server(*plugin.MuxBroker) (any, error) {
	return &DataServerRPCServer{
		dataServer: dataServerPlugin.DataServer,
	}, nil
}

func (dataServerPlugin *DataServerPlugin) Client(_ *plugin.MuxBroker, client *rpc.Client) (any, error) {
	return &DataServerRPCClient{client: client}, nil
}

type DataServerRPCServer struct {
	dataServer model.Data
}

func (dataServerRPCServer *DataServerRPCServer) QueryData(request *model.QueryDataRequest, response *model.QueryDataResponse) error {
	if dataServerRPCServer.dataServer == nil {
		return external.ErrorNotSupportedDataServer
	}

	// RPC boundary: context cannot be propagated through net/rpc, use Background
	*response = *dataServerRPCServer.dataServer.QueryData(context.Background(), request)

	return nil
}

func (dataServerRPCServer *DataServerRPCServer) RemoveDatasource(request *model.RemoveDatasourceRequest, response *error) error {
	if dataServerRPCServer.dataServer == nil {
		return external.ErrorNotSupportedDataServer
	}

	*response = dataServerRPCServer.dataServer.RemoveDatasource(request)

	return nil
}

type DataServerRPCClient struct {
	client *rpc.Client
}

func (dataServerRPCClient *DataServerRPCClient) QueryData(request *model.QueryDataRequest) *model.QueryDataResponse {
	response := &model.QueryDataResponse{Results: map[string]model.QueryDataResult{}}

	if err := dataServerRPCClient.client.Call("Plugin.QueryData", request, response); err != nil {
		for _, query := range request.Queries {
			response.Results[query.ID] = model.QueryDataResult{Error: err.Error()}
		}

		return response
	}

	return response
}

func (dataServerRPCClient *DataServerRPCClient) RemoveDatasource(request *model.RemoveDatasourceRequest) error {
	var response error

	if err := dataServerRPCClient.client.Call("Plugin.RemoveDatasource", request, &response); err != nil {
		return err
	}

	return response
}
