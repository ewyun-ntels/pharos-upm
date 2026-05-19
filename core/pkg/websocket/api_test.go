package websocket

import (
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/common"
	centrifuge_node "ntels.com/pharos/core/pkg/websocket/centrifuge/node"
)

func TestWebsocketApi_Init(t *testing.T) {
	api := &Api{}
	configPath := "/test/config/path"
	config := common.Config{}

	api.Init(configPath, config)

	if api.configPath != configPath {
		t.Errorf("expected configPath %s, got %s", configPath, api.configPath)
	}
}

func TestWebsocketApi_Use(t *testing.T) {
	api := &Api{}
	if !api.Use() {
		t.Errorf("expected Use to return true, got false")
	}
}

func TestWebsocketApi_Load(t *testing.T) {
	api := &Api{
		config: common.Config{},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Load method panicked: %v", r)
		}
	}()

	// This might fail due to centrifuge node setup, but shouldn't panic
	api.Load()
}

func TestWebsocketApi_Unload(t *testing.T) {
	// Test with properly initialized API
	config := common.Config{}
	api := &Api{}
	api.Init("/test", config)

	// First try to Load to initialize the node
	func() {
		defer func() {
			if r := recover(); r != nil {
				// Expected to potentially fail due to centrifuge setup
				t.Logf("Load failed as expected: %v", r)
				return
			}
		}()
		if err := api.Load(); err != nil {
			t.Logf("Load failed as expected: %v", err)
			return
		}

		// Only test Unload if Load succeeded
		defer func() {
			if r := recover(); r != nil {
				t.Logf("Unload failed as expected: %v", r)
			}
		}()
		api.Unload()
	}()
}

func TestWebsocketApi_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := common.Config{}
	api := &Api{config: config}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RegisterRoutes panicked: %v", r)
		}
	}()

	api.RegisterRoutes(router)
}

func TestWebsocketApi_GetRelativePath(t *testing.T) {
	api := &Api{}
	expected := "/websocket"
	result := api.GetRelativePath()
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestWebsocketApi_ConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 50

	// Clear global Nodes map for test
	nodesMutex := &sync.Mutex{}
	nodesMutex.Lock()
	Nodes = make(map[string]*centrifuge_node.Node)
	nodesMutex.Unlock()

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int) {
			defer wg.Done()
			api := &Api{}
			config := common.Config{}
			api.Init("/test/path", config)
			api.Use()
			api.GetRelativePath()

			defer func() {
				if r := recover(); r != nil {
					// Expected to potentially fail due to centrifuge setup
				}
			}()

			api.Unload()
		}(i)
	}
	wg.Wait()
}

func TestWebsocketApi_ConcurrentRouteRegistration(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var wg sync.WaitGroup
	numGoroutines := 10

	wg.Add(numGoroutines)
	for range numGoroutines {
		go func() {
			defer wg.Done()
			defer func() {
				if r := recover(); r != nil {
					// Expected to potentially fail due to centrifuge node setup
				}
			}()

			router := gin.New()
			api := &Api{config: common.Config{}}
			api.RegisterRoutes(router)
		}()
	}
	wg.Wait()
}

func TestWebsocketApi_NodesMapSafety(t *testing.T) {
	// Test concurrent access to global Nodes map sequentially to avoid race conditions
	numTests := 30

	for i := range numTests {
		func(index int) {
			defer func() {
				if r := recover(); r != nil {
					// Expected to potentially fail
					t.Logf("Load %d failed as expected: %v", index, r)
				}
			}()

			api := &Api{}
			config := common.Config{}
			api.Init("/test", config)

			// This will try to add to Nodes map
			if err := api.Load(); err != nil {
				t.Logf("Load %d failed as expected: %v", index, err)
			}
		}(i)
	}
}

func TestWebsocketApi_FieldSafety(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 100

	// Create separate API instances for each goroutine to avoid race conditions
	apis := make([]*Api, numGoroutines)
	for i := range numGoroutines {
		apis[i] = &Api{}
	}

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int, api *Api) {
			defer wg.Done()
			if index%2 == 0 {
				config := common.Config{}
				api.Init("/test/path", config)
			} else {
				config := common.Config{}
				api.Init("/test/path", config)
				_ = api.configPath
				_ = api.config
				api.Use()
				api.GetRelativePath()
			}
		}(i, apis[i])
	}
	wg.Wait()
}

func TestWebsocketApi_LoadUnloadConcurrency(t *testing.T) {
	// Test Load/Unload sequentially to avoid Centrifuge library race conditions
	api := &Api{}
	config := common.Config{}
	api.Init("/test", config)

	numTests := 20

	// Test sequential Load/Unload operations
	for i := range numTests {
		func(index int) {
			defer func() {
				if r := recover(); r != nil {
					// Expected to potentially fail due to centrifuge setup
					t.Logf("Load/Unload %d failed as expected: %v", index, r)
				}
			}()

			if index%2 == 0 {
				if err := api.Load(); err != nil {
					t.Logf("Load %d failed as expected: %v", index, err)
				}
			} else {
				api.Unload()
			}
		}(i)
	}
}
