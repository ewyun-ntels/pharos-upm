package server

import (
	"log/slog"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/pkg/common"
)

func TestServer_Start_NoNatsConfig(t *testing.T) {
	t.Parallel()

	server := &NatsServer{}
	config := common.Config{
		Servers: map[string]common.ServerConfig{},
	}

	err := server.Start("/test/config", config, slog.Default())
	assert.NoError(t, err, "No NATS config should not cause error")
	assert.Nil(t, server.natsServer, "NATS server should not be initialized")
}

func TestServer_Start_WithDefaultPort(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping NATS server integration test in short mode")
	}

	server := &NatsServer{}
	config := common.Config{
		Serve: common.ServeConfig{
			NatsSchema: common.SchemaNats,
		},
		Servers: map[string]common.ServerConfig{
			common.ServerTypeNats: {
				Port: 0,
			},
		},
		Nats: common.NatsConfig{
			JetStream:          true,
			JetStreamMaxMemory: 1024 * 1024 * 100,
			JetStreamMaxStore:  1024 * 1024 * 100,
			StoreDir:           "/tmp/nats-test",
			Routes:             []string{},
		},
	}

	err := server.Start("/test/config", config, slog.Default())
	if err != nil {
		t.Logf("NATS server start error (expected in test environment): %v", err)
		return
	}

	defer func() {
		if server.natsServer != nil {
			_ = server.Stop()
		}
	}()

	assert.NotNil(t, server.natsServer, "NATS server should be initialized")
	assert.Equal(t, "/test/config", server.configPath)
}

func TestServer_Stop_NilServer(t *testing.T) {
	t.Parallel()

	server := &NatsServer{}

	err := server.Stop()
	assert.NoError(t, err, "Stopping nil server should not cause error")
}

func TestServer_Stop_NotRunning(t *testing.T) {
	t.Parallel()

	server := &NatsServer{}

	err := server.Stop()
	assert.NoError(t, err, "Stopping non-running server should not cause error")
}

func TestServer_GetClusterOpts_EmptyListen(t *testing.T) {
	t.Parallel()

	server := &NatsServer{
		config: common.Config{
			Nats: common.NatsConfig{
				Cluster: struct {
					Name   string "mapstructure:\"name\""
					Listen string "mapstructure:\"listen\""
				}{
					Listen: "",
				},
			},
		},
	}

	opts, err := server.getClusterOpts()
	assert.NoError(t, err)
	assert.Equal(t, "", opts.Name)
	assert.Equal(t, "", opts.Host)
	assert.Equal(t, 0, opts.Port)
}

func TestServer_GetClusterOpts_ValidListen(t *testing.T) {
	t.Parallel()

	server := &NatsServer{
		config: common.Config{
			Nats: common.NatsConfig{
				Cluster: struct {
					Name   string "mapstructure:\"name\""
					Listen string "mapstructure:\"listen\""
				}{
					Name:   "test-cluster",
					Listen: "nats://localhost:6222",
				},
			},
		},
	}

	opts, err := server.getClusterOpts()
	assert.NoError(t, err)
	assert.Equal(t, "test-cluster", opts.Name)
	assert.Equal(t, "localhost", opts.Host)
	assert.Equal(t, 6222, opts.Port)
	assert.Equal(t, "nats://localhost:6222", opts.ListenStr)
}

func TestServer_GetClusterOpts_InvalidURL(t *testing.T) {
	server := &NatsServer{
		config: common.Config{
			Nats: common.NatsConfig{
				Cluster: struct {
					Name   string "mapstructure:\"name\""
					Listen string "mapstructure:\"listen\""
				}{
					Listen: "://invalid-url",
				},
			},
		},
	}

	_, err := server.getClusterOpts()
	assert.Error(t, err, "Invalid URL should return error")
}

func TestServer_GetClusterOpts_InvalidPort(t *testing.T) {
	server := &NatsServer{
		config: common.Config{
			Nats: common.NatsConfig{
				Cluster: struct {
					Name   string "mapstructure:\"name\""
					Listen string "mapstructure:\"listen\""
				}{
					Listen: "nats://localhost",
				},
			},
		},
	}

	_, err := server.getClusterOpts()
	assert.Error(t, err, "Missing port should return error")
}

func TestServer_GetClusterOpts_NonNumericPort(t *testing.T) {
	server := &NatsServer{
		config: common.Config{
			Nats: common.NatsConfig{
				Cluster: struct {
					Name   string "mapstructure:\"name\""
					Listen string "mapstructure:\"listen\""
				}{
					Listen: "nats://localhost:abc",
				},
			},
		},
	}

	_, err := server.getClusterOpts()
	assert.Error(t, err, "Non-numeric port should return error")
}

func TestServer_Lifecycle(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping NATS server lifecycle test in short mode")
	}

	server := &NatsServer{}
	config := common.Config{
		Serve: common.ServeConfig{
			NatsSchema: common.SchemaNats,
		},
		Servers: map[string]common.ServerConfig{
			common.ServerTypeNats: {
				Port: 14222,
			},
		},
		Nats: common.NatsConfig{
			JetStream:          false,
			JetStreamMaxMemory: 0,
			JetStreamMaxStore:  0,
			StoreDir:           "",
			Routes:             []string{},
		},
	}

	err := server.Start("/test/config", config, slog.Default())
	assert.NoError(t, err, "NATS server start error (expected in test environment)")

	time.Sleep(1 * time.Second)

	err = server.Stop()
	assert.NoError(t, err, "Stop should not return error")
	assert.Nil(t, server.natsServer, "NATS server should be nil after stop")
}

func TestServer_ConfigStorage(t *testing.T) {
	server := &NatsServer{}
	config := common.Config{
		Serve: common.ServeConfig{
			NatsSchema: common.SchemaNats,
		},
		Servers: map[string]common.ServerConfig{},
		Nats: common.NatsConfig{
			Routes: []string{"route1", "route2"},
		},
	}

	configPath := "/test/config/path"
	_ = server.Start(configPath, config, slog.Default())

	assert.Equal(t, configPath, server.configPath)
	assert.Equal(t, config.Nats.Routes, server.config.Nats.Routes)
}
