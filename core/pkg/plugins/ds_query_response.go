package plugins

import (
	"context"
	"database/sql"
	"errors"
	"maps"
	"sync"

	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/pkg/plugins/model"
	"ntels.com/pharos/core/pkg/plugins/shared"
)

type DsQueryResponse struct {
	Results map[string]model.QueryDataResult `json:"Results"`
	mu      sync.RWMutex
}

func (dsQueryResponse *DsQueryResponse) set(ctx context.Context, dsQueryRequest DsQueryRequest) {
	queryDataRequests := dsQueryResponse.makeQueryDataRequests(dsQueryRequest)

	wg := new(sync.WaitGroup)
	for id := range queryDataRequests {
		wg.Add(1)
		go func(queryDataRequest *model.QueryDataRequest) {
			defer wg.Done()

			response := shared.DatasourceClients.QueryData(ctx, queryDataRequest)

			// context가 취소된 경우 결과를 쓰지 않음 (caller가 이미 반환했을 수 있음)
			select {
			case <-ctx.Done():
				return
			default:
			}

			dsQueryResponse.mu.Lock()
			defer dsQueryResponse.mu.Unlock()
			maps.Copy(dsQueryResponse.Results, response.Results)
		}(queryDataRequests[id])
	}

	// context 취소 시 wg.Wait()을 건너뛰어 HTTP 연결을 즉시 해제
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
	case <-ctx.Done():
		// 클라이언트 연결 해제 - 쿼리 결과를 기다리지 않고 즉시 반환
		// 백그라운드 goroutine은 계속 실행되다 DB 타임아웃 시 종료
	}
}

func (dsQueryResponse *DsQueryResponse) makeQueryDataRequests(dsQueryRequest DsQueryRequest) map[string]*model.QueryDataRequest {
	queryDataRequests := map[string]*model.QueryDataRequest{}

	for _, query := range dsQueryRequest.Queries {
		if request, exist := queryDataRequests[query.DatasourceName]; exist {
			request.Queries = append(request.Queries, model.Query{ID: query.ID, SQL: query.SQL, Timeout: query.Timeout})
			continue
		}

		datasource := Datasource{}

		if provisioningDatasources.Exist(query.DatasourceName) {
			datasource = *provisioningDatasources.Get(query.DatasourceName)
		} else if err := datasource.SetFromDB(query.DatasourceName); errors.Is(err, sql.ErrNoRows) {
			dsQueryResponse.mu.Lock()
			dsQueryResponse.Results[query.ID] = model.QueryDataResult{Error: external.ErrorNotExistDatasource.Error()}
			dsQueryResponse.mu.Unlock()
			continue
		} else if err != nil {
			dsQueryResponse.mu.Lock()
			dsQueryResponse.Results[query.ID] = model.QueryDataResult{Error: err.Error()}
			dsQueryResponse.mu.Unlock()
			continue
		}

		databaseConfig, err := datasource.ToDatabaseConfig(config)
		if err != nil {
			dsQueryResponse.mu.Lock()
			dsQueryResponse.Results[query.ID] = model.QueryDataResult{Error: err.Error()}
			dsQueryResponse.mu.Unlock()
			continue
		}

		queryDataRequests[query.DatasourceName] = &model.QueryDataRequest{
			Datasource:     datasource.Datasource,
			DatabaseConfig: databaseConfig,
			Queries:        []model.Query{{ID: query.ID, SQL: query.SQL, Timeout: query.Timeout}},
		}
	}

	return queryDataRequests
}
