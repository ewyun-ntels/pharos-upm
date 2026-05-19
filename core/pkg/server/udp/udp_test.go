package udp

import (
	"testing"

	"github.com/panjf2000/gnet/v2"
	"github.com/stretchr/testify/assert"
)

func TestAddGnetServer(t *testing.T) {
	// Create a simple event handler for testing
	handler := &gnet.BuiltinEventEngine{}

	tests := []struct {
		name    string
		svrName string
		port    int
		handler gnet.EventHandler
	}{
		{
			name:    "Add UDP server on port 8080",
			svrName: "test-server-1",
			port:    8080,
			handler: handler,
		},
		{
			name:    "Add UDP server on port 9090",
			svrName: "test-server-2",
			port:    9090,
			handler: handler,
		},
		{
			name:    "Add UDP server with different name",
			svrName: "another-server",
			port:    7777,
			handler: handler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Add server
			AddGnetServer(tt.svrName, tt.port, tt.handler)

			// Verify server was added
			server := Servers.Get(tt.svrName)
			assert.NotNil(t, server, "Expected server to be non-nil")
		})
	}
}

func TestAddGnetServer_Multiple(t *testing.T) {
	handler1 := &gnet.BuiltinEventEngine{}
	handler2 := &gnet.BuiltinEventEngine{}

	// Add multiple servers
	AddGnetServer("server-a", 5000, handler1)
	AddGnetServer("server-b", 5001, handler2)
	AddGnetServer("server-c", 5002, handler1)

	// Verify all servers exist
	serverA := Servers.Get("server-a")
	assert.NotNil(t, serverA)

	serverB := Servers.Get("server-b")
	assert.NotNil(t, serverB)

	serverC := Servers.Get("server-c")
	assert.NotNil(t, serverC)
}

func TestAddGnetServer_Overwrite(t *testing.T) {
	handler1 := &gnet.BuiltinEventEngine{}
	handler2 := &gnet.BuiltinEventEngine{}

	name := "test-overwrite"

	// Add server first time
	AddGnetServer(name, 6000, handler1)
	server1 := Servers.Get(name)
	assert.NotNil(t, server1)

	// Add server second time with different port (overwrite)
	AddGnetServer(name, 6001, handler2)
	server2 := Servers.Get(name)
	assert.NotNil(t, server2)

	// Both should exist (Map stores latest value)
	assert.NotNil(t, server1)
	assert.NotNil(t, server2)
}

func TestServers_MapExists(t *testing.T) {
	// Verify Servers map is not nil
	assert.NotNil(t, Servers, "Expected Servers map to be initialized")
}

func BenchmarkAddGnetServer(b *testing.B) {
	handler := &gnet.BuiltinEventEngine{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddGnetServer("benchmark-server", 10000+i%1000, handler)
	}
}
