package command

import (
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
)

// MockExtension은 테스트용 Extension 구현체입니다.
// 실제 Extension을 시뮬레이션하여 로드/언로드 동작을 검증합니다.
type MockExtension struct {
	name         string
	loadCalled   bool
	unloadCalled bool
	loadError    error
}

func (m *MockExtension) Name() string {
	return m.name
}

func (m *MockExtension) Load(config common.Config) error {
	m.loadCalled = true
	return m.loadError
}

func (m *MockExtension) Unload() {
	m.unloadCalled = true
}

// TestExtension_Interface는 MockExtension이 Extension 인터페이스를 올바르게 구현하는지 검증합니다.
func TestExtension_Interface(t *testing.T) {
	// Extension 인터페이스가 올바르게 구현되었는지 확인
	var _ Extension = (*MockExtension)(nil)
}

func TestRegister_NewExtension(t *testing.T) {
	// 테스트 전 레지스트리 초기화
	globalRegistry = &ExtensionRegistry{
		extensions: make(map[string]Extension),
	}

	mock := &MockExtension{name: "test-extension"}
	Register(mock)

	// 확장이 등록되었는지 확인
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	ext, exists := globalRegistry.extensions["test-extension"]
	assert.True(t, exists)
	assert.Equal(t, mock, ext)
}

func TestRegister_DuplicateExtension(t *testing.T) {
	// 테스트 전 레지스트리 초기화
	globalRegistry = &ExtensionRegistry{
		extensions: make(map[string]Extension),
	}

	mock1 := &MockExtension{name: "duplicate-ext"}
	mock2 := &MockExtension{name: "duplicate-ext"}

	Register(mock1)
	Register(mock2) // 같은 이름으로 다시 등록

	// 두 번째 확장으로 덮어씌워졌는지 확인
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()

	ext, exists := globalRegistry.extensions["duplicate-ext"]
	assert.True(t, exists)
	assert.Equal(t, mock2, ext)
}

func TestLoadExtensions_Success(t *testing.T) {
	// 테스트 전 레지스트리 초기화
	globalRegistry = &ExtensionRegistry{
		extensions: make(map[string]Extension),
	}

	mock1 := &MockExtension{name: "ext1"}
	mock2 := &MockExtension{name: "ext2"}

	Register(mock1)
	Register(mock2)

	config := common.Config{}
	err := loadExtensions(config)

	assert.NoError(t, err)
	assert.True(t, mock1.loadCalled)
	assert.True(t, mock2.loadCalled)
}

func TestLoadExtensions_WithError(t *testing.T) {
	// 테스트 전 레지스트리 초기화
	globalRegistry = &ExtensionRegistry{
		extensions: make(map[string]Extension),
	}

	mockError := &MockExtension{
		name:      "error-ext",
		loadError: assert.AnError,
	}

	Register(mockError)

	config := common.Config{}
	err := loadExtensions(config)

	assert.Error(t, err)
	assert.True(t, mockError.loadCalled)
}

func TestUnloadExtensions(t *testing.T) {
	// 테스트 전 레지스트리 초기화
	globalRegistry = &ExtensionRegistry{
		extensions: make(map[string]Extension),
	}

	mock1 := &MockExtension{name: "ext1"}
	mock2 := &MockExtension{name: "ext2"}

	Register(mock1)
	Register(mock2)

	unloadExtensions()

	assert.True(t, mock1.unloadCalled)
	assert.True(t, mock2.unloadCalled)
}

func TestExtensionRegistry_ConcurrentAccess(t *testing.T) {
	// 테스트 전 레지스트리 초기화
	globalRegistry = &ExtensionRegistry{
		extensions: make(map[string]Extension),
	}

	var wg sync.WaitGroup
	const numGoroutines = 10

	// 동시에 여러 확장 등록
	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()
			mock := &MockExtension{name: "concurrent-ext"}
			Register(mock)
		}(i)
	}

	wg.Wait()

	// 레이스 컨디션 없이 완료되었는지 확인
	globalRegistry.mu.RLock()
	defer globalRegistry.mu.RUnlock()
	assert.Len(t, globalRegistry.extensions, 1)
}

func TestExtensionRegistry_EmptyRegistry(t *testing.T) {
	// 테스트 전 레지스트리 초기화
	globalRegistry = &ExtensionRegistry{
		extensions: make(map[string]Extension),
	}

	config := common.Config{}

	// 빈 레지스트리에서 로드/언로드 호출
	err := loadExtensions(config)
	assert.NoError(t, err)

	unloadExtensions() // 에러 없이 완료되어야 함
}

func TestExtensionRegistry_MultipleLoadUnload(t *testing.T) {
	// 테스트 전 레지스트리 초기화
	globalRegistry = &ExtensionRegistry{
		extensions: make(map[string]Extension),
	}

	mock := &MockExtension{name: "multi-ext"}
	Register(mock)

	config := common.Config{}

	// 여러 번 로드/언로드 반복
	for range 3 {
		mock.loadCalled = false
		mock.unloadCalled = false

		err := loadExtensions(config)
		require.NoError(t, err)
		assert.True(t, mock.loadCalled)

		unloadExtensions()
		assert.True(t, mock.unloadCalled)
	}
}

// Benchmark tests
func BenchmarkRegister(b *testing.B) {
	globalRegistry = &ExtensionRegistry{
		extensions: make(map[string]Extension),
	}

	for b.Loop() {
		mock := &MockExtension{name: "bench-ext"}
		Register(mock)
	}
}

func BenchmarkLoadExtensions(b *testing.B) {
	globalRegistry = &ExtensionRegistry{
		extensions: make(map[string]Extension),
	}

	mock := &MockExtension{name: "bench-ext"}
	Register(mock)

	config := common.Config{}

	for b.Loop() {
		_ = loadExtensions(config)
	}
}

func BenchmarkUnloadExtensions(b *testing.B) {
	globalRegistry = &ExtensionRegistry{
		extensions: make(map[string]Extension),
	}

	mock := &MockExtension{name: "bench-ext"}
	Register(mock)

	for b.Loop() {
		unloadExtensions()
	}
}
