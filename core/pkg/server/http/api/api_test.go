package api

import (
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

// MockApi는 테스트용 Api 구현체입니다
type MockApi struct {
	initialized bool
	loaded      bool
	use         bool
	path        string
	mu          sync.RWMutex
}

func (m *MockApi) Init(_ string, _ common.Config) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.initialized = true
}

func (m *MockApi) Use() bool {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.use
}

func (m *MockApi) Load() error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loaded = true
	return nil
}

func (m *MockApi) Unload() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.loaded = false
}

func (m *MockApi) RegisterRoutes(_ gin.IRoutes) {
	// Mock implementation
}

func (m *MockApi) GetRelativePath() string {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.path
}

func TestAddApi_ConcurrentAccess(t *testing.T) {
	// t.Parallel() 제거 - addApi가 전역 변수이므로 테스트 간 격리 필요

	// 테스트 전 addApi 슬라이스 초기화
	apiMutex.Lock()
	originalAddApi := addApi
	addApi = nil
	apiMutex.Unlock()

	defer func() {
		apiMutex.Lock()
		addApi = originalAddApi
		apiMutex.Unlock()
	}()

	const numGoroutines = 10
	var wg sync.WaitGroup

	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			mockApi := &MockApi{
				use:  true,
				path: "/test",
			}
			AddApi(mockApi)
		}(i)
	}

	wg.Wait()

	// 모든 API가 추가되었는지 확인
	apiMutex.RLock()
	length := len(addApi)
	apiMutex.RUnlock()
	assert.Equal(t, numGoroutines, length)
}

func TestGetApis_ConcurrentAccess(t *testing.T) {
	// t.Parallel() 제거 - addApi가 전역 변수이므로 테스트 간 격리 필요

	// 테스트 전 addApi 슬라이스 초기화
	apiMutex.Lock()
	originalAddApi := addApi
	addApi = nil
	apiMutex.Unlock()

	defer func() {
		apiMutex.Lock()
		addApi = originalAddApi
		apiMutex.Unlock()
	}()

	// 테스트용 MockApi 추가
	mockApi := &MockApi{
		use:  true,
		path: "/test",
	}
	AddApi(mockApi)

	config := common.Config{
		User: common.UsersConfig{},
		Database: orm.DatabaseConfig{
			Driver: orm.DriverSqlite,
			SQLite: orm.SQLiteConfig{
				Path: ":memory:",
			},
		},
	}
	const numGoroutines = 5
	var wg sync.WaitGroup
	results := make([][]Api, numGoroutines)

	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			results[id] = GetApis("/test/config", config)
		}(i)
	}

	wg.Wait()

	// 모든 결과가 동일한지 확인
	for i := 1; i < numGoroutines; i++ {
		assert.Equal(t, len(results[0]), len(results[i]),
			"All goroutines should return the same number of APIs")
	}
}

func TestApi_Interface_Implementation(t *testing.T) {
	t.Parallel()

	mockApi := &MockApi{
		use:  true,
		path: "/test",
	}

	// Api 인터페이스 구현 확인
	var _ Api = mockApi

	// 기본 동작 테스트
	config := common.Config{}
	mockApi.Init("/test/config", config)

	assert.True(t, mockApi.Use())
	assert.NoError(t, mockApi.Load())
	assert.Equal(t, "/test", mockApi.GetRelativePath())

	mockApi.Unload()
}

func TestApi_StateConsistency(t *testing.T) {
	t.Parallel()

	mockApi := &MockApi{
		use:  true,
		path: "/test",
	}

	const numGoroutines = 10
	var wg sync.WaitGroup

	// 동시에 여러 goroutine에서 상태를 변경
	wg.Add(numGoroutines * 3) // Init, Load, Unload 각각 numGoroutines 개

	for range numGoroutines {
		// Init 테스트
		go func() {
			defer wg.Done()
			config := common.Config{}
			mockApi.Init("/test/config", config)
		}()

		// Load 테스트
		go func() {
			defer wg.Done()
			_ = mockApi.Load()
		}()

		// Unload 테스트
		go func() {
			defer wg.Done()
			mockApi.Unload()
		}()
	}

	wg.Wait()

	// 최종 상태 확인 (race condition 없이 접근)
	assert.NotPanics(t, func() {
		_ = mockApi.Use()
		_ = mockApi.GetRelativePath()
	})
}
