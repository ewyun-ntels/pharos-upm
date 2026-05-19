package gnet

import (
	"log/slog"
	"testing"

	"github.com/panjf2000/gnet/v2"
	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/pkg/common"
)

func TestServer_getAddress(t *testing.T) {
	tests := []struct {
		name     string
		protocol string
		port     int
		expected string
	}{
		{
			name:     "UDP address",
			protocol: "udp",
			port:     8080,
			expected: "udp://:8080",
		},
		{
			name:     "TCP address",
			protocol: "tcp",
			port:     9090,
			expected: "tcp://:9090",
		},
		{
			name:     "TCP on port 80",
			protocol: "tcp",
			port:     80,
			expected: "tcp://:80",
		},
		{
			name:     "UDP on high port",
			protocol: "udp",
			port:     65535,
			expected: "udp://:65535",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := &Server{
				Protocol:     tt.protocol,
				Port:         tt.port,
				EventHandler: &gnet.BuiltinEventEngine{},
			}

			address := server.getAddress()
			assert.Equal(t, tt.expected, address)
		})
	}
}

func TestServer_Fields(t *testing.T) {
	handler := &gnet.BuiltinEventEngine{}

	server := &Server{
		Protocol:     "tcp",
		Port:         8080,
		EventHandler: handler,
	}

	assert.Equal(t, "tcp", server.Protocol)
	assert.Equal(t, 8080, server.Port)
	assert.NotNil(t, server.EventHandler)
}

func TestServer_Start(t *testing.T) {
	handler := &gnet.BuiltinEventEngine{}

	server := &Server{
		Protocol:     "tcp",
		Port:         9999,
		EventHandler: handler,
	}

	config := common.Config{}
	configPath := "/tmp/test-config.toml"
	logger := slog.Default()

	// Start returns immediately (goroutine), should not error
	err := server.Start(configPath, config, logger)
	assert.NoError(t, err, "Expected Start to return without error")

	// Note: Actual server startup happens in goroutine
	// We're testing the interface, not the actual network binding
}

func TestServer_MultipleInstances(t *testing.T) {
	handler1 := &gnet.BuiltinEventEngine{}
	handler2 := &gnet.BuiltinEventEngine{}

	server1 := &Server{
		Protocol:     "udp",
		Port:         10001,
		EventHandler: handler1,
	}

	server2 := &Server{
		Protocol:     "tcp",
		Port:         10002,
		EventHandler: handler2,
	}

	assert.Equal(t, "udp://:10001", server1.getAddress())
	assert.Equal(t, "tcp://:10002", server2.getAddress())
	assert.NotEqual(t, server1.Port, server2.Port)
	assert.NotEqual(t, server1.Protocol, server2.Protocol)
}

func BenchmarkServer_getAddress(b *testing.B) {
	server := &Server{
		Protocol:     "tcp",
		Port:         8080,
		EventHandler: &gnet.BuiltinEventEngine{},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = server.getAddress()
	}
}
