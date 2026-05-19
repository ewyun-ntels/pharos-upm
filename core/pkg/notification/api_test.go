package notification

import (
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/common"
)

func TestNotificationApi_Init(t *testing.T) {
	api := &Api{}
	configPath := "/test/config/path"
	config := common.Config{}

	api.Init(configPath, config)

	if api.configPath != configPath {
		t.Errorf("expected configPath %s, got %s", configPath, api.configPath)
	}
}

func TestNotificationApi_Use(t *testing.T) {
	api := &Api{}
	if !api.Use() {
		t.Errorf("expected Use to return true, got false")
	}
}

func TestNotificationApi_Load(t *testing.T) {
	api := &Api{
		config: common.Config{},
	}

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("Load method panicked: %v", r)
		}
	}()

	api.Load()
}

func TestNotificationApi_Unload(t *testing.T) {
	api := &Api{}
	api.Unload()
}

func TestNotificationApi_RegisterRoutes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	router := gin.New()

	config := common.Config{}

	api := &Api{config: config}
	api.RegisterRoutes(router)
}

func TestNotificationApi_GetRelativePath(t *testing.T) {
	api := &Api{}
	expected := "/notification"
	result := api.GetRelativePath()
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}
}

func TestNotificationApi_ConcurrentAccess(t *testing.T) {
	var wg sync.WaitGroup
	numGoroutines := 50

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int) {
			defer wg.Done()
			api := &Api{}
			config := common.Config{}
			api.Init("/test/path", config)
			api.Use()
			api.GetRelativePath()
			api.Unload()
		}(i)
	}
	wg.Wait()
}

func TestNotificationApi_ConcurrentLoadRules(t *testing.T) {
	const numGoroutines = 10
	var wg sync.WaitGroup

	for range numGoroutines {
		wg.Go(func() {

			// Use separate API instance to avoid race condition
			api := &Api{
				config: common.Config{},
			}
			api.loadRules()
		})
	}
	wg.Wait()
}

func TestNotificationApi_LoadRulesFlag(t *testing.T) {
	// Test the isLoading flag functionality
	originalIsLoading := isLoading
	defer func() {
		isLoading = originalIsLoading
	}()

	isLoading = true

	api := &Api{
		config: common.Config{},
	}

	// Should return nil immediately due to isLoading flag
	err := api.loadRules()
	if err != nil {
		t.Errorf("expected nil error when isLoading is true, got %v", err)
	}
}
