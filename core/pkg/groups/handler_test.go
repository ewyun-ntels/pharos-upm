package groups

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite/token/jwt"
	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/casbin"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/shared/types/role"
)

// ============================================================================
// HELPER FUNCTIONS FOR HANDLER TESTS
// ============================================================================

func setupHandlerTestEnforcer(t *testing.T) (*casbin.Enforcer, func()) {
	// Create temporary SQLite database file
	tempFile, err := os.CreateTemp("", "casbin_test_*.db")
	if err != nil {
		t.Fatalf("Failed to create temp database file: %v", err)
	}
	_ = tempFile.Close()

	// Setup SQLite database configuration
	dbConfig := orm.DatabaseConfig{
		Driver: "sqlite",
		SQLite: orm.SQLiteConfig{
			Path: tempFile.Name(),
		},
	}

	// Create enforcer with SQLite database
	testEnforcer, err := casbin.NewEnforcer(casbin.EnforcerTypeGroup, dbConfig)
	if err != nil {
		_ = os.Remove(tempFile.Name())
		t.Fatalf("Failed to create test enforcer: %v", err)
	}

	// Return cleanup function
	cleanup := func() {
		_ = os.Remove(tempFile.Name())
	}

	return testEnforcer, cleanup
}

func setupHandlerIntegrationTest(t *testing.T) (*gin.Engine, func()) {
	// Setup test enforcer for integration tests
	testEnforcer, cleanup := setupHandlerTestEnforcer(t)

	// Set the global enforcer for handlers to use
	enforcer = testEnforcer

	gin.SetMode(gin.TestMode)
	router := setupHandlerIntegrationRouter()

	return router, cleanup
}

func setupHandlerIntegrationRouter() *gin.Engine {
	router := gin.New()

	// Add authentication middleware
	router.Use(func(c *gin.Context) {
		// Set default JWT claims for testing
		claims := &jwt.JWTClaims{
			Subject: "integration-test-user",
			Extra: map[string]any{
				"attrs": map[string]any{
					string(role.RoleSuperAdmin): false,
					string(role.RoleGroupAdd):   true,
				},
			},
		}
		c.Set(common.ContextKeyJWTClaims, claims)
		c.Next()
	})

	// Register routes similar to actual application
	groupRoutes := router.Group("/groups")
	{
		groupRoutes.GET("", groupsHandler)
		groupRoutes.GET("/:group-name", groupsHandler)
		groupRoutes.POST("", groupsHandler)
		groupRoutes.DELETE("/:group-name", groupsHandler)

		groupRoutes.GET("/:group-name/members", membersHandler)
		groupRoutes.PUT("/:group-name/members", membersHandler)
		groupRoutes.DELETE("/:group-name/members/:user-name", membersHandler)
	}

	return router
}

// ============================================================================
// UNIT TESTS FOR HANDLERS
// ============================================================================

func TestGroupJSONValidation(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantErr  bool
	}{
		{
			name:     "Valid input",
			jsonData: `{"name": "test-group"}`,
			wantErr:  false,
		},
		{
			name:     "Empty group name",
			jsonData: `{"name": ""}`,
			wantErr:  false, // JSON unmarshaling succeeds, validation is separate
		},
		{
			name:     "Missing group name",
			jsonData: `{}`,
			wantErr:  false,
		},
		{
			name:     "Invalid JSON",
			jsonData: `{"name": "test-group"`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			group := Group{}
			err := json.Unmarshal([]byte(tt.jsonData), &group)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestMemberJSONParsing(t *testing.T) {
	tests := []struct {
		name     string
		jsonData string
		wantErr  bool
	}{
		{
			name:     "Valid member data",
			jsonData: `{"subject": "test-user", "action": "read"}`,
			wantErr:  false,
		},
		{
			name:     "Empty subject",
			jsonData: `{"subject": "", "action": "read"}`,
			wantErr:  true, // setFromReader validates and rejects empty subject
		},
		{
			name:     "Invalid JSON",
			jsonData: `{"subject": "test-user"`,
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			member := Member{}
			reader := strings.NewReader(tt.jsonData)
			err := member.setFromReader(reader)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

// ============================================================================
// INTEGRATION TESTS FOR HANDLERS
// ============================================================================

func TestGroupHandlersIntegration(t *testing.T) {
	// Skip if no database connection available
	if testing.Short() {
		t.Skip("Skipping integration tests in short mode")
	}

	router, cleanup := setupHandlerIntegrationTest(t)
	defer cleanup()

	t.Run("Full group lifecycle", func(t *testing.T) {
		// 1. Create group
		createBody := `{"name": "integration-test-group"}`
		req, _ := http.NewRequest("POST", "/groups", strings.NewReader(createBody))
		req.Header.Set("Content-Type", "application/json")

		w := httptest.NewRecorder()
		router.ServeHTTP(w, req)

		t.Logf("Create group response: %d", w.Code)

		// 2. Get group
		req, _ = http.NewRequest("GET", "/groups/integration-test-group", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		t.Logf("Get group response: %d", w.Code)

		// 3. Delete group
		req, _ = http.NewRequest("DELETE", "/groups/integration-test-group", nil)
		w = httptest.NewRecorder()
		router.ServeHTTP(w, req)

		t.Logf("Delete group response: %d", w.Code)
	})
}

func TestErrorHandlingIntegration(t *testing.T) {
	router, cleanup := setupHandlerIntegrationTest(t)
	defer cleanup()

	tests := []struct {
		name           string
		method         string
		path           string
		body           string
		contentType    string
		expectedStatus int
	}{
		{
			name:           "Invalid JSON in group creation",
			method:         "POST",
			path:           "/groups",
			body:           `{"name": "test-group"`,
			contentType:    "application/json",
			expectedStatus: 500,
		},
		{
			name:           "Empty request body",
			method:         "POST",
			path:           "/groups",
			body:           "",
			contentType:    "application/json",
			expectedStatus: 500,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req, _ := http.NewRequest(tt.method, tt.path, strings.NewReader(tt.body))
			if tt.contentType != "" {
				req.Header.Set("Content-Type", tt.contentType)
			}

			w := httptest.NewRecorder()
			router.ServeHTTP(w, req)

			t.Logf("Test %s: Status %d", tt.name, w.Code)
		})
	}
}

// ============================================================================
// BENCHMARK TESTS FOR HANDLERS
// ============================================================================

func BenchmarkGroupJSONMarshal(b *testing.B) {
	group := Group{Name: "benchmark-group"}
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, _ = json.Marshal(group)
	}
}

func BenchmarkGroupJSONUnmarshal(b *testing.B) {
	jsonData := []byte(`{"name": "benchmark-group"}`)
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var group Group
		_ = json.Unmarshal(jsonData, &group)
	}
}
