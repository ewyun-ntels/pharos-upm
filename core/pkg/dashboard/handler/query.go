package handler

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/casbin"
	"ntels.com/pharos/core/internal/query_builder"
	"ntels.com/pharos/core/pkg/plugins"
	"ntels.com/pharos/shared/types/dashboard"
)

type Query struct {
	Panel struct {
		Kind string `json:"kind"`
		ID   string `json:"id"`
	} `json:"panel"`

	Run struct {
		DatasourceName string `json:"datasourceName"`
		Query          string `json:"query"`
		Timeout        int    `json:"timeout"`

		permission string
	} `json:"run"`

	Variables map[string]any `json:"variables"`
}

func (query *Query) PostHandler(c *gin.Context) (int, any) {
	if err := query.setFromReader(c.Request.Body); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	if err := query.setRun(c); err != nil {
		if errors.Is(err, external.ErrorNoSuchPolicy) {
			return http.StatusForbidden, nil
		}
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	// Inspect 모드 확인
	inspectMode := c.Query("inspect") == "true"

	if inspectMode {
		// Inspect 응답 생성
		if statusCode, inspectResponse, err := query.getInspectResponse(c); err != nil {
			slog.Error("Failed to get inspect response", "error", err)
			return statusCode, external.ErrorResponse{Message: err.Error()}
		} else {
			return statusCode, inspectResponse
		}
	}

	// 기존 응답 (하위 호환성)
	if statusCode, databaseResponse, err := query.getPostResponse(c); err != nil {
		if !errors.Is(err, context.Canceled) {
			slog.Error("Failed to get database response", "error", err)
		}

		errorMessage := err.Error()
		if query.Run.permission == casbin.ActionViewer {
			switch statusCode {
			case http.StatusRequestTimeout:
				errorMessage = "query timeout"
			default:
				errorMessage = "query error"
			}
		}

		return statusCode, external.ErrorResponse{Message: errorMessage}
	} else {
		return statusCode, databaseResponse
	}
}

func (query *Query) getPostResponse(ctx context.Context) (int, orm.DatabaseResponse, error) {
	if sql, err := query_builder.GetQuery(query.Run.Query, query.Variables); err != nil {
		return http.StatusInternalServerError, orm.DatabaseResponse{}, err
	} else if databaseResponse, err := query.getDatabaseResponse(ctx, string(sql)); errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return http.StatusRequestTimeout, orm.DatabaseResponse{}, err
	} else if err != nil {
		return http.StatusInternalServerError, orm.DatabaseResponse{}, err
	} else {
		hideSQL(&databaseResponse)
		return http.StatusOK, databaseResponse, nil
	}
}

func hideSQL(response *orm.DatabaseResponse) {
	response.SQL = ""
}

// getInspectResponse returns extended response with debug information for query inspection
func (query *Query) getInspectResponse(ctx context.Context) (int, dashboard.QueryInspectResponse, error) {
	// Variable 치환 전 원본 쿼리 저장
	rawQuery := query.Run.Query

	// Variable 치환
	executedSQL, err := query_builder.GetQuery(query.Run.Query, query.Variables)
	if err != nil {
		return http.StatusInternalServerError, dashboard.QueryInspectResponse{}, err
	}

	// Datasource 실행
	databaseResponse, queryErr := query.getDatabaseResponse(ctx, string(executedSQL))

	// Inspect 정보 구성
	inspectInfo := dashboard.InspectInfo{
		RawQuery:       rawQuery,
		ExecutedQuery:  string(executedSQL),
		Variables:      query.Variables,
		DatasourceName: query.Run.DatasourceName,
		Timestamp:      time.Now(),
	}

	// QueryInspectResponse 생성 (quicktype 생성 타입)
	inspectResponse := dashboard.QueryInspectResponse{
		Data: databaseResponse.Data,
		Meta: databaseResponse.Meta,
		Rows: databaseResponse.Rows,
		Statistics: dashboard.QueryInspectResponseStatistics{
			Elapsed: databaseResponse.Statistics.Elapsed,
		},
		Inspect: inspectInfo,
	}

	if queryErr != nil {
		// 에러 발생 시에도 Inspect 정보는 반환 (디버깅 용도)
		errorMsg := queryErr.Error()
		inspectResponse.Inspect.Error = &errorMsg

		if errors.Is(queryErr, context.DeadlineExceeded) {
			return http.StatusRequestTimeout, inspectResponse, nil
		}
		return http.StatusInternalServerError, inspectResponse, nil
	}

	return http.StatusOK, inspectResponse, nil
}

func (query *Query) setFromReader(reader io.Reader) error {
	if body, err := io.ReadAll(reader); err != nil {
		return err
	} else if err := json.Unmarshal(body, query); err != nil {
		return err
	}

	if query.Run.Timeout == 0 {
		// timeout이 0일 경우 초기화 수행
		query.Run.Timeout = 60
	}

	return nil
}

func (query *Query) getDashboardResponse(c *gin.Context) (map[string]any, error) {
	check := func(action string) bool {
		if action == casbin.ActionViewer && len(query.Run.Query) != 0 {
			return false
		}

		return true
	}

	id := c.Param("id")
	jwtClaims := getJWTClaims(c)
	if response, err := svc.makeGetResponse(id, jwtClaims, check, false); err != nil {
		return nil, err
	} else {
		return response, nil
	}
}

// ewyun-20260514: variables map에서 Prometheus QueryRange용 시간 범위 추출
func (query *Query) extractTimeRange() (orm.PrometheusTimeRange, bool) {
	toInt64 := func(key string) (int64, bool) {
		v, ok := query.Variables[key]
		if !ok {
			return 0, false
		}
		val, ok := v.(float64)
		return int64(val), ok
	}

	start, ok1 := toInt64("__start_time")
	end, ok2 := toInt64("__end_time")
	step, ok3 := toInt64("__step")

	// ewyun-20260514: use QueryRange only when step is positive; variable preview sends __step=0.
	if ok1 && ok2 && ok3 && start > 0 && end > start && step > 0 {
		return orm.PrometheusTimeRange{
			StartTime: start,
			EndTime:   end,
			Step:      int(step),
		}, true
	}
	return orm.PrometheusTimeRange{}, false
}

func (query *Query) getDatabaseResponse(ctx context.Context, sql string) (orm.DatabaseResponse, error) {
	const id = "A"

	// ewyun-20260514: Prometheus QueryRange를 위해 시간 범위를 context에 주입
	if tr, ok := query.extractTimeRange(); ok {
		ctx = orm.WithPrometheusTimeRange(ctx, tr)
	}

	response := plugins.DatasourceQuery(ctx, plugins.DsQueryRequest{
		Queries: []plugins.DsQuery{{ID: id, DatasourceName: query.Run.DatasourceName, SQL: sql, Timeout: query.Run.Timeout}},
	})

	// context 취소 확인 (클라이언트 연결 해제 or 타임아웃)
	if err := ctx.Err(); err != nil {
		return orm.DatabaseResponse{}, err
	}

	if response.Results[id].Error == context.DeadlineExceeded.Error() {
		return orm.DatabaseResponse{}, context.DeadlineExceeded
	} else if response.Results[id].Error == context.Canceled.Error() {
		return orm.DatabaseResponse{}, context.Canceled
	} else if len(response.Results[id].Error) != 0 {
		return orm.DatabaseResponse{}, errors.New(response.Results[id].Error)
	}

	return response.Results[id].Frame, nil
}

func (query *Query) setRun(c *gin.Context) error {
	dashboardResponse, err := query.getDashboardResponse(c)
	if err != nil {
		return err
	}

	permissionValue, exists := dashboardResponse["permission"]
	if !exists {
		return external.ErrorMissingPermission
	}
	permission, ok := permissionValue.(string)
	if !ok {
		return external.ErrorInvalidPermissionType
	}
	query.Run.permission = permission
	if c.Query("inspect") == "true" && permission == casbin.ActionViewer {
		return external.ErrorNoSuchPolicy
	}

	// Owner/Editor가 직접 query를 보낸 경우
	if len(query.Run.DatasourceName) != 0 && len(query.Run.Query) != 0 {
		return nil
	}

	// Panel reference 방식: 원본 config에서 query 추출 (hide 안 된 버전 사용)
	configValue, exists := dashboardResponse["config"]
	if !exists || configValue == nil {
		return external.ErrorMissingConfig
	}

	// config는 *dashboard.DashboardConfig 타입으로 반환됨
	dashboardConfig, ok := configValue.(*dashboard.DashboardConfig)
	if !ok {
		return external.ErrorInvalidConfigType
	}
	return query.panelToRun(dashboardConfig)
}

func (query *Query) panelToRun(dashboardConfig *dashboard.DashboardConfig) error {
	switch query.Panel.Kind {
	case "panels":
		// panels의 경우 dataProvider.chartQuery 배열에서 datasourceName과 query 추출
		for _, panel := range dashboardConfig.Panels {
			if panel.ID == query.Panel.ID {
				// query.Run에 이미 datasourceName과 query가 설정되어 있으면 사용
				// (Frontend에서 Owner/Editor가 직접 query를 보낸 경우)
				if query.Run.DatasourceName != "" && query.Run.Query != "" {
					slog.Info("Using preset query", "panelId", query.Panel.ID, "datasource", query.Run.DatasourceName)
					return nil
				}

				// 그렇지 않으면 panel의 dataProvider.chartQuery[0]에서 추출
				slog.Info("Panel found, extracting chartQuery", "panelId", panel.ID, "hasDataProvider", panel.DataProvider != nil, "chartQueryLength", func() int {
					if panel.DataProvider != nil {
						return len(panel.DataProvider.ChartQuery)
					}
					return 0
				}())

				if panel.DataProvider != nil && len(panel.DataProvider.ChartQuery) > 0 {
					chartQuery := panel.DataProvider.ChartQuery[0]
					slog.Info("ChartQuery extracted", "datasource", chartQuery.DatasourceName, "queryLen", len(chartQuery.Query))
					query.Run.DatasourceName = chartQuery.DatasourceName
					query.Run.Query = chartQuery.Query
					return nil
				}

				// chartQuery가 없으면 권한 없음 (Viewer) 또는 잘못된 패널 구성
				slog.Error("Panel has no chartQuery", "panelId", query.Panel.ID)
				return fmt.Errorf("panel %s has no chartQuery", query.Panel.ID)
			}
		}
	// Filter panels (replaces headerL, headerR, left)
	case "headerL", "headerR", "left":
		for _, filter := range dashboardConfig.Filters {
			// Check if filter ID matches and kind matches the requested kind
			if filter.ID == query.Panel.ID && string(filter.Kind) == query.Panel.Kind {
				if filter.DatasourceName != nil {
					query.Run.DatasourceName = *filter.DatasourceName
				}
				if filter.Query != nil {
					query.Run.Query = *filter.Query
				}
				return nil
			}
		}
	default:
		return external.ErrorNotSupportedPanelKind
	}

	return external.ErrorInvalidPanelID
}
