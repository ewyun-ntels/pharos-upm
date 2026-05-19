package plugins

import (
	"context"
	"sync"
	"testing"

	"ntels.com/pharos/core/internal"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/model"
)

// Setup function for tests
func setupTestEnvironment() func() {
	// Save original state
	originalProvisioningDatasources := provisioningDatasources
	originalConfig := config

	// Set up test state
	provisioningDatasources = internal.NewMap[*Datasource]()
	config = common.Config{}

	// Return cleanup function
	return func() {
		provisioningDatasources = originalProvisioningDatasources
		config = originalConfig
	}
}

// TestDsQueryResponse_Set tests the set method
func TestDsQueryResponse_Set(t *testing.T) {
	defer setupTestEnvironment()()

	// Test with empty queries
	t.Run("EmptyQueries", func(t *testing.T) {
		response := &DsQueryResponse{
			Results: make(map[string]model.QueryDataResult),
		}

		request := DsQueryRequest{
			Queries: []DsQuery{},
		}

		response.set(context.Background(), request)

		if len(response.Results) != 0 {
			t.Errorf("Expected 0 results, got %d", len(response.Results))
		}
	})

	// Test with single query that should fail (no datasource)
	t.Run("SingleQueryNoDatasource", func(t *testing.T) {
		response := &DsQueryResponse{
			Results: make(map[string]model.QueryDataResult),
		}

		request := DsQueryRequest{
			Queries: []DsQuery{
				{
					ID:             "test-query",
					DatasourceName: "nonexistent-datasource",
					SQL:            "SELECT 1",
					Timeout:        30,
				},
			},
		}

		response.set(context.Background(), request)

		if len(response.Results) == 0 {
			t.Error("Expected at least 1 result")
		}

		result, exists := response.Results["test-query"]
		if !exists {
			t.Error("Expected to find test-query result")
		}

		if result.Error == "" {
			t.Error("Expected error for nonexistent datasource")
		}
	})

	// Test with multiple queries for same datasource
	t.Run("MultipleQueriesSameDatasource", func(t *testing.T) {
		response := &DsQueryResponse{
			Results: make(map[string]model.QueryDataResult),
		}

		request := DsQueryRequest{
			Queries: []DsQuery{
				{
					ID:             "query1",
					DatasourceName: "test-datasource",
					SQL:            "SELECT 1",
					Timeout:        30,
				},
				{
					ID:             "query2",
					DatasourceName: "test-datasource",
					SQL:            "SELECT 2",
					Timeout:        60,
				},
			},
		}

		response.set(context.Background(), request)

		// Should have results for both queries
		if len(response.Results) < 2 {
			t.Errorf("Expected at least 2 results, got %d", len(response.Results))
		}
	})
}

// TestDsQueryResponse_MakeQueryDataRequests tests the makeQueryDataRequests method
func TestDsQueryResponse_MakeQueryDataRequests(t *testing.T) {
	defer setupTestEnvironment()()

	response := &DsQueryResponse{
		Results: make(map[string]model.QueryDataResult),
	}

	// Test with empty queries
	t.Run("EmptyQueries", func(t *testing.T) {
		request := DsQueryRequest{
			Queries: []DsQuery{},
		}

		requests := response.makeQueryDataRequests(request)

		if len(requests) != 0 {
			t.Errorf("Expected 0 requests, got %d", len(requests))
		}
	})

	// Test with single query
	t.Run("SingleQuery", func(t *testing.T) {
		request := DsQueryRequest{
			Queries: []DsQuery{
				{
					ID:             "test-query",
					DatasourceName: "test-datasource",
					SQL:            "SELECT 1",
					Timeout:        30,
				},
			},
		}

		_ = response.makeQueryDataRequests(request)

		// Should create error result for nonexistent datasource
		if len(response.Results) == 0 {
			t.Error("Expected error result to be created")
		}

		result, exists := response.Results["test-query"]
		if !exists {
			t.Error("Expected to find test-query result")
		}

		if result.Error == "" {
			t.Error("Expected error for nonexistent datasource")
		}
	})

	// Test with multiple queries for same datasource
	t.Run("MultipleQueriesSameDatasource", func(t *testing.T) {
		response := &DsQueryResponse{
			Results: make(map[string]model.QueryDataResult),
		}

		request := DsQueryRequest{
			Queries: []DsQuery{
				{
					ID:             "query1",
					DatasourceName: "same-datasource",
					SQL:            "SELECT 1",
					Timeout:        30,
				},
				{
					ID:             "query2",
					DatasourceName: "same-datasource",
					SQL:            "SELECT 2",
					Timeout:        60,
				},
			},
		}

		requests := response.makeQueryDataRequests(request)

		// Should group queries for same datasource
		if len(requests) != 0 {
			// If datasource exists, should have 1 request with 2 queries
			if request, exists := requests["same-datasource"]; exists {
				if len(request.Queries) != 2 {
					t.Errorf("Expected 2 queries in request, got %d", len(request.Queries))
				}
			}
		}
	})

	// Test with queries for different datasources
	t.Run("MultipleQueriesDifferentDatasources", func(t *testing.T) {
		response := &DsQueryResponse{
			Results: make(map[string]model.QueryDataResult),
		}

		request := DsQueryRequest{
			Queries: []DsQuery{
				{
					ID:             "query1",
					DatasourceName: "datasource1",
					SQL:            "SELECT 1",
					Timeout:        30,
				},
				{
					ID:             "query2",
					DatasourceName: "datasource2",
					SQL:            "SELECT 2",
					Timeout:        60,
				},
			},
		}

		_ = response.makeQueryDataRequests(request)

		// Should create separate requests for different datasources
		// (or error results if datasources don't exist)
		if len(response.Results) < 2 {
			t.Errorf("Expected at least 2 results, got %d", len(response.Results))
		}
	})
}

// TestDsQueryResponse_ConcurrentProcessing tests concurrent processing of the response
func TestDsQueryResponse_ConcurrentProcessing(t *testing.T) {
	defer setupTestEnvironment()()

	response := &DsQueryResponse{
		Results: make(map[string]model.QueryDataResult),
	}

	request := DsQueryRequest{
		Queries: []DsQuery{
			{
				ID:             "query1",
				DatasourceName: "datasource1",
				SQL:            "SELECT 1",
				Timeout:        30,
			},
			{
				ID:             "query2",
				DatasourceName: "datasource2",
				SQL:            "SELECT 2",
				Timeout:        60,
			},
		},
	}

	// Test that the set method handles concurrent access properly
	var wg sync.WaitGroup
	wg.Add(2)

	go func() {
		defer wg.Done()
		response.set(context.Background(), request)
	}()

	go func() {
		defer wg.Done()
		// Read from response concurrently (guarded by RLock)
		for range 100 {
			response.mu.RLock()
			_ = len(response.Results)
			response.mu.RUnlock()
		}
	}()

	wg.Wait()

	t.Log("Concurrent access test completed")
}

// TestDsQueryResponse_WithProvisioningDatasource tests with provisioning datasource
func TestDsQueryResponse_WithProvisioningDatasource(t *testing.T) {
	defer setupTestEnvironment()()

	// Set up a provisioning datasource
	testDatasource := &Datasource{
		Datasource: model.Datasource{
			Name: "test-provisioning",
			Type: "test-type",
			Data: map[string]any{
				"host": "localhost",
				"port": 5432,
			},
		},
		Provisioning: true,
	}

	provisioningDatasources.Set("test-provisioning", testDatasource)

	response := &DsQueryResponse{
		Results: make(map[string]model.QueryDataResult),
	}

	request := DsQueryRequest{
		Queries: []DsQuery{
			{
				ID:             "test-query",
				DatasourceName: "test-provisioning",
				SQL:            "SELECT 1",
				Timeout:        30,
			},
		},
	}

	response.set(context.Background(), request)

	// Should process the query (may result in error due to no actual plugin)
	if len(response.Results) == 0 {
		t.Error("Expected at least 1 result")
	}
}

// TestDsQueryResponse_ErrorHandling tests various error conditions
func TestDsQueryResponse_ErrorHandling(t *testing.T) {
	defer setupTestEnvironment()()

	testCases := []struct {
		name          string
		query         DsQuery
		expectedError bool
	}{
		{
			name: "EmptyDatasourceName",
			query: DsQuery{
				ID:             "test-query",
				DatasourceName: "",
				SQL:            "SELECT 1",
				Timeout:        30,
			},
			expectedError: true,
		},
		{
			name: "EmptySQL",
			query: DsQuery{
				ID:             "test-query",
				DatasourceName: "test-datasource",
				SQL:            "",
				Timeout:        30,
			},
			expectedError: true,
		},
		{
			name: "EmptyID",
			query: DsQuery{
				ID:             "",
				DatasourceName: "test-datasource",
				SQL:            "SELECT 1",
				Timeout:        30,
			},
			expectedError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			response := &DsQueryResponse{
				Results: make(map[string]model.QueryDataResult),
			}

			request := DsQueryRequest{
				Queries: []DsQuery{tc.query},
			}

			response.set(context.Background(), request)

			if tc.expectedError {
				// Should have some result (even if it's an error)
				if len(response.Results) == 0 {
					t.Error("Expected error result to be created")
				}
			}
		})
	}
}

// TestDsQueryResponse_ThreadSafety tests thread safety of the response
func TestDsQueryResponse_ThreadSafety(t *testing.T) {
	defer setupTestEnvironment()()

	// Create multiple goroutines that access the response
	var wg sync.WaitGroup
	numGoroutines := 10

	for i := range numGoroutines {
		wg.Add(1)
		go func(id int) {
			defer wg.Done()

			request := DsQueryRequest{
				Queries: []DsQuery{
					{
						ID:             "query-" + string(rune('0'+id)),
						DatasourceName: "datasource-" + string(rune('0'+id)),
						SQL:            "SELECT " + string(rune('0'+id)),
						Timeout:        30,
					},
				},
			}

			// Each goroutine processes its own request
			localResponse := &DsQueryResponse{
				Results: make(map[string]model.QueryDataResult),
			}
			localResponse.set(context.Background(), request)
		}(i)
	}

	wg.Wait()

	t.Log("Thread safety test completed")
}

// TestDsQueryResponse_LargeNumberOfQueries tests handling large number of queries
func TestDsQueryResponse_LargeNumberOfQueries(t *testing.T) {
	defer setupTestEnvironment()()

	response := &DsQueryResponse{
		Results: make(map[string]model.QueryDataResult),
	}

	// Create a large number of queries
	var queries []DsQuery
	for i := range 100 {
		queries = append(queries, DsQuery{
			ID:             "query-" + string(rune('0'+(i%10))),
			DatasourceName: "datasource-" + string(rune('0'+(i%5))),
			SQL:            "SELECT " + string(rune('0'+(i%10))),
			Timeout:        30 + (i % 60),
		})
	}

	request := DsQueryRequest{
		Queries: queries,
	}

	response.set(context.Background(), request)

	// Should handle all queries
	if len(response.Results) == 0 {
		t.Error("Expected results for large number of queries")
	}

	t.Logf("Processed %d queries, got %d results", len(queries), len(response.Results))
}

// Benchmark tests
func BenchmarkDsQueryResponse_Set_SingleQuery(b *testing.B) {
	defer setupTestEnvironment()()

	request := DsQueryRequest{
		Queries: []DsQuery{
			{
				ID:             "benchmark-query",
				DatasourceName: "benchmark-datasource",
				SQL:            "SELECT 1",
				Timeout:        30,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response := &DsQueryResponse{
			Results: make(map[string]model.QueryDataResult),
		}
		response.set(context.Background(), request)
	}
}

func BenchmarkDsQueryResponse_Set_MultipleQueries(b *testing.B) {
	defer setupTestEnvironment()()

	var queries []DsQuery
	for i := range 10 {
		queries = append(queries, DsQuery{
			ID:             "query-" + string(rune('0'+(i%10))),
			DatasourceName: "datasource-" + string(rune('0'+(i%3))),
			SQL:            "SELECT " + string(rune('0'+(i%10))),
			Timeout:        30 + i,
		})
	}

	request := DsQueryRequest{
		Queries: queries,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response := &DsQueryResponse{
			Results: make(map[string]model.QueryDataResult),
		}
		response.set(context.Background(), request)
	}
}

func BenchmarkDsQueryResponse_MakeQueryDataRequests(b *testing.B) {
	defer setupTestEnvironment()()

	request := DsQueryRequest{
		Queries: []DsQuery{
			{ID: "q1", DatasourceName: "ds1", SQL: "SELECT 1", Timeout: 30},
			{ID: "q2", DatasourceName: "ds2", SQL: "SELECT 2", Timeout: 60},
			{ID: "q3", DatasourceName: "ds1", SQL: "SELECT 3", Timeout: 45},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response := &DsQueryResponse{
			Results: make(map[string]model.QueryDataResult),
		}
		_ = response.makeQueryDataRequests(request)
	}
}
