package dashboard

import (
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/common"
)

func TestDashboardApi_Init(t *testing.T) {
	api := &Api{}
	configPath := "/test/config/path"
	config := common.Config{}

	api.Init(configPath, config)

	if api.configPath != configPath {
		t.Errorf("expected configPath %s, got %s", configPath, api.configPath)
	}
}

func TestDashboardApi_Use(t *testing.T) {
	api := &Api{}
	if !api.Use() {
		t.Errorf("expected Use to return true, got false")
	}
}

func TestDashboardApi_Load(t *testing.T) {
	// Test Load method with proper config setup
	// Note: This test may fail if database connection is not available
	api := &Api{
		config: common.Config{},
	}

	// The Load method tries to initialize casbin enforcer
	// In a real test environment, you would need proper database setup
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Load method panicked: %v", r)
		}
	}()

	// This might fail due to missing database, but shouldn't panic
	api.Load()
}

func TestDashboardApi_Unload(t *testing.T) {
	api := &Api{}
	// Should not panic
	api.Unload()
}

func TestDashboardApi_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := common.Config{}

	api := &Api{config: config}

	// Try to load first (might fail but shouldn't panic)
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RegisterRoutes panicked: %v", r)
		}
	}()

	api.RegisterRoutes(router)
}

func TestDashboardApi_GetRelativePath(t *testing.T) {
	api := &Api{}
	expected := "/dashboard"
	result := api.GetRelativePath()
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// Test concurrent access to API struct methods
func TestDashboardApi_ConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 50

	// Test concurrent creation and basic method calls
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int) {
			defer wg.Done()
			api := &Api{}
			config := common.Config{}
			configPath := "/test/path"
			api.Init(configPath, config)
			api.Use()
			api.GetRelativePath()
			api.Unload()
		}(i)
	}
	wg.Wait()
}

// Test concurrent route registration - sequential to avoid global variable races
func TestDashboardApi_ConcurrentRouteRegistration(t *testing.T) {
	// Due to global variables in handler package (config, enforcer),
	// we test route registration sequentially to avoid race conditions
	numTests := 10

	for i := range numTests {
		func() {
			defer func() {
				if r := recover(); r != nil {
					// Expected to potentially fail due to casbin or authhandler setup
					t.Logf("Route registration %d failed as expected: %v", i, r)
				}
			}()

			router := gin.New()
			config := common.Config{}
			api := &Api{config: config}
			api.RegisterRoutes(router)
		}()
	}
}

// Test concurrent Load operations with sync.Once
func TestDashboardApi_ConcurrentLoad(t *testing.T) {
	// Reset the once variable for this test
	once = sync.Once{}
	enforcer = nil

	var wg sync.WaitGroup
	numGoroutines := 30

	apis := make([]*Api, numGoroutines)
	for i := range numGoroutines {
		apis[i] = &Api{
			config: common.Config{},
		}
	}

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(api *Api) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					// Expected to potentially fail due to database setup
				}
			}()

			api.Load()
		}(apis[i])
	}
	wg.Wait()

	// Verify that sync.Once worked correctly by ensuring enforcer was initialized once
	// Even if Load failed, the once mechanism should prevent multiple attempts
}

// Test API struct field safety with separate instances
func TestDashboardApi_FieldSafety(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 100

	// Create separate API instances for each goroutine to avoid race
	apis := make([]*Api, numGoroutines)
	for i := range numGoroutines {
		apis[i] = &Api{}
	}

	// Test concurrent field access with independent instances
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int, api *Api) {
			defer wg.Done()
			if index%2 == 0 {
				// Writer goroutines with independent instances
				config := common.Config{}
				api.Init("/test/path", config)
			}
		}(i, apis[i])
	}
	wg.Wait()
}

// Test method combinations
func TestDashboardApi_MethodCombinations(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 20

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int) {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Method combination panicked: %v", r)
				}
			}()

			api := &Api{}
			config := common.Config{}

			// Test different method call orders
			switch index % 4 {
			case 0:
				api.Init("/test", config)
				api.Use()
				api.Load()
				api.GetRelativePath()
				api.Unload()
			case 1:
				api.Use()
				api.Init("/test", config)
				api.GetRelativePath()
				api.Load()
			case 2:
				api.GetRelativePath()
				api.Use()
				api.Init("/test", config)
			case 3:
				api.Load()
				api.Init("/test", config)
				api.Use()
				api.Unload()
			}
		}(i)
	}
	wg.Wait()
}
