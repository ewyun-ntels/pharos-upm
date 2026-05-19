package metrics

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

func setupTestApi() *Api {
	api := &Api{}
	config := common.Config{
		Serve: common.ServeConfig{},
		Metrics: common.MetricsConfig{
			Use: true,
		},
		Statistics: common.StatisticsConfig{
			Database: orm.DatabaseConfig{
				Driver: orm.DriverSqlite,
			},
		},
	}
	api.Init("/test/config.yaml", config)
	return api
}

func TestApiInit(t *testing.T) {
	api := &Api{}
	configPath := "/test/config.yaml"
	config := common.Config{
		Serve: common.ServeConfig{},
		Metrics: common.MetricsConfig{
			Use: true,
		},
	}

	api.Init(configPath, config)

	assert.Equal(t, configPath, api.configPath)
	assert.Equal(t, config, api.config)
	assert.NotNil(t, api.registry)
}

func TestApiUse_MetricsEnabled(t *testing.T) {
	api := &Api{}
	config := common.Config{
		Serve: common.ServeConfig{},
		Metrics: common.MetricsConfig{
			Use: true,
		},
	}
	api.Init("/test/config.yaml", config)

	assert.True(t, api.Use())
}

func TestApiUse_MetricsDisabled(t *testing.T) {
	api := &Api{}
	config := common.Config{
		Serve: common.ServeConfig{},
		Metrics: common.MetricsConfig{
			Use: false,
		},
	}
	api.Init("/test/config.yaml", config)

	assert.False(t, api.Use())
}

func TestApiLoad(t *testing.T) {
	api := setupTestApi()

	// Add a mock collector
	mockCol := &mockCollector{}
	AddCollector("test", mockCol)

	err := api.Load()
	assert.NoError(t, err)
	assert.NotNil(t, api.registry)
}

func TestApiUnload(t *testing.T) {
	api := setupTestApi()

	// Add and load collectors
	mockCol := &mockCollector{}
	AddCollector("test", mockCol)
	err := api.Load()
	assert.NoError(t, err)

	// Unload should not panic
	api.Unload()
	assert.NotNil(t, api.registry)
}

func TestApiGetRelativePath(t *testing.T) {
	api := &Api{}
	assert.Equal(t, "/metrics", api.GetRelativePath())
}

func TestApiRegisterRoutes(t *testing.T) {
	api := setupTestApi()
	err := api.Load()
	assert.NoError(t, err)

	// Create test router
	gin.SetMode(gin.TestMode)
	router := gin.New()

	// Register routes
	api.RegisterRoutes(router)

	// Test the endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics", nil)
	router.ServeHTTP(w, req)

	// Should return 200 or 404 depending on whether the path was registered correctly
	// Since we're using gin.IRoutes directly, the path should be registered at root "/"
	w2 := httptest.NewRecorder()
	req2, _ := http.NewRequest("GET", "/", nil)
	router.ServeHTTP(w2, req2)

	assert.Equal(t, http.StatusOK, w2.Code)
}

func TestApiRegisterRoutes_FullPath(t *testing.T) {
	api := setupTestApi()
	err := api.Load()
	assert.NoError(t, err)

	// Create test router with group
	gin.SetMode(gin.TestMode)
	router := gin.New()
	group := router.Group("/api")

	// Register routes
	api.RegisterRoutes(group)

	// Test the endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/api/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Contains(t, w.Body.String(), "# HELP", "Should contain Prometheus metrics")
}

func TestApiIntegration(t *testing.T) {
	// Full integration test
	api := &Api{}
	config := common.Config{
		Serve: common.ServeConfig{},
		Metrics: common.MetricsConfig{
			Use: true,
		},
		Statistics: common.StatisticsConfig{
			Database: orm.DatabaseConfig{
				Driver: orm.DriverSqlite,
			},
		},
	}

	// Initialize
	api.Init("/test/config.yaml", config)
	assert.True(t, api.Use())

	// Add test collector
	testCollector := prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "test_counter",
			Help: "A test counter",
		},
		[]string{"label"},
	)
	AddCollector("test_integration", testCollector)

	// Load
	err := api.Load()
	assert.NoError(t, err)

	// Create router and register routes
	gin.SetMode(gin.TestMode)
	router := gin.New()
	metricsGroup := router.Group(api.GetRelativePath())
	api.RegisterRoutes(metricsGroup)

	// Increment test counter
	testCollector.WithLabelValues("test").Inc()

	// Test HTTP endpoint
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/metrics/", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	body := w.Body.String()
	assert.Contains(t, body, "# HELP")
	assert.Contains(t, body, "# TYPE")

	// Unload
	api.Unload()
}

func TestApiLoad_CreatesNewRegistry(t *testing.T) {
	api := setupTestApi()

	// Load first time
	err := api.Load()
	assert.NoError(t, err)
	registry1 := api.registry

	// Load second time
	err = api.Load()
	assert.NoError(t, err)
	registry2 := api.registry

	// Should create new registry each time
	assert.NotEqual(t, registry1, registry2)
}

func BenchmarkApiLoad(b *testing.B) {
	api := setupTestApi()
	AddCollector("bench", &mockCollector{})

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = api.Load()
	}
}

func BenchmarkApiRegisterRoutes(b *testing.B) {
	api := setupTestApi()
	_ = api.Load()

	gin.SetMode(gin.TestMode)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		router := gin.New()
		api.RegisterRoutes(router)
	}
}
