package handler

import (
	"bytes"
	"encoding/json"
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

// setupDashboardTest Dashboard 테스트를 위한 설정
func setupDashboardTest(t *testing.T) (common.Config, *casbin.Enforcer) {
	t.Helper()

	// orm.DriverSqlite를 사용한 기본 테스트 환경 설정
	cfg, testEnforcer := setupTestEnvironment(t)

	// Dashboard 테스트에 필요한 추가 테이블 설정
	err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		// dashboard 테이블이 없으면 생성 (setupTestEnvironment에서 이미 생성되지만 확실히 하기 위해)
		_, err := db.Exec(`
			CREATE TABLE IF NOT EXISTS dashboard (
				id TEXT PRIMARY KEY,
				config TEXT NOT NULL,
				folder_id TEXT,
				annotations TEXT NOT NULL DEFAULT '[]'
			);
		`)
		return err
	})
	if err != nil {
		t.Fatalf("failed to ensure dashboard table exists: %v", err)
	}

	return cfg, testEnforcer
}

// createTestContext 테스트용 Gin 컨텍스트 생성
func createTestContext(method, path string, body []byte, jwtClaims *jwt.JWTClaims) (*gin.Context, *httptest.ResponseRecorder) {
	gin.SetMode(gin.TestMode)

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)

	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}

	c.Request = req

	// JWT claims 설정
	if jwtClaims != nil {
		c.Set(common.ContextKeyJWTClaims, jwtClaims)
	}

	return c, w
}

// TestDashboardPostHandler POST 핸들러 테스트
func TestDashboardPostHandler(t *testing.T) {
	cfg, testEnforcer := setupDashboardTest(t)
	config = cfg
	enforcer = testEnforcer

	tests := []struct {
		name           string
		jwtClaims      *jwt.JWTClaims
		requestBody    string
		expectedStatus int
		expectError    bool
	}{
		{
			name:           "Empty subject",
			jwtClaims:      &jwt.JWTClaims{Subject: ""},
			requestBody:    `{"title": "Test Dashboard"}`,
			expectedStatus: http.StatusBadRequest,
			expectError:    true,
		},
		{
			name:           "Valid request",
			jwtClaims:      &jwt.JWTClaims{Subject: "test-user"},
			requestBody:    `{"title": "Test Dashboard", "panels": []}`,
			expectedStatus: http.StatusOK,
			expectError:    false,
		},
		{
			name:           "Invalid JSON",
			jwtClaims:      &jwt.JWTClaims{Subject: "test-user"},
			requestBody:    `{"invalid": json}`,
			expectedStatus: http.StatusBadRequest, // JSON 파싱 실패시 BadRequest 반환
			expectError:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc = NewDashboardService(config, enforcer)

			c, w := createTestContext("POST", "/dashboard", []byte(tt.requestBody), tt.jwtClaims)

			statusCode, response := svc.PostHandler(c)

			if statusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, statusCode)
			}

			if tt.expectError {
				if errorResp, ok := response.(external.ErrorResponse); !ok {
					t.Errorf("expected error response, got %T", response)
				} else if errorResp.Message == "" {
					t.Errorf("expected error message, got empty")
				}
			} else {
				if respMap, ok := response.(map[string]string); ok {
					if id, exists := respMap["id"]; !exists || id == "" {
						t.Errorf("expected ID in response, got %v", response)
					} else {
						// UUID 형식 검증
						if _, err := uuid.Parse(id); err != nil {
							t.Errorf("expected valid UUID, got %s", id)
						}
					}
				}
			}

			_ = w // ResponseRecorder 사용하지 않음 (직접 핸들러 호출)
		})
	}
}

// TestDashboardGetHandler GET 핸들러 테스트
func TestDashboardGetHandler(t *testing.T) {
	cfg, testEnforcer := setupDashboardTest(t)
	config = cfg
	enforcer = testEnforcer

	// 테스트 데이터 삽입
	testDashboardID := uuid.New().String()
	testConfig := `{"title": "Test Dashboard", "panels": []}`
	testSubject := "test-user"

	// 데이터베이스에 테스트 대시보드 삽입
	err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`INSERT INTO dashboard(id, config) VALUES (?, ?);`, testDashboardID, testConfig)
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
		dashboardID    string
		jwtClaims      *jwt.JWTClaims
		expectedStatus int
	}{
		{
			name:           "Get specific dashboard with permission",
			dashboardID:    testDashboardID,
			jwtClaims:      &jwt.JWTClaims{Subject: testSubject},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Get specific dashboard without permission",
			dashboardID:    testDashboardID,
			jwtClaims:      &jwt.JWTClaims{Subject: "other-user"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Get all dashboards",
			dashboardID:    "",
			jwtClaims:      &jwt.JWTClaims{Subject: testSubject},
			expectedStatus: http.StatusOK,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc = NewDashboardService(config, enforcer)

			path := "/dashboard"
			if tt.dashboardID != "" {
				path += "/" + tt.dashboardID
			}

			c, _ := createTestContext("GET", path, nil, tt.jwtClaims)

			// URL 파라미터 설정
			if tt.dashboardID != "" {
				c.Params = []gin.Param{{Key: "id", Value: tt.dashboardID}}
			}

			statusCode, response := svc.GetHandler(c)

			if statusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, statusCode)
			}

			if statusCode == http.StatusOK {
				if tt.dashboardID != "" {
					// 단일 대시보드 응답 검증
					if respMap, ok := response.(map[string]any); ok {
						if _, exists := respMap["permission"]; !exists {
							t.Errorf("expected permission in response")
						}
						if _, exists := respMap["config"]; !exists {
							t.Errorf("expected config in response")
						}
					}
				} else {
					// 다중 대시보드 응답 검증
					if respMap, ok := response.(map[string]any); ok {
						if len(respMap) == 0 {
							t.Log("no dashboards returned (expected for user with limited permissions)")
						}
					}
				}
			}
		})
	}
}

// TestDashboardPutHandler PUT 핸들러 테스트
func TestDashboardPutHandler(t *testing.T) {
	cfg, testEnforcer := setupDashboardTest(t)
	config = cfg
	enforcer = testEnforcer

	// 테스트 데이터 삽입
	testDashboardID := uuid.New().String()
	testConfig := `{"title": "Original Dashboard"}`
	testSubject := "test-user"

	// 데이터베이스에 테스트 대시보드 삽입
	err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`INSERT INTO dashboard(id, config) VALUES (?, ?);`, testDashboardID, testConfig)
		return err
	})
	if err != nil {
		t.Fatalf("failed to insert test dashboard: %v", err)
	}

	// Owner 권한 추가
	policy := casbin.Policy{testSubject, testDashboardID, casbin.ActionOwner, PermissionKindUser}
	err = enforcer.AddPolicy(policy)
	if err != nil {
		t.Fatalf("failed to add policy: %v", err)
	}

	tests := []struct {
		name           string
		dashboardID    string
		jwtClaims      *jwt.JWTClaims
		requestBody    string
		expectedStatus int
	}{
		{
			name:           "Update with owner permission",
			dashboardID:    testDashboardID,
			jwtClaims:      &jwt.JWTClaims{Subject: testSubject},
			requestBody:    `{"title": "Updated Dashboard"}`,
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Update without permission",
			dashboardID:    testDashboardID,
			jwtClaims:      &jwt.JWTClaims{Subject: "other-user"},
			requestBody:    `{"title": "Unauthorized Update"}`,
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Update with empty ID",
			dashboardID:    "",
			jwtClaims:      &jwt.JWTClaims{Subject: testSubject},
			requestBody:    `{"title": "No ID"}`,
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc = NewDashboardService(config, enforcer)

			path := "/dashboard/" + tt.dashboardID
			c, _ := createTestContext("PUT", path, []byte(tt.requestBody), tt.jwtClaims)
			c.Params = []gin.Param{{Key: "id", Value: tt.dashboardID}}

			statusCode, response := svc.PutHandler(c)

			if statusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, statusCode)
			}

			if statusCode == http.StatusBadRequest {
				if errorResp, ok := response.(external.ErrorResponse); !ok {
					t.Errorf("expected error response for bad request")
				} else if !strings.Contains(errorResp.Message, "empty") {
					t.Errorf("expected 'empty' in error message, got: %s", errorResp.Message)
				}
			}
		})
	}
}

// TestDashboardDeleteHandler DELETE 핸들러 테스트
func TestDashboardDeleteHandler(t *testing.T) {
	cfg, testEnforcer := setupDashboardTest(t)
	config = cfg
	enforcer = testEnforcer

	// 테스트 데이터 삽입
	testDashboardID := uuid.New().String()
	testConfig := `{"title": "Dashboard to Delete"}`
	testSubject := "test-user"

	// 데이터베이스에 테스트 대시보드 삽입
	err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`INSERT INTO dashboard(id, config) VALUES (?, ?);`, testDashboardID, testConfig)
		return err
	})
	if err != nil {
		t.Fatalf("failed to insert test dashboard: %v", err)
	}

	// Owner 권한 추가
	policy := casbin.Policy{testSubject, testDashboardID, casbin.ActionOwner, PermissionKindUser}
	err = enforcer.AddPolicy(policy)
	if err != nil {
		t.Fatalf("failed to add policy: %v", err)
	}

	tests := []struct {
		name           string
		dashboardID    string
		jwtClaims      *jwt.JWTClaims
		expectedStatus int
	}{
		{
			name:           "Delete with owner permission",
			dashboardID:    testDashboardID,
			jwtClaims:      &jwt.JWTClaims{Subject: testSubject},
			expectedStatus: http.StatusOK,
		},
		{
			name:           "Delete without permission",
			dashboardID:    "non-existent-id",
			jwtClaims:      &jwt.JWTClaims{Subject: "other-user"},
			expectedStatus: http.StatusForbidden,
		},
		{
			name:           "Delete with empty ID",
			dashboardID:    "",
			jwtClaims:      &jwt.JWTClaims{Subject: testSubject},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			svc = NewDashboardService(config, enforcer)

			path := "/dashboard/" + tt.dashboardID
			c, _ := createTestContext("DELETE", path, nil, tt.jwtClaims)
			c.Params = []gin.Param{{Key: "id", Value: tt.dashboardID}}

			statusCode, response := svc.DeleteHandler(c)

			if statusCode != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, statusCode)
			}

			if statusCode == http.StatusBadRequest {
				if errorResp, ok := response.(external.ErrorResponse); !ok {
					t.Errorf("expected error response for bad request")
				} else if !strings.Contains(errorResp.Message, "empty") {
					t.Errorf("expected 'empty' in error message, got: %s", errorResp.Message)
				}
			}
		})
	}
}

// TestDashboardHideConfig hideConfig 메서드 테스트
func TestDashboardHideConfig(t *testing.T) {
	cfg, testEnforcer := setupDashboardTest(t)
	config = cfg
	enforcer = testEnforcer
	svc = NewDashboardService(config, enforcer)

	datasourceName := "test-datasource"
	query := "SELECT * FROM test"
	filterDatasource := "header-datasource"
	filterQuery := "SELECT header"

	originalConfig := &dashboard.DashboardConfig{
		Title: "Test Dashboard",
		Panels: []dashboard.Panel{
			{
				Title: "Panel 1",
				DataProvider: &dashboard.DataProvider{
					ChartQuery: []dashboard.ChartQuery{
						{
							DatasourceName: datasourceName,
							Query:          query,
							Label:          "Test Query",
						},
					},
				},
			},
		},
		Filters: []dashboard.FilterConfig{
			{
				ID:             "filter-1",
				Type:           "select",
				Kind:           dashboard.HeaderL,
				DatasourceName: &filterDatasource,
				Query:          &filterQuery,
			},
		},
	}

	tests := []struct {
		name                   string
		shouldHide             bool
		expectDatasourceHidden bool
		expectQueryHidden      bool
	}{
		{
			name:                   "Hide sensitive information",
			shouldHide:             true,
			expectDatasourceHidden: true,
			expectQueryHidden:      true,
		},
		{
			name:                   "Keep sensitive information",
			shouldHide:             false,
			expectDatasourceHidden: false,
			expectQueryHidden:      false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// 원본 설정을 복사
			configBytes, _ := json.Marshal(originalConfig)
			var configCopy dashboard.DashboardConfig
			_ = json.Unmarshal(configBytes, &configCopy)

			if tt.shouldHide {
				svc.hideConfig(&configCopy)
			}

			// Panels의 ChartQuery 검증
			if len(configCopy.Panels) > 0 && configCopy.Panels[0].DataProvider != nil {
				for _, cq := range configCopy.Panels[0].DataProvider.ChartQuery {
					if tt.expectDatasourceHidden {
						if cq.DatasourceName != "" {
							t.Errorf("datasourceName should be hidden, but got: %s", cq.DatasourceName)
						}
						if cq.Query != "" {
							t.Errorf("query should be hidden, but got: %s", cq.Query)
						}
					} else {
						if cq.DatasourceName == "" {
							t.Errorf("datasourceName should not be hidden")
						}
						if cq.Query == "" {
							t.Errorf("query should not be hidden")
						}
					}
				}
			}

			// Filters 검증
			for _, filter := range configCopy.Filters {
				if tt.expectDatasourceHidden {
					if filter.DatasourceName != nil {
						t.Errorf("filter datasourceName should be hidden, but got: %v", *filter.DatasourceName)
					}
					if filter.Query != nil {
						t.Errorf("filter query should be hidden, but got: %v", *filter.Query)
					}
				} else {
					if filter.DatasourceName == nil {
						t.Errorf("filter datasourceName should not be hidden")
					}
					if filter.Query == nil {
						t.Errorf("filter query should not be hidden")
					}
				}
			}
		})
	}
}

// TestDashboardGets gets 메서드 테스트
func TestDashboardGets(t *testing.T) {
	cfg, testEnforcer := setupDashboardTest(t)
	config = cfg
	enforcer = testEnforcer

	// 테스트 데이터 삽입
	testDashboards := []struct {
		id     string
		config string
	}{
		{uuid.New().String(), `{"title": "Dashboard 1"}`},
		{uuid.New().String(), `{"title": "Dashboard 2"}`},
		{uuid.New().String(), `{"title": "Dashboard 3"}`},
	}

	for _, td := range testDashboards {
		err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
			_, err := db.Exec(`INSERT INTO dashboard(id, config) VALUES (?, ?);`, td.id, td.config)
			return err
		})
		if err != nil {
			t.Fatalf("failed to insert test dashboard: %v", err)
		}
	}

	svc = NewDashboardService(config, enforcer)
	dashboards, err := svc.gets()

	if err != nil {
		t.Fatalf("gets() failed: %v", err)
	}

	if len(dashboards) != len(testDashboards) {
		t.Errorf("expected %d dashboards, got %d", len(testDashboards), len(dashboards))
	}

	// 반환된 대시보드 ID들이 예상된 것들인지 확인
	foundIDs := make(map[string]bool)
	for _, d := range dashboards {
		foundIDs[d.ID] = true
	}

	for _, td := range testDashboards {
		if !foundIDs[td.id] {
			t.Errorf("expected dashboard ID %s not found in results", td.id)
		}
	}
}
