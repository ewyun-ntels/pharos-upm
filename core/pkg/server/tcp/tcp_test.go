package tcp

import (
	"testing"

	"github.com/panjf2000/gnet/v2"
	"github.com/stretchr/testify/assert"
)

func TestAddGnetServer(t *testing.T) {
	handler := &gnet.BuiltinEventEngine{}

	tests := []struct {
		name    string
		svrName string
		port    int
		handler gnet.EventHandler
	}{
		{
			name:    "Add TCP server on port 8080",
			svrName: "tcp-server-1",
			port:    8080,
			handler: handler,
		},
		{
			name:    "Add TCP server on port 9090",
			svrName: "tcp-server-2",
			port:    9090,
			handler: handler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			AddGnetServer(tt.svrName, tt.port, tt.handler)

			server := Servers.Get(tt.svrName)
			assert.NotNil(t, server, "Expected server to be non-nil")
		})
	}
}

func TestAddGnetServer_Multiple(t *testing.T) {
	handler := &gnet.BuiltinEventEngine{}

	AddGnetServer("tcp-a", 7000, handler)
	AddGnetServer("tcp-b", 7001, handler)
	AddGnetServer("tcp-c", 7002, handler)

	serverA := Servers.Get("tcp-a")
	assert.NotNil(t, serverA)

	serverB := Servers.Get("tcp-b")
	assert.NotNil(t, serverB)

	serverC := Servers.Get("tcp-c")
	assert.NotNil(t, serverC)
}

func TestServers_MapExists(t *testing.T) {
	assert.NotNil(t, Servers, "Expected Servers map to be initialized")
}

func BenchmarkAddGnetServer(b *testing.B) {
	handler := &gnet.BuiltinEventEngine{}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		AddGnetServer("tcp-benchmark", 20000+i%1000, handler)
	}
}
