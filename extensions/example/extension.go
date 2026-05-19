package example

import (
	"log/slog"

	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/server"
)

// ExampleExtension은 테스트용 예제 확장 모듈입니다.
type ExampleExtension struct{}

func init() {
	// 빌드 시 SITE_MODE=example이 설정되면 자동으로 등록됩니다.
	server.Register(&ExampleExtension{})
}

func (e *ExampleExtension) Name() string {
	return "example"
}

func (e *ExampleExtension) Load(_ common.Config) error {
	slog.Info("Example extension loaded successfully")
	return nil
}

func (e *ExampleExtension) Unload() {
	slog.Info("Example extension unloaded")
}
