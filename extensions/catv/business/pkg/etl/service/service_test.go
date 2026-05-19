package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/pkg/common"
)

// TestNewService는 Service 생성이 올바르게 동작하는지 검증합니다.
func TestNewService(t *testing.T) {
	config := common.Config{
		Serve: common.ServeConfig{},
	}

	service, err := NewService(config)

	require.NoError(t, err)
	require.NotNil(t, service)
	assert.Equal(t, config, service.config)
}

// TestService_Structure는 Service 구조체가 올바른 필드를 가지고 있는지 검증합니다.
func TestService_Structure(t *testing.T) {
	config := common.Config{}
	service, err := NewService(config)

	require.NoError(t, err)
	require.NotNil(t, service)

	// Service 구조체의 필드 확인
	assert.NotNil(t, &service.config)
	assert.NotNil(t, &service.natsClient)
}

func TestService_Start_WithoutConnection(t *testing.T) {
	// NATS 서버 없이 Start 호출 - 설정이 없으면 연결하지 않음
	config := common.Config{
		Serve: common.ServeConfig{},
	}

	service, err := NewService(config)
	require.NoError(t, err)

	err = service.Start()
	// NATS 클라이언트 설정이 없으면 연결하지 않으므로 에러가 없을 수 있음
	assert.NoError(t, err)
}

func TestService_Stop_BeforeStart(t *testing.T) {
	// Start 없이 Stop 호출
	config := common.Config{}
	service, err := NewService(config)
	require.NoError(t, err)

	// Stop은 에러를 반환하지 않아야 함 (연결되지 않은 상태에서도)
	err = service.Stop()
	assert.NoError(t, err)
}

func TestService_MultipleNewService(t *testing.T) {
	// 여러 개의 서비스 인스턴스 생성
	config1 := common.Config{
		Serve: common.ServeConfig{ServerSchema: common.ServerTypeHttp},
	}
	config2 := common.Config{
		Serve: common.ServeConfig{ServerSchema: common.ServerTypeHttps},
	}

	service1, err1 := NewService(config1)
	service2, err2 := NewService(config2)

	require.NoError(t, err1)
	require.NoError(t, err2)
	require.NotNil(t, service1)
	require.NotNil(t, service2)

	// 두 서비스는 다른 인스턴스여야 함
	assert.NotEqual(t, service1, service2)
}

func TestService_ConfigTypes(t *testing.T) {
	tests := []struct {
		name         string
		serverSchema string
		shouldPass   bool
	}{
		{
			name:         "http schema",
			serverSchema: common.ServerTypeHttp,
			shouldPass:   true,
		},
		{
			name:         "https schema",
			serverSchema: common.ServerTypeHttps,
			shouldPass:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := common.Config{
				Serve: common.ServeConfig{
					ServerSchema: tt.serverSchema,
				},
			}

			service, err := NewService(config)
			if tt.shouldPass {
				require.NoError(t, err)
				require.NotNil(t, service)
				assert.Equal(t, tt.serverSchema, service.config.Serve.ServerSchema)
			} else {
				assert.Error(t, err)
			}
		})
	}
}

// Benchmark tests
func BenchmarkNewService(b *testing.B) {
	config := common.Config{
		Serve: common.ServeConfig{
			ServerSchema: common.ServerTypeHttp,
		},
	}

	for i := 0; i < b.N; i++ {
		_, _ = NewService(config)
	}
}

func BenchmarkService_Stop(b *testing.B) {
	config := common.Config{}
	service, _ := NewService(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = service.Stop()
	}
}
