package plugins

import (
	"context"

	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/model"
)

var config common.Config

var provisioningDatasources = internal.NewMap[*Datasource]()

func DatasourceQuery(ctx context.Context, dsQueryRequest DsQueryRequest) *DsQueryResponse {
	dsQueryResponse := DsQueryResponse{
		Results: map[string]model.QueryDataResult{},
	}

	dsQueryResponse.set(ctx, dsQueryRequest)

	return &dsQueryResponse
}
