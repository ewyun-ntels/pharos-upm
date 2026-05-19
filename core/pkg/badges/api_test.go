package badges

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/common"
)

func TestBadgesApi_Init(t *testing.T) {
	api := &Api{}
	configPath := "/test/config/path"
	config := common.Config{}

	api.Init(configPath, config)

	if api.configPath != configPath {
		t.Errorf("expected configPath %s, got %s", configPath, api.configPath)
	}
}

func TestBadgesApi_Use(t *testing.T) {
	api := &Api{}
	if !api.Use() {
		t.Errorf("expected Use to return true, got false")
	}
}

func TestBadgesApi_Load(t *testing.T) {
	tests := []struct {
		name     string
		config   common.Config
		expected bool
	}{
		{
			name: "load with nil cache TTL",
			config: common.Config{
				Badges: common.BadgesConfig{
					CacheTTL: nil,
				},
			},
			expected: true,
		},
		{
			name: "load with negative cache TTL",
			config: common.Config{
				Badges: common.BadgesConfig{
					CacheTTL: func() *time.Duration { d := -time.Minute; return &d }(),
				},
			},
			expected: true,
		},
		{
			name: "load with positive cache TTL",
			config: common.Config{
				Badges: common.BadgesConfig{
					CacheTTL: func() *time.Duration { d := 5 * time.Minute; return &d }(),
				},
			},
			expected: true,
		},
		{
			name: "load with zero cache TTL",
			config: common.Config{
				Badges: common.BadgesConfig{
					CacheTTL: func() *time.Duration { d := time.Duration(0); return &d }(),
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			api := &Api{config: tt.config}
			err := api.Load()
			if (err == nil) != tt.expected {
				t.Errorf("expected success %v, got error %v", tt.expected, err)
			}
		})
	}
}

func TestBadgesApi_Unload(t *testing.T) {
	api := &Api{}
	// Should not panic
	api.Unload()
}

func TestBadgesApi_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := common.Config{}

	api := &Api{config: config}
	api.RegisterRoutes(router)

	// Test that routes are registered
	routes := []struct {
		method string
		path   string
	}{
		{"GET", ""},
		{"POST", ""},
		{"GET", "/test-name"},
	}

	for _, route := range routes {
		w := httptest.NewRecorder()
		req, _ := http.NewRequest(route.method, route.path, nil)
		router.ServeHTTP(w, req)

		// Routes should be registered (not 404)
		// Note: They might return 401/403 due to auth requirements
		if w.Code == http.StatusNotFound {
			t.Errorf("route %s %s not registered", route.method, route.path)
		}
	}
}

func TestBadgesApi_GetRelativePath(t *testing.T) {
	api := &Api{}
	expected := "/badges"
	result := api.GetRelativePath()
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

// Test concurrent access to API struct methods
func TestBadgesApi_ConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 100

	// Test concurrent creation and initialization - race-free approach
	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int) {
			defer wg.Done()
			// Each goroutine works on its own API instance
			api := &Api{}
			config := common.Config{
				Badges: common.BadgesConfig{
					CacheTTL: func() *time.Duration { d := time.Minute; return &d }(),
				},
			}
			configPath := "/test/path"
			api.Init(configPath, config)
			api.Use()
			// Skip Load() to avoid global variable race
			api.GetRelativePath()
			api.Unload()
		}(i)
	}
	wg.Wait()
}

// Test concurrent route registration
func TestBadgesApi_ConcurrentRouteRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var wg sync.WaitGroup
	numGoroutines := 20

	wg.Add(numGoroutines)
	for range numGoroutines {
		go func() {
			defer wg.Done()
			// Each goroutine creates its own router and API instance
			router := gin.New()
			config := common.Config{
				Badges: common.BadgesConfig{
					CacheTTL: func() *time.Duration { d := time.Minute; return &d }(),
				},
			}
			api := &Api{config: config}
			// Skip Load to avoid global variable race
			api.RegisterRoutes(router)
		}()
	}
	wg.Wait()
}

// Test sequential Load operations to avoid race on global cache
func TestBadgesApi_SequentialLoad(t *testing.T) {
	// Test different cache TTL configurations sequentially
	configs := []common.Config{
		{Badges: common.BadgesConfig{CacheTTL: nil}},
		{Badges: common.BadgesConfig{CacheTTL: func() *time.Duration { d := -time.Minute; return &d }()}},
		{Badges: common.BadgesConfig{CacheTTL: func() *time.Duration { d := time.Duration(0); return &d }()}},
		{Badges: common.BadgesConfig{CacheTTL: func() *time.Duration { d := 5 * time.Minute; return &d }()}},
	}

	for i, config := range configs {
		t.Run(fmt.Sprintf("config_%d", i), func(t *testing.T) {
			api := &Api{}
			api.Init("/test", config)
			err := api.Load()
			if err != nil {
				t.Errorf("Load failed: %v", err)
			}
		})
	}
}

// Test cache initialization without race conditions
func TestBadgesApi_CacheInitialization(t *testing.T) {
	// Test cache initialization sequentially to avoid race
	api := &Api{}
	config := common.Config{
		Badges: common.BadgesConfig{
			CacheTTL: func() *time.Duration { d := time.Minute; return &d }(),
		},
	}
	api.Init("/test", config)

	err := api.Load()
	if err != nil {
		t.Errorf("Load failed: %v", err)
	}

	// Verify cache was initialized (if accessible)
	// Note: badgeCache is a global variable, so this test
	// verifies the functionality without concurrent access
}
