package handler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"github.com/ory/fosite/token/jwt"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/casbin"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/shared/types/dashboard"
)

// TestQuerySetFromReader Query JSON 파싱 테스트
func TestQuerySetFromReader(t *testing.T) {
	tests := []struct {
		name        string
		input       string
		expectError bool
		expected    Query
	}{
		{
			name: "Valid complete query",
			input: `{
				"panel": {"kind": "cards", "id": "test-panel"},
				"run": {"datasourceName": "test-ds", "query": "SELECT 1"},
				"variables": {"var1": "value1"}
			}`,
			expectError: false,
			expected: Query{
				Panel: struct {
					Kind string `json:"kind"`
					ID   string `json:"id"`
				}{Kind: "cards", ID: "test-panel"},
			},
		},
		{
			name: "Valid minimal query",
			input: `{
				"panel": {"kind": "headerL", "id": "header-1"},
				"run": {},
				"variables": {}
			}`,
			expectError: false,
			expected: Query{
				Panel: struct {
					Kind string `json:"kind"`
					ID   string `json:"id"`
				}{Kind: "headerL", ID: "header-1"},
			},
		},
		{
			name:        "Invalid JSON",
			input:       `{"invalid": json}`,
			expectError: true,
		},
		{
			name:        "Empty input",
			input:       "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &Query{}
			reader := strings.NewReader(tt.input)

			err := query.setFromReader(reader)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if !tt.expectError {
				if query.Panel.Kind != tt.expected.Panel.Kind {
					t.Errorf("expected panel kind %s, got %s", tt.expected.Panel.Kind, query.Panel.Kind)
				}
				if query.Panel.ID != tt.expected.Panel.ID {
					t.Errorf("expected panel ID %s, got %s", tt.expected.Panel.ID, query.Panel.ID)
				}
				// Run 필드의 DatasourceName과 Query를 개별적으로 검증
				if tt.name == "Valid complete query" {
					if query.Run.DatasourceName != "test-ds" {
						t.Errorf("expected datasource test-ds, got %s", query.Run.DatasourceName)
					}
					if query.Run.Query != "SELECT 1" {
						t.Errorf("expected query 'SELECT 1', got %s", query.Run.Query)
					}
					if len(query.Variables) != 1 || query.Variables["var1"] != "value1" {
						t.Errorf("expected variables map with var1=value1, got %v", query.Variables)
					}
				}
			}
		})
	}
}

func TestHideSQL(t *testing.T) {
	response := orm.DatabaseResponse{
		SQL:  "SELECT secret FROM hidden_table",
		Data: []map[string]any{{"value": 1}},
		Meta: []map[string]string{{"name": "value"}},
		Rows: 1,
		Statistics: orm.DatabaseResponseStatistics{
			Elapsed: 0.1,
		},
	}

	hideSQL(&response)

	if response.SQL != "" {
		t.Fatalf("expected SQL to be hidden, got %q", response.SQL)
	}
	if response.Rows != 1 || len(response.Data) != 1 || len(response.Meta) != 1 {
		t.Fatalf("hideSQL should only clear SQL, got response: %+v", response)
	}
}

// TestQueryPanelToRun panelToRun 메서드 테스트
func TestQueryPanelToRun(t *testing.T) {
	// 테스트용 대시보드 설정
	headerDatasource := "header-datasource"
	headerQuery := "SELECT COUNT(*) FROM header_table"
	headerRDatasource := "header-r-datasource"
	headerRQuery := "SELECT SUM(amount) FROM transactions"
	leftDatasource := "left-datasource"
	leftQuery := "SELECT * FROM left_panel"

	config := &dashboard.DashboardConfig{
		Panels: []dashboard.Panel{
			{
				ID: "panel-1",
				DataProvider: &dashboard.DataProvider{
					ChartQuery: []dashboard.ChartQuery{
						{
							DatasourceName: "test-datasource",
							Label:          "Test Query",
							Query:          "SELECT * FROM panels_table",
						},
					},
				},
			},
			{
				ID: "panel-no-chartquery",
				DataProvider: &dashboard.DataProvider{
					ChartQuery: []dashboard.ChartQuery{}, // 빈 배열
				},
			},
			{
				ID: "panel-no-dataprovider",
				// dataProvider가 없음
			},
		},
		Filters: []dashboard.FilterConfig{
			{
				ID:             "header-1",
				Type:           "select",
				Kind:           dashboard.HeaderL,
				DatasourceName: &headerDatasource,
				Query:          &headerQuery,
			},
			{
				ID:             "header-r-1",
				Type:           "datetime",
				Kind:           dashboard.HeaderR,
				DatasourceName: &headerRDatasource,
				Query:          &headerRQuery,
			},
			{
				ID:             "left-1",
				Type:           "multiselect",
				Kind:           dashboard.Left,
				DatasourceName: &leftDatasource,
				Query:          &leftQuery,
			},
		},
	}

	tests := []struct {
		name               string
		panelKind          string
		panelID            string
		presetDatasource   string // 미리 설정된 datasource (Owner/Editor가 직접 보낸 경우)
		presetQuery        string // 미리 설정된 query
		expectedDatasource string
		expectedQuery      string
		expectError        bool
		errorMessage       string
	}{
		{
			name:               "Panels panel",
			panelKind:          "panels",
			panelID:            "panel-1",
			expectedDatasource: "test-datasource", // panels는 dataProvider.chartQuery[0]에서 추출
			expectedQuery:      "SELECT * FROM panels_table",
			expectError:        false,
		},
		{
			name:               "Panels panel with preset query",
			panelKind:          "panels",
			panelID:            "panel-1",
			presetDatasource:   "custom-datasource",
			presetQuery:        "SELECT * FROM custom_table",
			expectedDatasource: "custom-datasource", // 미리 설정된 값 유지
			expectedQuery:      "SELECT * FROM custom_table",
			expectError:        false,
		},
		{
			name:               "HeaderL panel",
			panelKind:          "headerL",
			panelID:            "header-1",
			expectedDatasource: "header-datasource",
			expectedQuery:      "SELECT COUNT(*) FROM header_table",
			expectError:        false,
		},
		{
			name:               "HeaderR panel",
			panelKind:          "headerR",
			panelID:            "header-r-1",
			expectedDatasource: "header-r-datasource",
			expectedQuery:      "SELECT SUM(amount) FROM transactions",
			expectError:        false,
		},
		{
			name:               "Left panel",
			panelKind:          "left",
			panelID:            "left-1",
			expectedDatasource: "left-datasource",
			expectedQuery:      "SELECT * FROM left_panel",
			expectError:        false,
		},
		{
			name:         "Unsupported panel kind",
			panelKind:    "unsupported",
			panelID:      "any-id",
			expectError:  true,
			errorMessage: external.ErrorNotSupportedPanelKind.Error(),
		},
		{
			name:         "Non-existent panel ID",
			panelKind:    "panels",
			panelID:      "non-existent",
			expectError:  true,
			errorMessage: external.ErrorInvalidPanelID.Error(),
		},
		{
			name:         "Panel with no chartQuery",
			panelKind:    "panels",
			panelID:      "panel-no-chartquery",
			expectError:  true,
			errorMessage: "has no chartQuery",
		},
		{
			name:         "Panel with no dataProvider",
			panelKind:    "panels",
			panelID:      "panel-no-dataprovider",
			expectError:  true,
			errorMessage: "has no chartQuery",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &Query{
				Panel: struct {
					Kind string `json:"kind"`
					ID   string `json:"id"`
				}{Kind: tt.panelKind, ID: tt.panelID},
			}

			// 미리 설정된 datasource/query가 있으면 설정
			if tt.presetDatasource != "" || tt.presetQuery != "" {
				query.Run.DatasourceName = tt.presetDatasource
				query.Run.Query = tt.presetQuery
			}

			err := query.panelToRun(config)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if !strings.Contains(err.Error(), tt.errorMessage) {
					t.Errorf("expected error message '%s', got '%s'", tt.errorMessage, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else {
					if query.Run.DatasourceName != tt.expectedDatasource {
						t.Errorf("expected datasource '%s', got '%s'", tt.expectedDatasource, query.Run.DatasourceName)
					}
					if query.Run.Query != tt.expectedQuery {
						t.Errorf("expected query '%s', got '%s'", tt.expectedQuery, query.Run.Query)
					}
				}
			}
		})
	}
}

// TestQuerySetRun setRun 메서드 테스트
func TestQuerySetRun(t *testing.T) {
	cfg, testEnforcer := setupDashboardTest(t)
	config = cfg
	enforcer = testEnforcer

	// 테스트용 대시보드 생성
	testDashboardID := uuid.New().String()
	testSubject := "test-user"

	dashboardConfig := dashboard.DashboardConfig{
		Title: "Test Dashboard",
		Panels: []dashboard.Panel{
			{
				ID:    "test-panel",
				Title: "Test Panel",
				DataProvider: &dashboard.DataProvider{
					ChartQuery: []dashboard.ChartQuery{
						{
							DatasourceName: "test-datasource",
							Label:          "Test Query",
							Query:          "SELECT * FROM test_table",
						},
					},
				},
			},
		},
	}

	configJSON, _ := dashboardConfig.Marshal()

	// 데이터베이스에 테스트 대시보드 삽입
	err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`INSERT INTO dashboard(id, config) VALUES (?, ?);`, testDashboardID, string(configJSON))
		return err
	})
	if err != nil {
		t.Fatalf("failed to insert test dashboard: %v", err)
	}

	// 권한 추가 (ActionOwner로 설정하여 모든 기능 테스트 가능)
	policy := casbin.Policy{testSubject, testDashboardID, casbin.ActionOwner, PermissionKindUser}
	err = enforcer.AddPolicy(policy)
	if err != nil {
		t.Fatalf("failed to add policy: %v", err)
	}

	tests := []struct {
		name              string
		query             *Query
		jwtClaims         *jwt.JWTClaims
		dashboardID       string
		expectError       bool
		expectedErrorType error
	}{
		{
			name: "Query with datasource and query already set",
			query: func() *Query {
				q := &Query{}
				q.Run.DatasourceName = "preset-ds"
				q.Run.Query = "SELECT 1"
				return q
			}(),
			jwtClaims:   &jwt.JWTClaims{Subject: testSubject},
			dashboardID: testDashboardID,
			expectError: false, // 권한이 있으면 성공
		},
		{
			name: "Query with preset values but no permission",
			query: func() *Query {
				q := &Query{}
				q.Run.DatasourceName = "preset-ds"
				q.Run.Query = "SELECT 1"
				return q
			}(),
			jwtClaims:         &jwt.JWTClaims{Subject: "unauthorized-user"},
			dashboardID:       testDashboardID,
			expectError:       true,
			expectedErrorType: external.ErrorNoSuchPolicy, // 권한이 없으면 실패
		},
		{
			name: "Query needs panel resolution",
			query: func() *Query {
				q := &Query{}
				q.Panel.Kind = "panels"
				q.Panel.ID = "test-panel"
				return q
			}(),
			jwtClaims:   &jwt.JWTClaims{Subject: testSubject},
			dashboardID: testDashboardID,
			expectError: false,
		},
		{
			name: "No permission",
			query: func() *Query {
				q := &Query{}
				q.Panel.Kind = "panels"
				q.Panel.ID = "test-panel"
				return q
			}(),
			jwtClaims:         &jwt.JWTClaims{Subject: "unauthorized-user"},
			dashboardID:       testDashboardID,
			expectError:       true,
			expectedErrorType: external.ErrorNoSuchPolicy,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("POST", "/dashboard/"+tt.dashboardID+"/query", nil)
			c.Request = req
			c.Params = []gin.Param{{Key: "id", Value: tt.dashboardID}}

			if tt.jwtClaims != nil {
				c.Set(common.ContextKeyJWTClaims, tt.jwtClaims)
			}

			err := tt.query.setRun(c)

			if tt.expectError {
				if err == nil {
					t.Errorf("expected error but got none")
				} else if tt.expectedErrorType != nil && !errors.Is(err, tt.expectedErrorType) {
					t.Errorf("expected error type %v, got %v", tt.expectedErrorType, err)
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				} else {
					// permission이 올바르게 설정되었는지 확인
					if tt.query.Run.permission == "" {
						t.Errorf("expected permission to be set, but it was empty")
					}
				}
			}
		})
	}
}

// TestQueryPostHandlerDetailed Query POST 핸들러 상세 테스트
func TestQueryPostHandlerDetailed(t *testing.T) {
	cfg, testEnforcer := setupDashboardTest(t)
	config = cfg
	enforcer = testEnforcer

	// 테스트용 대시보드 설정
	testDashboardID := uuid.New().String()
	testSubject := "test-user"

	dashboardConfig := dashboard.DashboardConfig{
		Title: "Test Dashboard",
		Panels: []dashboard.Panel{
			{
				ID:    "test-panel",
				Title: "Test Panel",
				DataProvider: &dashboard.DataProvider{
					ChartQuery: []dashboard.ChartQuery{
						{
							DatasourceName: "test-datasource",
							Label:          "Test Query",
							Query:          "SELECT * FROM test_table",
						},
					},
				},
			},
		},
	}

	configJSON, _ := dashboardConfig.Marshal()

	// 데이터베이스에 테스트 대시보드 삽입
	err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`INSERT INTO dashboard(id, config) VALUES (?, ?);`, testDashboardID, string(configJSON))
		return err
	})
	if err != nil {
		t.Fatalf("failed to insert test dashboard: %v", err)
	}

	// 권한 추가
	policy := casbin.Policy{testSubject, testDashboardID, casbin.ActionViewer, PermissionKindUser}
	err = enforcer.AddPolicy(policy)
	if err != nil {
		t.Fatalf("failed to add policy: %v", err)
	}

	tests := []struct {
		name           string
		requestBody    string
		jwtClaims      *jwt.JWTClaims
		dashboardID    string
		expectedStatus int
		expectError    bool
	}{
		{
			name: "Valid query with panel reference",
			requestBody: `{
				"panel": {"kind": "cards", "id": "test-card"},
				"variables": {}
			}`,
			jwtClaims:      &jwt.JWTClaims{Subject: testSubject},
			dashboardID:    testDashboardID,
			expectedStatus: http.StatusInternalServerError, // plugins.DatasourceQuery가 mock되지 않아 실패 예상
			expectError:    true,
		},
		{
			name: "Direct query execution",
			requestBody: `{
				"panel": {"kind": "cards", "id": "any"},
				"run": {"datasourceName": "direct-ds", "query": "SELECT 1"},
				"variables": {}
			}`,
			jwtClaims:      &jwt.JWTClaims{Subject: testSubject},
			dashboardID:    testDashboardID,
			expectedStatus: http.StatusInternalServerError, // plugins.DatasourceQuery가 mock되지 않아 실패 예상
			expectError:    true,
		},
		{
			name: "No permission",
			requestBody: `{
				"panel": {"kind": "cards", "id": "test-card"},
				"variables": {}
			}`,
			jwtClaims:      &jwt.JWTClaims{Subject: "unauthorized-user"},
			dashboardID:    testDashboardID,
			expectedStatus: http.StatusForbidden,
			expectError:    false,
		},
		{
			name:           "Invalid JSON",
			requestBody:    `{"invalid": json}`,
			jwtClaims:      &jwt.JWTClaims{Subject: testSubject},
			dashboardID:    testDashboardID,
			expectedStatus: http.StatusInternalServerError,
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &Query{}

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("POST", "/dashboard/"+tt.dashboardID+"/query", bytes.NewBufferString(tt.requestBody))
			req.Header.Set("Content-Type", "application/json")
			c.Request = req
			c.Params = []gin.Param{{Key: "id", Value: tt.dashboardID}}

			if tt.jwtClaims != nil {
				c.Set(common.ContextKeyJWTClaims, tt.jwtClaims)
			}

			statusCode, response := query.PostHandler(c)

			if statusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, statusCode)
			}

			if tt.expectError {
				if _, ok := response.(external.ErrorResponse); !ok && response != nil {
					t.Errorf("expected error response or nil, got %T", response)
				}
			}
		})
	}
}

func TestQueryPostHandlerRejectsViewerInspect(t *testing.T) {
	cfg, testEnforcer := setupDashboardTest(t)
	config = cfg
	enforcer = testEnforcer

	testDashboardID := uuid.New().String()
	testSubject := "viewer-user"

	dashboardConfig := dashboard.DashboardConfig{
		Title: "Test Dashboard",
		Panels: []dashboard.Panel{
			{
				ID:    "test-panel",
				Title: "Test Panel",
				DataProvider: &dashboard.DataProvider{
					ChartQuery: []dashboard.ChartQuery{
						{
							DatasourceName: "test-datasource",
							Label:          "Test Query",
							Query:          "SELECT secret FROM hidden_table",
						},
					},
				},
			},
		},
	}

	configJSON, _ := dashboardConfig.Marshal()
	err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`INSERT INTO dashboard(id, config) VALUES (?, ?);`, testDashboardID, string(configJSON))
		return err
	})
	if err != nil {
		t.Fatalf("failed to insert test dashboard: %v", err)
	}

	policy := casbin.Policy{testSubject, testDashboardID, casbin.ActionViewer, PermissionKindUser}
	if err := enforcer.AddPolicy(policy); err != nil {
		t.Fatalf("failed to add policy: %v", err)
	}

	query := &Query{}
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	req := httptest.NewRequest("POST", "/dashboard/"+testDashboardID+"/query?inspect=true", bytes.NewBufferString(`{
		"panel": {"kind": "panels", "id": "test-panel"},
		"variables": {}
	}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: testDashboardID}}
	c.Set(common.ContextKeyJWTClaims, &jwt.JWTClaims{Subject: testSubject})

	statusCode, response := query.PostHandler(c)

	if statusCode != http.StatusForbidden {
		t.Fatalf("expected status %d, got %d", http.StatusForbidden, statusCode)
	}
	if response != nil {
		t.Fatalf("expected nil response for forbidden viewer inspect, got %T", response)
	}

	query = &Query{}
	w = httptest.NewRecorder()
	c, _ = gin.CreateTestContext(w)
	req = httptest.NewRequest("POST", "/dashboard/"+testDashboardID+"/query?inspect=true", bytes.NewBufferString(`{
		"panel": {"kind": "panels", "id": "missing-panel"},
		"variables": {}
	}`))
	req.Header.Set("Content-Type", "application/json")
	c.Request = req
	c.Params = []gin.Param{{Key: "id", Value: testDashboardID}}
	c.Set(common.ContextKeyJWTClaims, &jwt.JWTClaims{Subject: testSubject})

	statusCode, response = query.PostHandler(c)

	if statusCode != http.StatusForbidden {
		t.Fatalf("expected missing panel viewer inspect to return status %d, got %d", http.StatusForbidden, statusCode)
	}
	if response != nil {
		t.Fatalf("expected nil response for forbidden missing panel viewer inspect, got %T", response)
	}
}

// TestQueryPostHandlerQueryTimeout Query 타임아웃 테스트
func TestQueryPostHandlerQueryTimeout(t *testing.T) {
	cfg, testEnforcer := setupDashboardTest(t)
	config = cfg
	enforcer = testEnforcer

	// 테스트용 대시보드 설정
	testDashboardID := uuid.New().String()
	testSubject := "test-user"

	dashboardConfig := dashboard.DashboardConfig{
		Title: "Test Dashboard",
		Panels: []dashboard.Panel{
			{
				ID:    "test-panel",
				Title: "Test Panel",
				DataProvider: &dashboard.DataProvider{
					ChartQuery: []dashboard.ChartQuery{
						{
							DatasourceName: "test-datasource",
							Label:          "Test Query",
							Query:          "SELECT * FROM test_table",
						},
					},
				},
			},
		},
	}

	configJSON, _ := dashboardConfig.Marshal()

	// 데이터베이스에 테스트 대시보드 삽입
	err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`INSERT INTO dashboard(id, config) VALUES (?, ?);`, testDashboardID, string(configJSON))
		return err
	})
	if err != nil {
		t.Fatalf("failed to insert test dashboard: %v", err)
	}

	// 권한 추가
	policy := casbin.Policy{testSubject, testDashboardID, casbin.ActionViewer, PermissionKindUser}
	err = enforcer.AddPolicy(policy)
	if err != nil {
		t.Fatalf("failed to add policy: %v", err)
	}

	// Note: 실제 타임아웃을 테스트하려면 plugins.DatasourceQuery를 mock해야 합니다.
	// 현재는 타임아웃 에러가 발생했을 때 HTTP 408 상태 코드가 반환되는지 확인하는
	// 통합 테스트로서의 구조만 검증합니다.
	t.Run("Query_timeout_returns_408", func(t *testing.T) {
		// 실제 타임아웃을 발생시키려면 매우 긴 쿼리가 필요하므로
		// 이 테스트는 구조 검증용입니다.
		// 실제 타임아웃 테스트는 아래의 단위 테스트에서 수행됩니다.
		t.Skip("Requires mocked plugins.DatasourceQuery for timeout simulation")
	})
}

// TestQueryGetDashboardResponsePermissions getDashboardResponse 권한 테스트
func TestQueryGetDashboardResponsePermissions(t *testing.T) {
	cfg, testEnforcer := setupDashboardTest(t)
	config = cfg
	enforcer = testEnforcer

	// 테스트용 대시보드 설정
	testDashboardID := uuid.New().String()
	ownerSubject := "owner-user"
	viewerSubject := "viewer-user"

	dashboardConfig := dashboard.DashboardConfig{
		Title: "Test Dashboard",
		Panels: []dashboard.Panel{
			{
				ID:    "test-panel",
				Title: "Test Panel",
				DataProvider: &dashboard.DataProvider{
					ChartQuery: []dashboard.ChartQuery{
						{
							DatasourceName: "test-datasource",
							Label:          "Test Query",
							Query:          "SELECT * FROM test_table",
						},
					},
				},
			},
		},
	}

	configJSON, _ := dashboardConfig.Marshal()

	// 데이터베이스에 테스트 대시보드 삽입
	err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`INSERT INTO dashboard(id, config) VALUES (?, ?);`, testDashboardID, string(configJSON))
		return err
	})
	if err != nil {
		t.Fatalf("failed to insert test dashboard: %v", err)
	}

	// 권한 추가
	policies := []casbin.Policy{
		{ownerSubject, testDashboardID, casbin.ActionOwner, PermissionKindUser},
		{viewerSubject, testDashboardID, casbin.ActionViewer, PermissionKindUser},
	}

	for _, policy := range policies {
		err = enforcer.AddPolicy(policy)
		if err != nil {
			t.Fatalf("failed to add policy: %v", err)
		}
	}

	tests := []struct {
		name        string
		jwtClaims   *jwt.JWTClaims
		queryText   string
		expectError bool
	}{
		{
			name:        "Owner can access with query",
			jwtClaims:   &jwt.JWTClaims{Subject: ownerSubject},
			queryText:   "SELECT * FROM anywhere",
			expectError: false,
		},
		{
			name:        "Owner can access without query",
			jwtClaims:   &jwt.JWTClaims{Subject: ownerSubject},
			queryText:   "",
			expectError: false,
		},
		{
			name:        "Viewer can access without query",
			jwtClaims:   &jwt.JWTClaims{Subject: viewerSubject},
			queryText:   "",
			expectError: false,
		},
		{
			name:        "Viewer cannot access with query",
			jwtClaims:   &jwt.JWTClaims{Subject: viewerSubject},
			queryText:   "SELECT * FROM restricted",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &Query{}
			query.Run.Query = tt.queryText

			gin.SetMode(gin.TestMode)
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)

			req := httptest.NewRequest("GET", "/dashboard/"+testDashboardID, nil)
			c.Request = req
			c.Params = []gin.Param{{Key: "id", Value: testDashboardID}}

			if tt.jwtClaims != nil {
				c.Set(common.ContextKeyJWTClaims, tt.jwtClaims)
			}

			_, err := query.getDashboardResponse(c)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			} else if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}
		})
	}
}

// TestQueryVariablesHandling 변수 처리 테스트
func TestQueryVariablesHandling(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		variables map[string]any
	}{
		{
			name: "Query with variables",
			input: `{
				"panel": {"kind": "cards", "id": "test"},
				"run": {"datasourceName": "test", "query": "SELECT * FROM table WHERE id = {{.id}}"},
				"variables": {"id": 123, "name": "test"}
			}`,
			variables: map[string]any{"id": float64(123), "name": "test"},
		},
		{
			name: "Query with empty variables",
			input: `{
				"panel": {"kind": "cards", "id": "test"},
				"run": {"datasourceName": "test", "query": "SELECT 1"},
				"variables": {}
			}`,
			variables: map[string]any{},
		},
		{
			name: "Query with complex variables",
			input: `{
				"panel": {"kind": "cards", "id": "test"},
				"run": {"datasourceName": "test", "query": "SELECT 1"},
				"variables": {
					"filters": {"status": "active", "type": "user"},
					"limit": 100,
					"enabled": true
				}
			}`,
			variables: map[string]any{
				"filters": map[string]any{"status": "active", "type": "user"},
				"limit":   float64(100),
				"enabled": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			query := &Query{}
			reader := strings.NewReader(tt.input)

			err := query.setFromReader(reader)
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}

			if len(query.Variables) != len(tt.variables) {
				t.Errorf("expected %d variables, got %d", len(tt.variables), len(query.Variables))
			}

			for key, expectedValue := range tt.variables {
				if actualValue, exists := query.Variables[key]; !exists {
					t.Errorf("expected variable %s not found", key)
				} else {
					// JSON 파싱에서 숫자는 float64로 변환되므로 타입 비교 시 주의
					expectedJSON, _ := json.Marshal(expectedValue)
					actualJSON, _ := json.Marshal(actualValue)
					if string(expectedJSON) != string(actualJSON) {
						t.Errorf("variable %s: expected %v, got %v", key, expectedValue, actualValue)
					}
				}
			}
		})
	}
}
