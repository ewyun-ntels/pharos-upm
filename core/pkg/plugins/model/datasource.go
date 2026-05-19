package model

import (
	"context"
	"encoding/json"

	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/common"
)

type Datasource struct {
	Name string         `json:"name" db:"name"`
	Type string         `json:"type" db:"type"`
	Data map[string]any `json:"data" db:"-"`
}

func (datasource *Datasource) ToDatabaseConfig(config common.Config) (orm.DatabaseConfig, error) {
	databaseConfig := orm.DatabaseConfig{}

	bytes, err := json.Marshal(datasource.Data)
	if err != nil {
		return databaseConfig, err
	}

	switch datasource.Type {
	case orm.DriverClickHouse:
		databaseConfig.Driver = orm.DriverClickHouse
		return databaseConfig, json.Unmarshal(bytes, &databaseConfig.ClickHouse)
	case orm.DriverPostgreSQL:
		databaseConfig.Driver = orm.DriverPostgreSQL
		return databaseConfig, json.Unmarshal(bytes, &databaseConfig.PostgreSQL)
	case orm.DriverAltibase:
		databaseConfig.Driver = orm.DriverAltibase
		if err := json.Unmarshal(bytes, &databaseConfig.Altibase); err != nil {
			return databaseConfig, err
		}

		version, ok := datasource.Data["version"].(float64)
		if !ok {
			return databaseConfig, external.ErrorNotSupported
		}
		databaseConfig.Altibase.Driver = config.SharedLibrary.Altibase[int(version)].Driver

		return databaseConfig, nil
	case orm.DriverVertica:
		databaseConfig.Driver = orm.DriverVertica
		return databaseConfig, json.Unmarshal(bytes, &databaseConfig.Vertica)
	case orm.DriverPrometheus:
		databaseConfig.Driver = orm.DriverPrometheus
		return databaseConfig, json.Unmarshal(bytes, &databaseConfig.Prometheus)
	case orm.DriverElasticsearch:
		databaseConfig.Driver = orm.DriverElasticsearch
		return databaseConfig, json.Unmarshal(bytes, &databaseConfig.Elasticsearch)
	default:
		return orm.DatabaseConfig{}, external.ErrorNotImplemented
	}
}

type DatasourceClient struct {
	DataClient   Data
	StreamClient Stream
}

type DatasourceClients struct {
	*internal.Map[*DatasourceClient]
}

func (datasourceClients *DatasourceClients) QueryData(ctx context.Context, request *QueryDataRequest) *QueryDataResponse {
	response := &QueryDataResponse{
		Results: make(map[string]QueryDataResult),
	}

	client := datasourceClients.Get(request.Datasource.Type)
	if client == nil {
		for _, query := range request.Queries {
			response.Results[query.ID] = QueryDataResult{Error: external.ErrorNotSupportedDatasourceType.Error()}
		}
		return response
	}

	if client.DataClient != nil {
		return client.DataClient.QueryData(ctx, request)
	}

	for _, query := range request.Queries {
		response.Results[query.ID] = QueryDataResult{Error: external.ErrorNotSupportedDataServer.Error()}
	}
	return response
}

func (datasourceClients *DatasourceClients) RunStream(request *RunStreamRequest) *RunStreamResponse {
	response := &RunStreamResponse{}

	client := datasourceClients.Get(request.Datasource.Type)
	if client == nil {
		response.Error = external.ErrorNotSupportedDatasourceType.Error()
		return response
	}

	if client.StreamClient != nil {
		return client.StreamClient.RunStream(request)
	}

	response.Error = external.ErrorNotSupportedStreamServer.Error()
	return response
}

func (datasourceClients *DatasourceClients) SubscribeStream(request *SubscribeStreamRequest) *SubscribeStreamResponse {
	response := &SubscribeStreamResponse{}

	client := datasourceClients.Get(request.Datasource.Type)
	if client == nil {
		response.Error = external.ErrorNotSupportedDatasourceType.Error()
		return response
	}

	if client.StreamClient != nil {
		return client.StreamClient.SubscribeStream(request)
	}

	response.Error = external.ErrorNotSupportedStreamServer.Error()
	return response
}

func (datasourceClients *DatasourceClients) UnsubscribeStream(request *UnsubscribeStreamRequest) error {
	client := datasourceClients.Get(request.Datasource.Type)
	if client == nil {
		return external.ErrorNotSupportedDatasourceType
	}

	if client.StreamClient != nil {
		return client.StreamClient.UnsubscribeStream(request)
	}

	return external.ErrorNotSupportedStreamServer
}

func (datasourceClients *DatasourceClients) PublishStream(request *PublishStreamRequest) *PublishStreamResponse {
	response := &PublishStreamResponse{}

	client := datasourceClients.Get(request.Datasource.Type)
	if client == nil {
		response.Error = external.ErrorNotSupportedDatasourceType.Error()
		return response
	}

	if client.StreamClient != nil {
		return client.StreamClient.PublishStream(request)
	}

	response.Error = external.ErrorNotSupportedStreamServer.Error()
	return response
}

func (datasourceClients *DatasourceClients) RemoveDatasource(request *RemoveDatasourceRequest) error {
	client := datasourceClients.Get(request.Datasource.Type)
	if client == nil {
		return external.ErrorNotSupportedDatasourceType
	}

	if client.DataClient != nil {
		if err := client.DataClient.RemoveDatasource(request); err != nil {
			return err
		}
	}

	if client.StreamClient != nil {
		if err := client.StreamClient.RemoveDatasource(request); err != nil {
			return err
		}
	}

	return nil
}
