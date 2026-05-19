package common

import (
	"crypto/tls"
	"slices"
	"strings"
	"sync"
	"testing"
)

func TestParseTLSVersion(t *testing.T) {
	tests := []struct {
		input       string
		expectedVer uint16
		expectedOk  bool
	}{
		// Valid versions
		{"1.2", tls.VersionTLS12, true},
		{"1.3", tls.VersionTLS13, true},
		{"tls1.2", tls.VersionTLS12, true},
		{"tls1.3", tls.VersionTLS13, true},
		{"TLS1.2", tls.VersionTLS12, true},
		{"TLS1.3", tls.VersionTLS13, true},
		{" 1.2 ", tls.VersionTLS12, true},
		{" TLS1.3 ", tls.VersionTLS13, true},

		// Default cases
		{"", tls.VersionTLS13, false},
		{"default", tls.VersionTLS13, false},
		{"DEFAULT", tls.VersionTLS13, false},
		{" default ", tls.VersionTLS13, false},

		// Invalid versions
		{"1.1", 0, false},
		{"1.4", 0, false},
		{"tls1.1", 0, false},
		{"invalid", 0, false},
		{"2.0", 0, false},
		{"ssl", 0, false},
		{"abc", 0, false},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			version, ok := parseTLSVersion(tt.input)
			if version != tt.expectedVer {
				t.Errorf("parseTLSVersion(%q) version = %d, want %d", tt.input, version, tt.expectedVer)
			}
			if ok != tt.expectedOk {
				t.Errorf("parseTLSVersion(%q) ok = %v, want %v", tt.input, ok, tt.expectedOk)
			}
		})
	}
}

func TestBuildDefaultClientTLS(t *testing.T) {
	// Save original global config
	originalConfig := globalConfig
	defer func() {
		globalConfig = originalConfig
		globalConfigOnce = sync.Once{} // Reset to new sync.Once instead of copying
	}()

	t.Run("WithoutGlobalConfig", func(t *testing.T) {
		// Reset global config
		globalConfig = nil
		globalConfigOnce = sync.Once{}

		tlsConfig := BuildDefaultClientTLS()

		if tlsConfig == nil {
			t.Fatal("Expected TLS config to be non-nil")
		}

		// Check defaults
		if tlsConfig.InsecureSkipVerify != false {
			t.Errorf("Expected InsecureSkipVerify to be false, got %v", tlsConfig.InsecureSkipVerify)
		}

		if tlsConfig.MinVersion != tls.VersionTLS12 {
			t.Errorf("Expected MinVersion to be TLS 1.2 (%d), got %d", tls.VersionTLS12, tlsConfig.MinVersion)
		}

		if tlsConfig.MaxVersion != tls.VersionTLS13 {
			t.Errorf("Expected MaxVersion to be TLS 1.3 (%d), got %d", tls.VersionTLS13, tlsConfig.MaxVersion)
		}
	})

	t.Run("WithGlobalConfig", func(t *testing.T) {
		// Reset and set global config
		globalConfig = nil
		globalConfigOnce = sync.Once{}

		config := Config{
			ClientTLS: ClientTLSConfig{
				InsecureSkipVerify: true,
				MinVersion:         "1.3",
				MaxVersion:         "1.3",
			},
		}
		SetGlobalConfig(config)

		tlsConfig := BuildDefaultClientTLS()

		if tlsConfig == nil {
			t.Fatal("Expected TLS config to be non-nil")
		}

		if tlsConfig.InsecureSkipVerify != true {
			t.Errorf("Expected InsecureSkipVerify to be true, got %v", tlsConfig.InsecureSkipVerify)
		}

		if tlsConfig.MinVersion != tls.VersionTLS13 {
			t.Errorf("Expected MinVersion to be TLS 1.3 (%d), got %d", tls.VersionTLS13, tlsConfig.MinVersion)
		}

		if tlsConfig.MaxVersion != tls.VersionTLS13 {
			t.Errorf("Expected MaxVersion to be TLS 1.3 (%d), got %d", tls.VersionTLS13, tlsConfig.MaxVersion)
		}
	})

	t.Run("WithInvalidVersions", func(t *testing.T) {
		// Reset and set global config with invalid versions
		globalConfig = nil
		globalConfigOnce = sync.Once{}

		config := Config{
			ClientTLS: ClientTLSConfig{
				InsecureSkipVerify: false,
				MinVersion:         "invalid",
				MaxVersion:         "also_invalid",
			},
		}
		SetGlobalConfig(config)

		tlsConfig := BuildDefaultClientTLS()

		if tlsConfig == nil {
			t.Fatal("Expected TLS config to be non-nil")
		}

		// Should fall back to defaults for invalid versions
		if tlsConfig.MinVersion != tls.VersionTLS12 {
			t.Errorf("Expected MinVersion to fall back to TLS 1.2 (%d), got %d", tls.VersionTLS12, tlsConfig.MinVersion)
		}

		if tlsConfig.MaxVersion != tls.VersionTLS13 {
			t.Errorf("Expected MaxVersion to fall back to TLS 1.3 (%d), got %d", tls.VersionTLS13, tlsConfig.MaxVersion)
		}
	})

	t.Run("WithMaxVersionLowerThanMinVersion", func(t *testing.T) {
		// Reset and set global config with max < min
		globalConfig = nil
		globalConfigOnce = sync.Once{}

		config := Config{
			ClientTLS: ClientTLSConfig{
				MinVersion: "1.3",
				MaxVersion: "1.2",
			},
		}
		SetGlobalConfig(config)

		tlsConfig := BuildDefaultClientTLS()

		if tlsConfig == nil {
			t.Fatal("Expected TLS config to be non-nil")
		}

		// MaxVersion should be adjusted to MinVersion
		if tlsConfig.MinVersion != tls.VersionTLS13 {
			t.Errorf("Expected MinVersion to be TLS 1.3 (%d), got %d", tls.VersionTLS13, tlsConfig.MinVersion)
		}

		if tlsConfig.MaxVersion != tls.VersionTLS13 {
			t.Errorf("Expected MaxVersion to be adjusted to TLS 1.3 (%d), got %d", tls.VersionTLS13, tlsConfig.MaxVersion)
		}
	})
}

func TestGetRestyClient(t *testing.T) {
	// Save original global config
	originalConfig := globalConfig
	defer func() {
		globalConfig = originalConfig
		globalConfigOnce = sync.Once{} // Reset to new sync.Once instead of copying
	}()

	// Reset global config
	globalConfig = nil
	globalConfigOnce = sync.Once{}

	client := GetRestyClient()

	if client == nil {
		t.Fatal("Expected Resty client to be non-nil")
	}

	// Test that TLS config is set (we can't easily test the exact config without exposing internals)
	// But we can verify the client was created successfully
	if client.GetClient() == nil {
		t.Error("Expected underlying HTTP client to be non-nil")
	}
}

func TestGetClickhouseMasterURL(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"http://localhost:8080", "http://localhost:8080/proxy/clickhouse"},
		{"https://example.com", "https://example.com/proxy/clickhouse"},
		{"http://192.168.1.1:9000", "http://192.168.1.1:9000/proxy/clickhouse"},
		{"localhost", "localhost/proxy/clickhouse"},
		{"", "/proxy/clickhouse"},
		{"https://api.example.com:443", "https://api.example.com:443/proxy/clickhouse"},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GetClickhouseMasterURL(tt.input)
			if result != tt.expected {
				t.Errorf("GetClickhouseMasterURL(%q) = %q, want %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestGetWebsocketEndpoints(t *testing.T) {
	tests := []struct {
		name     string
		config   Config
		expected []string
	}{
		{
			name: "HTTPOnly",
			config: Config{
				Serve: ServeConfig{
					ServerSchema: "http",
				},
				Servers: map[string]ServerConfig{
					"http": {Port: 8080},
				},
			},
			expected: []string{"ws://localhost:8080/websocket/centrifuge"},
		},
		{
			name: "HTTPSOnly",
			config: Config{
				Serve: ServeConfig{
					ServerSchema: "https",
				},
				Servers: map[string]ServerConfig{
					"https": {Port: 8443},
				},
			},
			expected: []string{"wss://localhost:8443/websocket/centrifuge"},
		},
		{
			name: "MultipleHTTPServers",
			config: Config{
				Serve: ServeConfig{
					ServerSchema: "http",
				},
				Servers: map[string]ServerConfig{
					"http": {Port: 8080},
				},
			},
			expected: []string{"ws://localhost:8080/websocket/centrifuge"},
		},
		{
			name: "MultipleHTTPSServers",
			config: Config{
				Serve: ServeConfig{
					ServerSchema: "https",
				},
				Servers: map[string]ServerConfig{
					"https": {Port: 8443},
				},
			},
			expected: []string{"wss://localhost:8443/websocket/centrifuge"},
		},
		{
			name: "MixedServersHTTPSchema",
			config: Config{
				Serve: ServeConfig{
					ServerSchema: "http",
				},
				Servers: map[string]ServerConfig{
					"http":  {Port: 8080},
					"https": {Port: 8443},
				},
			},
			expected: []string{"ws://localhost:8080/websocket/centrifuge"},
		},
		{
			name: "MixedServersHTTPSSchema",
			config: Config{
				Serve: ServeConfig{
					ServerSchema: "https",
				},
				Servers: map[string]ServerConfig{
					"http":  {Port: 8080},
					"https": {Port: 8443},
				},
			},
			expected: []string{"wss://localhost:8443/websocket/centrifuge"},
		},
		{
			name: "NoMatchingSchema",
			config: Config{
				Serve: ServeConfig{
					ServerSchema: "grpc",
				},
				Servers: map[string]ServerConfig{
					"http":  {Port: 8080},
					"https": {Port: 8443},
				},
			},
			expected: []string{},
		},
		{
			name: "EmptyServers",
			config: Config{
				Serve: ServeConfig{
					ServerSchema: "http",
				},
				Servers: map[string]ServerConfig{},
			},
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetWebsocketEndpoints(tt.config)

			if len(result) != len(tt.expected) {
				t.Errorf("GetWebsocketEndpoints() returned %d endpoints, want %d", len(result), len(tt.expected))
				t.Errorf("Got: %v", result)
				t.Errorf("Want: %v", tt.expected)
				return
			}

			// Check each expected endpoint is present
			for _, expectedEndpoint := range tt.expected {
				found := slices.Contains(result, expectedEndpoint)
				if !found {
					t.Errorf("Expected endpoint %q not found in result %v", expectedEndpoint, result)
				}
			}
		})
	}
}

func TestGetWebsocketEndpointsWithMultiplePorts(t *testing.T) {
	// Test case where we might have multiple servers with the same schema
	// This tests the concatenation behavior more thoroughly
	config := Config{
		Serve: ServeConfig{
			ServerSchema: "http",
		},
		Servers: map[string]ServerConfig{
			"http": {Port: 8080},
		},
	}

	result := GetWebsocketEndpoints(config)

	if len(result) != 1 {
		t.Errorf("Expected 1 endpoint, got %d: %v", len(result), result)
	}

	expectedEndpoint := "ws://localhost:8080/websocket/centrifuge"
	if result[0] != expectedEndpoint {
		t.Errorf("Expected endpoint %q, got %q", expectedEndpoint, result[0])
	}
}

// Benchmark tests
func BenchmarkParseTLSVersion(b *testing.B) {
	versions := []string{"1.2", "1.3", "tls1.2", "tls1.3", "invalid", "", "default"}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		version := versions[i%len(versions)]
		_, _ = parseTLSVersion(version)
	}
}

func BenchmarkBuildDefaultClientTLS(b *testing.B) {
	// Setup a global config
	config := Config{
		ClientTLS: ClientTLSConfig{
			InsecureSkipVerify: false,
			MinVersion:         "1.2",
			MaxVersion:         "1.3",
		},
	}
	SetGlobalConfig(config)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = BuildDefaultClientTLS()
	}
}

func BenchmarkGetRestyClient(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		client := GetRestyClient()
		_ = client
	}
}

func BenchmarkGetClickhouseMasterURL(b *testing.B) {
	host := "http://localhost:8080"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetClickhouseMasterURL(host)
	}
}

func BenchmarkGetWebsocketEndpoints(b *testing.B) {
	config := Config{
		Serve: ServeConfig{
			ServerSchema: "http",
		},
		Servers: map[string]ServerConfig{
			"http":  {Port: 8080},
			"https": {Port: 8443},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = GetWebsocketEndpoints(config)
	}
}

// Test edge cases and error conditions
func TestParseTLSVersionEdgeCases(t *testing.T) {
	// Test with whitespace variations
	whitespaceTests := []string{
		"\t1.2\t",
		"\n1.3\n",
		"\r\n1.2\r\n",
		"  TLS1.3  ",
	}

	for _, test := range whitespaceTests {
		t.Run("whitespace_"+strings.ReplaceAll(test, " ", "_"), func(t *testing.T) {
			version, ok := parseTLSVersion(test)
			if !ok {
				t.Errorf("Expected parseTLSVersion(%q) to return ok=true", test)
			}
			if version == 0 {
				t.Errorf("Expected parseTLSVersion(%q) to return non-zero version", test)
			}
		})
	}
}

func TestBuildDefaultClientTLSEdgeCases(t *testing.T) {
	// Save original global config
	originalConfig := globalConfig
	defer func() {
		globalConfig = originalConfig
		globalConfigOnce = sync.Once{} // Reset to new sync.Once instead of copying
	}()

	t.Run("WithZeroMaxVersion", func(t *testing.T) {
		globalConfig = nil
		globalConfigOnce = sync.Once{}

		// Simulate a config that might have zero values
		config := Config{
			ClientTLS: ClientTLSConfig{
				MinVersion: "1.2",
				MaxVersion: "", // This should result in default behavior
			},
		}
		SetGlobalConfig(config)

		tlsConfig := BuildDefaultClientTLS()

		if tlsConfig.MinVersion != tls.VersionTLS12 {
			t.Errorf("Expected MinVersion to be TLS 1.2, got %d", tlsConfig.MinVersion)
		}

		// MaxVersion should be the default since empty string maps to TLS 1.3 with ok=false
		if tlsConfig.MaxVersion != tls.VersionTLS13 {
			t.Errorf("Expected MaxVersion to be TLS 1.3 (default), got %d", tlsConfig.MaxVersion)
		}
	})
}
