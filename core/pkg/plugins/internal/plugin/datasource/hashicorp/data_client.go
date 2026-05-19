package hashicorp

import (
	"context"

	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/model"
	"ntels.com/pharos/core/pkg/plugins/shared"
)

type DataClient struct {
	Config common.Config
}

func (dataClient *DataClient) QueryData(_ context.Context, request *model.QueryDataRequest) *model.QueryDataResponse {
	response := &model.QueryDataResponse{Results: map[string]model.QueryDataResult{}}

	dataServerRPCClient, err := dataClient.getDataServerRPCClient(request.Datasource.Type)
	if err != nil {
		for _, query := range request.Queries {
			response.Results[query.ID] = model.QueryDataResult{Error: err.Error()}
		}

		return response
	}

	response = dataServerRPCClient.QueryData(request)
	return response
}

func (dataClient *DataClient) RemoveDatasource(request *model.RemoveDatasourceRequest) error {
	dataServerRPCClient, err := dataClient.getDataServerRPCClient(request.Datasource.Type)
	if err != nil {
		return err
	}

	return dataServerRPCClient.RemoveDatasource(request)
}

func (dataClient *DataClient) getDataServerRPCClient(pluginName string) (*DataServerRPCClient, error) {
	if !shared.HashicorpClients.Exist(pluginName) {
		return nil, external.ErrorNotImplemented
	}

	client := shared.HashicorpClients.Get(pluginName)
	rpcClient, err := client.Client()
	if err != nil {
		return nil, err
	}

	raw, err := rpcClient.Dispense(PluginTypeDataServer)
	if err != nil {
		return nil, err
	}

	clientTyped, ok := raw.(*DataServerRPCClient)
	if !ok {
		return nil, external.ErrorInvalidDataServerRPCClientType
	}
	return clientTyped, nil
}
