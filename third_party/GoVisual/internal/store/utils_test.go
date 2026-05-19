package store

import (
	"encoding/base64"
	"net/http"
	"testing"
)

func TestGetJwtSubject(t *testing.T) {
	tests := []struct {
		name        string
		header      http.Header
		expected    string
		expectError bool
	}{
		{
			name:        "empty authorization header",
			header:      http.Header{},
			expected:    "",
			expectError: false,
		},
		{
			name: "invalid authorization format",
			header: http.Header{
				"Authorization": []string{"InvalidFormat"},
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "basic auth with valid client credentials",
			header: http.Header{
				"Authorization": []string{"Basic " + base64.StdEncoding.EncodeToString([]byte("test_client:test_secret"))},
			},
			expected:    "",
			expectError: false,
		},
		{
			name: "basic auth with empty client_id",
			header: http.Header{
				"Authorization": []string{"Basic " + base64.StdEncoding.EncodeToString([]byte(":test_secret"))},
			},
			expected:    "",
			expectError: false,
		},
		{
			name: "basic auth with invalid base64",
			header: http.Header{
				"Authorization": []string{"Basic invalid_base64!"},
			},
			expected:    "",
			expectError: false,
		},
		{
			name: "basic auth with invalid format (no colon)",
			header: http.Header{
				"Authorization": []string{"Basic " + base64.StdEncoding.EncodeToString([]byte("no_colon_format"))},
			},
			expected:    "",
			expectError: false,
		},
		{
			name: "bearer token (JWT) - will fail without valid JWT",
			header: http.Header{
				"Authorization": []string{"Bearer invalid.jwt.token"},
			},
			expected:    "",
			expectError: true,
		},
		{
			name: "unsupported auth type",
			header: http.Header{
				"Authorization": []string{"Digest some_digest_value"},
			},
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := getJwtSubject(tt.header)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %q but got %q", tt.expected, result)
			}
		})
	}
}

func TestExtractBasicAuthClientId(t *testing.T) {
	tests := []struct {
		name        string
		encoded     string
		expected    string
		expectError bool
	}{
		{
			name:        "valid client credentials",
			encoded:     base64.StdEncoding.EncodeToString([]byte("my_client_id:my_secret")),
			expected:    "client:my_client_id",
			expectError: false,
		},
		{
			name:        "client with complex id",
			encoded:     base64.StdEncoding.EncodeToString([]byte("oauth-client-123:very-secret-key-456")),
			expected:    "client:oauth-client-123",
			expectError: false,
		},
		{
			name:        "invalid base64",
			encoded:     "invalid-base64!@#",
			expected:    "",
			expectError: true,
		},
		{
			name:        "missing colon separator",
			encoded:     base64.StdEncoding.EncodeToString([]byte("no_colon_here")),
			expected:    "",
			expectError: true,
		},
		{
			name:        "empty client_id",
			encoded:     base64.StdEncoding.EncodeToString([]byte(":secret")),
			expected:    "",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractBasicAuthClientId(tt.encoded)

			if tt.expectError && err == nil {
				t.Errorf("expected error but got none")
			}

			if !tt.expectError && err != nil {
				t.Errorf("unexpected error: %v", err)
			}

			if result != tt.expected {
				t.Errorf("expected %q but got %q", tt.expected, result)
			}
		})
	}
}
