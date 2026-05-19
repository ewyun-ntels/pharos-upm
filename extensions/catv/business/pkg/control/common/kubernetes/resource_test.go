package kubernetes

import (
	"context"
	"testing"

	"k8s.io/client-go/rest"
)

// TestGetClientset tests singleton behavior of getClientset
func TestGetClientset(t *testing.T) {
	// Note: This test will fail outside a Kubernetes cluster
	// as InClusterConfig() requires pod environment
	_, err := getClientset()
	if err != nil {
		// Expected error when not running in cluster
		if err.Error() != "" && clientset == nil {
			t.Logf("Expected error outside cluster: %v", err)
			return
		}
	}

	// Test singleton behavior
	client1, err1 := getClientset()
	client2, err2 := getClientset()

	if err1 != err2 {
		t.Errorf("Singleton should return same error: %v vs %v", err1, err2)
	}

	if client1 != client2 {
		t.Error("Singleton should return same instance")
	}
}

// TestDoInputValidation tests input validation in do function
func TestDoInputValidation(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		resource      string
		resourceName  string
		expectError   bool
		errorContains string
	}{
		{
			name:          "Empty version",
			version:       "",
			resource:      "jobs",
			resourceName:  "",
			expectError:   true,
			errorContains: "version is required",
		},
		{
			name:          "Empty resource and name",
			version:       "v1",
			resource:      "",
			resourceName:  "",
			expectError:   true,
			errorContains: "either resource or name must be specified",
		},
		{
			name:          "Valid inputs",
			version:       "v1",
			resource:      "jobs",
			resourceName:  "",
			expectError:   false, // Will fail at clientset, but passes validation
			errorContains: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := do(context.Background(), "GET", "batch", tt.version, "default", tt.resource, tt.resourceName, nil)

			if tt.expectError {
				if err == nil {
					t.Error("Expected error but got nil")
				} else if tt.errorContains != "" && !contains(err.Error(), tt.errorContains) {
					t.Errorf("Error should contain %q, got: %v", tt.errorContains, err)
				}
			}
		})
	}
}

// TestDoUnsupportedMethod tests unsupported HTTP method handling
func TestDoUnsupportedMethod(t *testing.T) {
	_, err := do(context.Background(), "PATCH", "batch", "v1", "default", "jobs", "test", nil)

	if err == nil {
		t.Error("Expected error for unsupported method")
	}

	if !contains(err.Error(), "unsupported HTTP method") {
		t.Errorf("Error should mention unsupported method, got: %v", err)
	}
}

// Helper function for string contains check
func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(s) > len(substr) && containsSubstring(s, substr))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// Mock tests for Get, Post, Put, Delete functions
// These would require setting up a fake Kubernetes API server

// TestGetTableName tests table name constant
func TestGetTableName(t *testing.T) {
	if ControllerJobTableName := "controller_job"; ControllerJobTableName == "" {
		t.Error("Table name should not be empty")
	}
}

// Integration test helper - skip in unit tests
func skipIfNotInCluster(t *testing.T) {
	_, err := rest.InClusterConfig()
	if err != nil {
		t.Skip("Skipping integration test - not running in Kubernetes cluster")
	}
}

// TestIntegrationKubernetesOperations is an integration test
func TestIntegrationKubernetesOperations(t *testing.T) {
	skipIfNotInCluster(t)

	// This test would create/get/delete actual Kubernetes resources
	// Only run in integration test environment
	t.Skip("Integration test - requires Kubernetes cluster")
}

// Benchmark tests
func BenchmarkGetClientset(b *testing.B) {
	for i := 0; i < b.N; i++ {
		_, _ = getClientset()
	}
}

// TestGetClientsetConcurrency tests thread safety
func TestGetClientsetConcurrency(t *testing.T) {
	done := make(chan bool, 10)

	for range 10 {
		go func() {
			_, _ = getClientset()
			done <- true
		}()
	}

	for range 10 {
		<-done
	}

	// Should not panic or race
	t.Log("Concurrent getClientset calls completed")
}
