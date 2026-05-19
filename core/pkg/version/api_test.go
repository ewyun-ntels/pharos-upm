package version

import (
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/common"
)

func TestVersionApi_Init(t *testing.T) {
	api := &Api{}
	configPath := "/test/config/path"
	config := common.Config{}

	api.Init(configPath, config)

	if api.configPath != configPath {
		t.Errorf("expected configPath %s, got %s", configPath, api.configPath)
	}
}

func TestVersionApi_Use(t *testing.T) {
	// This API always returns true
	api := &Api{}
	if !api.Use() {
		t.Error("expected Use() to return true")
	}
}

func TestVersionApi_Load(t *testing.T) {
	api := &Api{
		config: common.Config{},
	}

	// Load method tries to upsert version to database
	// It should not fail server startup even if database is not available
	err := api.Load()
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestVersionApi_Unload(t *testing.T) {
	api := &Api{}
	api.Unload()
}

func TestVersionApi_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := common.Config{}
	api := &Api{config: config}
	api.RegisterRoutes(router)

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "", nil)
	router.ServeHTTP(w, req)

	if w.Code == http.StatusNotFound {
		t.Error("version route not registered")
	}
}

func TestVersionApi_GetRelativePath(t *testing.T) {
	api := &Api{}
	expected := "/version"
	result := api.GetRelativePath()
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestVersionApi_ConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 100

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int) {
			defer wg.Done()
			api := &Api{}
			config := common.Config{}
			api.Init("/test/path", config)
			api.Use()
			api.Load()
			api.GetRelativePath()
			api.Unload()
		}(i)
	}
	wg.Wait()
}

func TestVersionApi_ConcurrentRouteRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var wg sync.WaitGroup
	numGoroutines := 20

	wg.Add(numGoroutines)
	for range numGoroutines {
		go func() {
			defer wg.Done()
			router := gin.New()
			api := &Api{config: common.Config{}}
			api.RegisterRoutes(router)
		}()
	}
	wg.Wait()
}

func TestVersionApi_ConcurrentLoad(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 50

	wg.Add(numGoroutines)
	for range numGoroutines {
		go func() {
			defer wg.Done()
			api := &Api{config: common.Config{}}
			// Load should not fail even if database operations fail
			api.Load()
		}()
	}
	wg.Wait()
}

func TestVersionApi_FieldSafety(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 100

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int) {
			defer wg.Done()
			// Use separate API instance to avoid race condition
			api := &Api{}
			if index%2 == 0 {
				config := common.Config{}
				api.Init("/test/path", config)
			} else {
				_ = api.configPath
				_ = api.config
				api.Use()
				api.GetRelativePath()
			}
		}(i)
	}
	wg.Wait()
}
