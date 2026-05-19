package command

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestLoad_ConfigInitialization은 실제 설정 파일이 필요한 통합 테스트입니다.
// 단위 테스트 환경에서는 스킵됩니다.
func TestLoad_ConfigInitialization(t *testing.T) {
	// 이 테스트는 실제 설정 파일이 필요하므로 스킵
	t.Skip("Requires actual config files")
}

// TestUnload_Cleanup은 unload 함수가 안전하게 실행되는지 검증합니다.
func TestUnload_Cleanup(t *testing.T) {
	// unload 함수가 패닉 없이 실행되는지 확인
	assert.NotPanics(t, func() {
		Unload()
	})
}

// TestLoad_InvalidConfigPath는 잘못된 설정 파일 경로에 대한 에러 처리를 검증합니다.
func TestLoad_InvalidConfigPath(t *testing.T) {
	// 존재하지 않는 설정 파일 경로로 로드 시도
	_, err := Load([]string{"/nonexistent/path/config.toml"})
	assert.Error(t, err)
}

// TestLoad_EmptyConfigPath는 빈 설정 경로에 대한 에러 처리를 검증합니다.
func TestLoad_EmptyConfigPath(t *testing.T) {
	// 빈 설정 경로로 로드 시도
	_, err := Load([]string{})
	// 설정 로더의 동작에 따라 에러가 발생하거나 기본값이 사용될 수 있음
	// 여기서는 에러가 발생할 것으로 예상
	assert.Error(t, err)
}

// Benchmark tests
func BenchmarkUnload(b *testing.B) {
	for i := 0; i < b.N; i++ {
		Unload()
	}
}
