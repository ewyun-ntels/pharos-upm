package sql

import (
	"context"

	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/model"
)

type DataClient struct {
	Config common.Config
}

func (dataClient *DataClient) QueryData(ctx context.Context, request *model.QueryDataRequest) *model.QueryDataResponse {
	response := &model.QueryDataResponse{Results: map[string]model.QueryDataResult{}}

	for _, query := range request.Queries {
		if databaseResponse, err := request.DatabaseConfig.GetDatabaseResponse(ctx, query.SQL, query.Timeout); err != nil {
			response.Results[query.ID] = model.QueryDataResult{Error: err.Error()}
		} else {
			response.Results[query.ID] = model.QueryDataResult{Frame: databaseResponse}
		}
	}

	return response
}

func (dataClient *DataClient) RemoveDatasource(request *model.RemoveDatasourceRequest) error {
	return orm.DatabasePool.Remove(request.DatabaseConfig)
}
