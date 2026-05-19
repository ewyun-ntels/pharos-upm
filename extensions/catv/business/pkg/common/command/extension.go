package command

import (
	"log/slog"
	"maps"
	"sync"

	"ntels.com/pharos/core/pkg/common"
)

// Extension은 확장 모듈이 구현해야 하는 인터페이스입니다.
type Extension interface {
	// Name은 확장 모듈의 이름을 반환합니다.
	Name() string

	// Load는 확장 모듈을 초기화합니다.
	Load(config common.Config) error

	// Unload는 확장 모듈을 정리합니다.
	Unload()
}

// ExtensionRegistry는 확장 모듈을 관리하는 레지스트리입니다.
type ExtensionRegistry struct {
	mu         sync.RWMutex
	extensions map[string]Extension
}

var globalRegistry = &ExtensionRegistry{
	extensions: make(map[string]Extension),
}

// Register는 확장 모듈을 레지스트리에 등록합니다.
// 주로 extension 패키지의 init() 함수에서 호출됩니다.
func Register(ext Extension) {
	globalRegistry.mu.Lock()
	defer globalRegistry.mu.Unlock()

	name := ext.Name()
	if _, exists := globalRegistry.extensions[name]; exists {
		slog.Warn("Extension already registered, overwriting", "extension", name)
	}

	globalRegistry.extensions[name] = ext
	slog.Debug("Registered extension", "extension", name)
}

// loadExtensions는 등록된 모든 확장 모듈을 로드합니다.
func loadExtensions(config common.Config) error {
	// Lock을 짧게 유지하기 위해 확장 목록 복사
	globalRegistry.mu.RLock()
	extensions := make(map[string]Extension, len(globalRegistry.extensions))
	maps.Copy(extensions, globalRegistry.extensions)
	globalRegistry.mu.RUnlock()

	// Lock 해제 후 확장 로드
	for name, ext := range extensions {
		if err := ext.Load(config); err != nil {
			slog.Error("Failed to load extension", "extension", name, "error", err)
			return err
		}
		slog.Info("Loaded extension", "extension", name)
	}

	return nil
}

// unloadExtensions는 등록된 모든 확장 모듈을 언로드합니다.
func unloadExtensions() {
	// Lock을 짧게 유지하기 위해 확장 목록 복사
	globalRegistry.mu.RLock()
	extensions := make(map[string]Extension, len(globalRegistry.extensions))
	maps.Copy(extensions, globalRegistry.extensions)
	globalRegistry.mu.RUnlock()

	// Lock 해제 후 확장 언로드
	for name, ext := range extensions {
		ext.Unload()
		slog.Info("Unloaded extension", "extension", name)
	}
}
