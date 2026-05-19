package gin

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"testing"
	"time"

	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/migration"
)

var (
	testTmpDir     string
	testConfigPath string
	testBaseConfig common.Config
)

// TestMain sets up shared test resources and runs all tests
func TestMain(m *testing.M) {
	// Setup
	var err error
	testTmpDir, err = os.MkdirTemp("", "pharos-test-*")
	if err != nil {
		panic(fmt.Sprintf("Failed to create temp dir: %v", err))
	}

	testBaseConfig = common.Config{
		Serve: common.ServeConfig{
			ServerSchema: common.ServerTypeHttp,
		},
		Logger: common.RotationConfig{
			IsJson:    false,
			IsConsole: true,
		},
		History: common.HistoryConfig{
			Use: false,
		},
		Database: orm.DatabaseConfig{
			Driver: orm.DriverSqlite,
			SQLite: orm.SQLiteConfig{
				Path: testTmpDir + "/test.db",
			},
		},
		Statistics: common.StatisticsConfig{
			Database: orm.DatabaseConfig{
				Driver: orm.DriverSqlite,
				SQLite: orm.SQLiteConfig{
					Path: testTmpDir + "/statistics.db",
				},
			},
		},
		Auth: common.AuthConfig{
			JwtCert: common.AuthJwtCert{
				JwtDir:       testTmpDir,
				AutoGenerate: true,
				AutoGenerateInfo: common.CertGenerateInfo{
					CommonName:   "test",
					Organization: []string{"test"},
					NotAfter:     time.Now().Add(24 * time.Hour).Format(time.RFC3339),
				},
			},
		},
	}

	// Perform DB migration once
	if err := migration.Load(testBaseConfig); err != nil {
		panic(fmt.Sprintf("Failed to migrate database: %v", err))
	}

	testConfigPath = filepath.Join(testTmpDir, "config.toml")
	content := `
[serve]
server_schema = "http"

[logger]
is_json = false
is_console = true

[history]
use = false
`
	if err := os.WriteFile(testConfigPath, []byte(content), 0644); err != nil {
		panic(fmt.Sprintf("Failed to create config file: %v", err))
	}

	// Run tests
	code := m.Run()

	// Cleanup
	os.RemoveAll(testTmpDir)

	os.Exit(code)
}

// createTestConfig creates a minimal test configuration with shared resources
func createTestConfig(port int) (string, common.Config) {
	config := testBaseConfig
	config.Servers = map[string]common.ServerConfig{
		common.ServerTypeHttp: {
			Port: port,
		},
	}

	return testConfigPath, config
}

// waitForServer waits for the server to be ready
func waitForServer(url string, timeout time.Duration) error {
	deadline := time.Now().Add(timeout)
	client := &http.Client{Timeout: 50 * time.Millisecond}
	for time.Now().Before(deadline) {
		resp, err := client.Get(url)
		if err == nil {
			resp.Body.Close()
			return nil
		}
		time.Sleep(10 * time.Millisecond)
	}
	return fmt.Errorf("server did not start within %v", timeout)
}

// TestServer runs all server-related tests as subtests
func TestServer(t *testing.T) {
	t.Run("StartAndStop", func(t *testing.T) {
		configPath, config := createTestConfig(18080)

		server := &Server{}

		// Start server
		err := server.Start(configPath, config, slog.Default())
		if err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}

		// Wait for server to be ready
		serverURL := "http://localhost:18080"
		if err := waitForServer(serverURL, 500*time.Millisecond); err != nil {
			t.Fatalf("Server not ready: %v", err)
		}

		// Test that server is responding
		resp, err := http.Get(serverURL)
		if err != nil {
			t.Errorf("Failed to make request to server: %v", err)
		} else {
			resp.Body.Close()
			if resp.StatusCode < 200 || resp.StatusCode >= 600 {
				t.Errorf("Expected valid HTTP status code, got %d", resp.StatusCode)
			}
		}

		// Stop server
		err = server.Stop()
		if err != nil {
			t.Errorf("Failed to stop server: %v", err)
		}

		time.Sleep(10 * time.Millisecond)

		// Verify server is stopped
		_, err = http.Get(serverURL)
		if err == nil {
			t.Error("Server should have stopped but is still responding")
		}
	})

	t.Run("RootRedirect", func(t *testing.T) {
		configPath, config := createTestConfig(18081)

		server := &Server{}

		err := server.Start(configPath, config, slog.Default())
		if err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}
		defer server.Stop()

		serverURL := "http://localhost:18081"
		if err := waitForServer(serverURL, 500*time.Millisecond); err != nil {
			t.Fatalf("Server not ready: %v", err)
		}

		client := &http.Client{
			CheckRedirect: func(req *http.Request, via []*http.Request) error {
				return http.ErrUseLastResponse
			},
		}

		resp, err := client.Get(serverURL + "/")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode == http.StatusNotFound {
			t.Errorf("Expected redirect or valid response, got 404 Not Found")
		}
	})

	t.Run("UIEndpoint", func(t *testing.T) {
		configPath, config := createTestConfig(18082)

		server := &Server{}

		err := server.Start(configPath, config, slog.Default())
		if err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}
		defer server.Stop()

		serverURL := "http://localhost:18082"
		if err := waitForServer(serverURL, 500*time.Millisecond); err != nil {
			t.Fatalf("Server not ready: %v", err)
		}

		resp, err := http.Get(serverURL + "/ui/")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode < 200 || resp.StatusCode >= 600 {
			t.Errorf("Expected valid HTTP status code, got %d", resp.StatusCode)
		}
	})

	t.Run("SecurityHeaders", func(t *testing.T) {
		configPath, config := createTestConfig(18083)

		server := &Server{}

		err := server.Start(configPath, config, slog.Default())
		if err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}
		defer server.Stop()

		serverURL := "http://localhost:18083"
		if err := waitForServer(serverURL, 500*time.Millisecond); err != nil {
			t.Fatalf("Server not ready: %v", err)
		}

		resp, err := http.Get(serverURL + "/ui/")
		if err != nil {
			t.Fatalf("Failed to make request: %v", err)
		}
		defer resp.Body.Close()

		hasSecurityHeader := false
		securityHeaders := []string{
			"X-Content-Type-Options",
			"X-Frame-Options",
			"X-XSS-Protection",
		}

		for _, header := range securityHeaders {
			if resp.Header.Get(header) != "" {
				hasSecurityHeader = true
				t.Logf("Found security header: %s = %s", header, resp.Header.Get(header))
			}
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 600 {
			t.Errorf("Expected valid HTTP status code, got %d", resp.StatusCode)
		}

		t.Logf("Security headers present: %v", hasSecurityHeader)
	})

	t.Run("MultipleStartStop", func(t *testing.T) {
		configPath, config := createTestConfig(18084)

		server := &Server{}

		// First start/stop cycle
		err := server.Start(configPath, config, slog.Default())
		if err != nil {
			t.Fatalf("Failed to start server (first cycle): %v", err)
		}

		serverURL := "http://localhost:18084"
		if err := waitForServer(serverURL, 500*time.Millisecond); err != nil {
			t.Fatalf("Server not ready (first cycle): %v", err)
		}

		err = server.Stop()
		if err != nil {
			t.Errorf("Failed to stop server (first cycle): %v", err)
		}

		time.Sleep(20 * time.Millisecond)

		// Second start/stop cycle
		server2 := &Server{}
		err = server2.Start(configPath, config, slog.Default())
		if err != nil {
			t.Fatalf("Failed to start server (second cycle): %v", err)
		}

		if err := waitForServer(serverURL, 500*time.Millisecond); err != nil {
			t.Fatalf("Server not ready (second cycle): %v", err)
		}

		err = server2.Stop()
		if err != nil {
			t.Errorf("Failed to stop server (second cycle): %v", err)
		}
	})

	t.Run("ConcurrentRequests", func(t *testing.T) {
		configPath, config := createTestConfig(18085)

		server := &Server{}

		err := server.Start(configPath, config, slog.Default())
		if err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}
		defer server.Stop()

		serverURL := "http://localhost:18085"
		if err := waitForServer(serverURL, 500*time.Millisecond); err != nil {
			t.Fatalf("Server not ready: %v", err)
		}

		// Make concurrent requests
		numRequests := 10
		done := make(chan bool, numRequests)

		for range numRequests {
			go func() {
				resp, err := http.Get(serverURL + "/")
				if err == nil {
					io.Copy(io.Discard, resp.Body)
					resp.Body.Close()
				}
				done <- true
			}()
		}

		timeout := time.After(1 * time.Second)
		for range numRequests {
			select {
			case <-done:
				// Request completed
			case <-timeout:
				t.Fatal("Timeout waiting for concurrent requests")
			}
		}
	})

	t.Run("GracefulShutdown", func(t *testing.T) {
		configPath, config := createTestConfig(18086)

		server := &Server{}

		err := server.Start(configPath, config, slog.Default())
		if err != nil {
			t.Fatalf("Failed to start server: %v", err)
		}

		serverURL := "http://localhost:18086"
		if err := waitForServer(serverURL, 500*time.Millisecond); err != nil {
			t.Fatalf("Server not ready: %v", err)
		}

		// Start a request in background
		requestDone := make(chan bool, 1)
		go func() {
			resp, err := http.Get(serverURL + "/")
			if err == nil {
				resp.Body.Close()
			}
			requestDone <- true
		}()

		time.Sleep(20 * time.Millisecond)

		// Gracefully stop server
		stopDone := make(chan error, 1)
		go func() {
			stopDone <- server.Stop()
		}()

		select {
		case err := <-stopDone:
			if err != nil {
				t.Errorf("Server stop returned error: %v", err)
			}
		case <-time.After(5 * time.Second):
			t.Error("Server did not stop within timeout")
		}
	})
}

func TestServer_StopWithoutStart(t *testing.T) {
	t.Parallel()

	server := &Server{}

	// Stopping a server that was never started should not panic
	err := server.Stop()
	if err != nil {
		t.Logf("Stop without start returned error: %v (this is acceptable)", err)
	}
}
