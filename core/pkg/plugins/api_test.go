package plugins

import (
	"context"
	"sync"
	"testing"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/model"
)

// TestDatasourceQuery tests the DatasourceQuery function
func TestDatasourceQuery(t *testing.T) {
	// Test with empty queries
	t.Run("EmptyQueries", func(t *testing.T) {
		request := DsQueryRequest{
			Queries: []DsQuery{},
		}

		response := DatasourceQuery(context.Background(), request)

		if response.Results == nil {
			t.Error("Results should be initialized")
		}

		if len(response.Results) != 0 {
			t.Errorf("Expected empty results, got %d items", len(response.Results))
		}
	})

	// Test with single query
	t.Run("SingleQuery", func(t *testing.T) {
		request := DsQueryRequest{
			Queries: []DsQuery{
				{
					ID:             "query1",
					DatasourceName: "test-datasource",
					SQL:            "SELECT 1",
				},
			},
		}

		response := DatasourceQuery(context.Background(), request)

		if response.Results == nil {
			t.Error("Results should be initialized")
		}

		// Should contain at least the query ID (even if it fails)
		if len(response.Results) == 0 {
			t.Error("Results should contain at least one entry")
		}
	})

	// Test with multiple queries
	t.Run("MultipleQueries", func(t *testing.T) {
		request := DsQueryRequest{
			Queries: []DsQuery{
				{
					ID:             "query1",
					DatasourceName: "datasource1",
					SQL:            "SELECT 1",
				},
				{
					ID:             "query2",
					DatasourceName: "datasource2",
					SQL:            "SELECT 2",
				},
			},
		}

		response := DatasourceQuery(context.Background(), request)

		if response.Results == nil {
			t.Error("Results should be initialized")
		}

		// Should process all queries
		if len(response.Results) < 2 {
			t.Errorf("Expected at least 2 results, got %d", len(response.Results))
		}
	})

	// Test with queries for the same datasource
	t.Run("SameDatasourceQueries", func(t *testing.T) {
		request := DsQueryRequest{
			Queries: []DsQuery{
				{
					ID:             "query1",
					DatasourceName: "same-datasource",
					SQL:            "SELECT 1",
				},
				{
					ID:             "query2",
					DatasourceName: "same-datasource",
					SQL:            "SELECT 2",
				},
			},
		}

		response := DatasourceQuery(context.Background(), request)

		if response.Results == nil {
			t.Error("Results should be initialized")
		}

		// Should process all queries even for the same datasource
		if len(response.Results) < 2 {
			t.Errorf("Expected at least 2 results, got %d", len(response.Results))
		}
	})

	// Test response structure
	t.Run("ResponseStructure", func(t *testing.T) {
		request := DsQueryRequest{
			Queries: []DsQuery{
				{
					ID:             "test-query",
					DatasourceName: "test-datasource",
					SQL:            "SELECT 1",
				},
			},
		}

		response := DatasourceQuery(context.Background(), request)

		// Verify response structure
		if response.Results == nil {
			t.Error("Results map should be initialized")
		}

		// Check that Results is of correct type
		var _ map[string]model.QueryDataResult = response.Results
	})
}

// TestDsQueryResponse_Types tests the types used in DsQueryResponse
func TestDsQueryResponse_Types(t *testing.T) {
	response := &DsQueryResponse{
		Results: map[string]model.QueryDataResult{
			"test": {
				Error: "test error",
			},
		},
	}

	// Test that we can access results
	result, exists := response.Results["test"]
	if !exists {
		t.Error("Expected to find test result")
	}

	if result.Error != "test error" {
		t.Errorf("Expected 'test error', got %s", result.Error)
	}
}

// TestDatasourceQuery_Integration tests integration aspects
func TestDatasourceQuery_Integration(t *testing.T) {
	// Test that DatasourceQuery calls the set method properly
	request := DsQueryRequest{
		Queries: []DsQuery{
			{
				ID:             "integration-test",
				DatasourceName: "nonexistent-datasource",
				SQL:            "SELECT 1",
			},
		},
	}

	// This should not panic and should return a response
	response := DatasourceQuery(context.Background(), request)

	if response.Results == nil {
		t.Error("Results should be initialized even for failed queries")
	}

	// Check if the query was processed (should have an entry even if it failed)
	if len(response.Results) == 0 {
		t.Error("Should have at least one result entry")
	}
}

// TestDsQueryResponse_ConcurrentAccess tests concurrent access to response
func TestDsQueryResponse_ConcurrentAccess(t *testing.T) {
	response := &DsQueryResponse{
		Results: make(map[string]model.QueryDataResult),
	}

	// Initialize map with some test data
	response.Results["test"] = model.QueryDataResult{
		Error: "test error",
	}

	// Multiple goroutines should be able to read from response safely
	done := make(chan bool, 2)

	go func() {
		defer func() { done <- true }()
		for range 100 {
			_ = len(response.Results)
		}
	}()

	go func() {
		defer func() { done <- true }()
		for range 100 {
			_, exists := response.Results["test"]
			if !exists {
				// This is expected to always exist since we initialized it
				t.Errorf("Expected to find test result")
				return
			}
		}
	}()

	<-done
	<-done

	t.Log("Concurrent access test completed")
}

// Benchmark tests
func BenchmarkDatasourceQuery_SingleQuery(b *testing.B) {
	request := DsQueryRequest{
		Queries: []DsQuery{
			{
				ID:             "benchmark-query",
				DatasourceName: "benchmark-datasource",
				SQL:            "SELECT 1",
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DatasourceQuery(context.Background(), request)
	}
}

func BenchmarkDatasourceQuery_MultipleQueries(b *testing.B) {
	request := DsQueryRequest{
		Queries: []DsQuery{
			{ID: "q1", DatasourceName: "ds1", SQL: "SELECT 1"},
			{ID: "q2", DatasourceName: "ds2", SQL: "SELECT 2"},
			{ID: "q3", DatasourceName: "ds3", SQL: "SELECT 3"},
			{ID: "q4", DatasourceName: "ds4", SQL: "SELECT 4"},
			{ID: "q5", DatasourceName: "ds5", SQL: "SELECT 5"},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = DatasourceQuery(context.Background(), request)
	}
}

func BenchmarkDsQueryResponse_Creation(b *testing.B) {
	for i := 0; i < b.N; i++ {
		response := &DsQueryResponse{
			Results: map[string]model.QueryDataResult{},
		}
		// Access a field to prevent compiler optimization
		if response.Results == nil {
			b.Fatal("Results should not be nil")
		}
	}
}

// Test Api struct methods for race conditions with separate instances
func TestApi_ConcurrentAccess(t *testing.T) {
	config := common.Config{}

	var wg sync.WaitGroup
	numGoroutines := 50

	// Create separate API instances for each goroutine to avoid race conditions
	apis := make([]*Api, numGoroutines)
	for i := range numGoroutines {
		apis[i] = &Api{}
	}

	wg.Add(numGoroutines)
	for i := range numGoroutines {
		go func(index int, api *Api) {
			defer wg.Done()
			api.Init("/test/path", config)
			api.Use()
			api.GetRelativePath()

			defer func() {
				if r := recover(); r != nil {
					// Expected to potentially fail due to plugin setup
				}
			}()

			api.Unload()
		}(i, apis[i])
	}
	wg.Wait()
}

func TestApi_Methods(t *testing.T) {
	api := &Api{}
	configPath := "/test/config/path"
	config := common.Config{}

	api.Init(configPath, config)

	if api.configPath != configPath {
		t.Errorf("expected configPath %s, got %s", configPath, api.configPath)
	}

	if !api.Use() {
		t.Error("expected Use() to return true for master type")
	}

	expected := "/plugins"
	result := api.GetRelativePath()
	if result != expected {
		t.Errorf("expected %s, got %s", expected, result)
	}

	gin.SetMode(gin.TestMode)
	router := gin.New()

	defer func() {
		if r := recover(); r != nil {
			t.Errorf("RegisterRoutes panicked: %v", r)
		}
	}()

	api.RegisterRoutes(router)
}
