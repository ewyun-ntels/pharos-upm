package sql

import (
	"context"
	"fmt"
	"testing"

	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/plugins/model"
)

func TestDataClient_QueryData_Basic(t *testing.T) {
	config := orm.DatabaseConfig{
		Driver: orm.DriverSqlite,
		SQLite: orm.SQLiteConfig{
			Path: ":memory:",
		},
	}

	client := &DataClient{
		Config: common.Config{},
	}

	request := &model.QueryDataRequest{
		DatabaseConfig: config,
		Queries: []model.Query{
			{
				ID:      "A",
				SQL:     "SELECT 1 as test_column",
				Timeout: 30,
			},
		},
	}

	response := client.QueryData(context.Background(), request)

	if len(response.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(response.Results))
	}

	result, exists := response.Results["A"]
	if !exists {
		t.Errorf("Expected result for Query ID A not found")
		return
	}

	if result.Error != "" {
		t.Errorf("Unexpected error: %s", result.Error)
	}
}

func TestDataClient_QueryData_Multiple(t *testing.T) {
	config := orm.DatabaseConfig{
		Driver: orm.DriverSqlite,
		SQLite: orm.SQLiteConfig{
			Path: ":memory:",
		},
	}

	client := &DataClient{
		Config: common.Config{},
	}

	request := &model.QueryDataRequest{
		DatabaseConfig: config,
		Queries: []model.Query{
			{
				ID:      "A",
				SQL:     "SELECT 1 as id",
				Timeout: 30,
			},
			{
				ID:      "B",
				SQL:     "SELECT 'test' as name",
				Timeout: 30,
			},
		},
	}

	response := client.QueryData(context.Background(), request)

	if len(response.Results) != 2 {
		t.Errorf("Expected 2 results, got %d", len(response.Results))
	}

	// Check first query result
	resultA, exists := response.Results["A"]
	if !exists {
		t.Errorf("Expected result for Query ID A not found")
	} else if resultA.Error != "" {
		t.Errorf("Unexpected error for query A: %s", resultA.Error)
	}

	// Check second query result
	resultB, exists := response.Results["B"]
	if !exists {
		t.Errorf("Expected result for Query ID B not found")
	} else if resultB.Error != "" {
		t.Errorf("Unexpected error for query B: %s", resultB.Error)
	}
}

func TestDataClient_QueryData_Empty(t *testing.T) {
	config := orm.DatabaseConfig{
		Driver: orm.DriverSqlite,
		SQLite: orm.SQLiteConfig{
			Path: ":memory:",
		},
	}

	client := &DataClient{
		Config: common.Config{},
	}

	request := &model.QueryDataRequest{
		DatabaseConfig: config,
		Queries: []model.Query{
			{
				ID:      "A",
				SQL:     "",
				Timeout: 30,
			},
		},
	}

	response := client.QueryData(context.Background(), request)

	if len(response.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(response.Results))
	}

	result, exists := response.Results["A"]
	if !exists {
		t.Errorf("Expected result for Query ID A not found")
		return
	}

	// Empty query should cause an error
	if result.Error == "" {
		t.Errorf("Expected error for empty query but got none")
	}
}

func TestDataClient_QueryData_Invalid(t *testing.T) {
	config := orm.DatabaseConfig{
		Driver: orm.DriverSqlite,
		SQLite: orm.SQLiteConfig{
			Path: ":memory:",
		},
	}

	client := &DataClient{
		Config: common.Config{},
	}

	request := &model.QueryDataRequest{
		DatabaseConfig: config,
		Queries: []model.Query{
			{
				ID:      "A",
				SQL:     "INVALID SQL STATEMENT",
				Timeout: 30,
			},
		},
	}

	response := client.QueryData(context.Background(), request)

	if len(response.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(response.Results))
	}

	result, exists := response.Results["A"]
	if !exists {
		t.Errorf("Expected result for Query ID A not found")
		return
	}

	// Invalid SQL should cause an error
	if result.Error == "" {
		t.Errorf("Expected error for invalid SQL but got none")
	}
}

func TestDataClient_RemoveDatasource_Basic(t *testing.T) {
	config := orm.DatabaseConfig{
		Driver: orm.DriverSqlite,
		SQLite: orm.SQLiteConfig{
			Path: ":memory:",
		},
	}

	client := &DataClient{
		Config: common.Config{},
	}

	request := &model.RemoveDatasourceRequest{
		DatabaseConfig: config,
	}

	err := client.RemoveDatasource(request)
	if err != nil {
		t.Errorf("RemoveDatasource failed: %v", err)
	}
}

func TestDataClient_Integration_Flow(t *testing.T) {
	config := orm.DatabaseConfig{
		Driver: orm.DriverSqlite,
		SQLite: orm.SQLiteConfig{
			Path: ":memory:",
		},
	}

	client := &DataClient{
		Config: common.Config{},
	}

	// Test QueryData with a real query
	request := &model.QueryDataRequest{
		DatabaseConfig: config,
		Queries: []model.Query{
			{
				ID:      "A",
				SQL:     "SELECT 42 as answer, 'hello' as greeting",
				Timeout: 30,
			},
		},
	}

	response := client.QueryData(context.Background(), request)
	if len(response.Results) != 1 {
		t.Errorf("Expected 1 result, got %d", len(response.Results))
	}

	result, exists := response.Results["A"]
	if !exists {
		t.Errorf("Expected result for query A not found")
		return
	}

	if result.Error != "" {
		t.Errorf("QueryData failed: %s", result.Error)
		return
	}

	if result.Frame.Rows != 1 {
		t.Errorf("Expected 1 row, got %d", result.Frame.Rows)
	}

	if len(result.Frame.Data) != 1 {
		t.Errorf("Expected 1 data row, got %d", len(result.Frame.Data))
	}

	// Test RemoveDatasource
	removeRequest := &model.RemoveDatasourceRequest{
		DatabaseConfig: config,
	}

	err := client.RemoveDatasource(removeRequest)
	if err != nil {
		t.Errorf("RemoveDatasource failed: %v", err)
	}
}

// TestDataClient_QueryData_EdgeCases tests edge cases for QueryData
func TestDataClient_QueryData_EdgeCases(t *testing.T) {
	config := orm.DatabaseConfig{
		Driver: orm.DriverSqlite,
		SQLite: orm.SQLiteConfig{
			Path: ":memory:",
		},
	}

	client := &DataClient{
		Config: common.Config{},
	}

	t.Run("multiple queries with different types", func(t *testing.T) {
		request := &model.QueryDataRequest{
			DatabaseConfig: config,
			Queries: []model.Query{
				{
					ID:      "numbers",
					SQL:     "SELECT 1 as one, 2 as two, 3 as three",
					Timeout: 30,
				},
				{
					ID:      "strings",
					SQL:     "SELECT 'hello' as greeting, 'world' as target",
					Timeout: 30,
				},
			},
		}

		response := client.QueryData(context.Background(), request)

		if len(response.Results) != 2 {
			t.Errorf("Expected 2 results, got %d", len(response.Results))
		}

		// Check numbers query
		numbersResult, exists := response.Results["numbers"]
		if !exists {
			t.Errorf("Expected result for numbers query not found")
		} else {
			if numbersResult.Error != "" {
				t.Errorf("Numbers query failed: %s", numbersResult.Error)
			}
			if numbersResult.Frame.Rows != 1 {
				t.Errorf("Expected 1 row for numbers, got %d", numbersResult.Frame.Rows)
			}
		}

		// Check strings query
		stringsResult, exists := response.Results["strings"]
		if !exists {
			t.Errorf("Expected result for strings query not found")
		} else {
			if stringsResult.Error != "" {
				t.Errorf("Strings query failed: %s", stringsResult.Error)
			}
			if stringsResult.Frame.Rows != 1 {
				t.Errorf("Expected 1 row for strings, got %d", stringsResult.Frame.Rows)
			}
		}
	})

	t.Run("query with no results", func(t *testing.T) {
		request := &model.QueryDataRequest{
			DatabaseConfig: config,
			Queries: []model.Query{
				{
					ID:      "empty",
					SQL:     "SELECT 1 WHERE 0 = 1", // Always false condition
					Timeout: 30,
				},
			},
		}

		response := client.QueryData(context.Background(), request)

		if len(response.Results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(response.Results))
		}

		result, exists := response.Results["empty"]
		if !exists {
			t.Errorf("Expected result for empty query not found")
		} else {
			if result.Error != "" {
				t.Errorf("Empty query failed: %s", result.Error)
			}
			if result.Frame.Rows != 0 {
				t.Errorf("Expected 0 rows for empty result, got %d", result.Frame.Rows)
			}
		}
	})

	t.Run("query with special characters", func(t *testing.T) {
		request := &model.QueryDataRequest{
			DatabaseConfig: config,
			Queries: []model.Query{
				{
					ID:      "special",
					SQL:     "SELECT 'test''s value' as quoted, 'line1\nline2' as multiline",
					Timeout: 30,
				},
			},
		}

		response := client.QueryData(context.Background(), request)

		if len(response.Results) != 1 {
			t.Errorf("Expected 1 result, got %d", len(response.Results))
		}

		result, exists := response.Results["special"]
		if !exists {
			t.Errorf("Expected result for special query not found")
		} else {
			if result.Error != "" {
				t.Errorf("Special characters query failed: %s", result.Error)
			}
		}
	})
}

// TestDataClient_RemoveDatasource_EdgeCases tests edge cases for RemoveDatasource
func TestDataClient_RemoveDatasource_EdgeCases(t *testing.T) {
	client := &DataClient{
		Config: common.Config{},
	}

	t.Run("remove with different database types", func(t *testing.T) {
		configs := []orm.DatabaseConfig{
			{
				Driver: orm.DriverSqlite,
				SQLite: orm.SQLiteConfig{Path: ":memory:"},
			},
			{
				Driver: orm.DriverPostgreSQL,
				PostgreSQL: orm.PostgreSQLConfig{
					Host:     "localhost",
					Port:     5432,
					Username: "test",
					Password: "test",
					Database: "test",
				},
			},
		}

		for _, config := range configs {
			t.Run(config.Driver, func(t *testing.T) {
				request := &model.RemoveDatasourceRequest{
					DatabaseConfig: config,
				}

				err := client.RemoveDatasource(request)
				// For this test, we don't expect errors even if the database doesn't exist
				// The pool should handle gracefully
				if err != nil {
					t.Logf("RemoveDatasource for %s returned error (may be expected): %v", config.Driver, err)
				} else {
					t.Logf("RemoveDatasource for %s completed successfully", config.Driver)
				}
			})
		}
	})
}

// TestDataClient_Performance tests performance characteristics
func TestDataClient_Performance(t *testing.T) {
	config := orm.DatabaseConfig{
		Driver: orm.DriverSqlite,
		SQLite: orm.SQLiteConfig{
			Path: ":memory:",
		},
	}

	client := &DataClient{
		Config: common.Config{},
	}

	t.Run("large number of queries", func(t *testing.T) {
		var queries []model.Query
		for i := range 100 {
			queries = append(queries, model.Query{
				ID:      fmt.Sprintf("query_%d", i),
				SQL:     fmt.Sprintf("SELECT %d as number", i),
				Timeout: 30,
			})
		}

		request := &model.QueryDataRequest{
			DatabaseConfig: config,
			Queries:        queries,
		}

		response := client.QueryData(context.Background(), request)

		if len(response.Results) != 100 {
			t.Errorf("Expected 100 results, got %d", len(response.Results))
		}

		// Check that all queries executed successfully
		errorCount := 0
		for _, result := range response.Results {
			if result.Error != "" {
				errorCount++
			}
		}

		if errorCount > 0 {
			t.Errorf("Expected 0 errors, got %d", errorCount)
		}
	})
}

// BenchmarkDataClient_QueryData benchmarks the QueryData operation
func BenchmarkDataClient_QueryData(b *testing.B) {
	config := orm.DatabaseConfig{
		Driver: orm.DriverSqlite,
		SQLite: orm.SQLiteConfig{
			Path: ":memory:",
		},
	}

	client := &DataClient{
		Config: common.Config{},
	}

	request := &model.QueryDataRequest{
		DatabaseConfig: config,
		Queries: []model.Query{
			{
				ID:      "benchmark",
				SQL:     "SELECT 42 as answer",
				Timeout: 30,
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		response := client.QueryData(context.Background(), request)
		if len(response.Results) != 1 {
			b.Errorf("Expected 1 result, got %d", len(response.Results))
		}
	}
}

// TestDataClient_ConfigurationsCompatibility tests different database configurations
func TestDataClient_ConfigurationsCompatibility(t *testing.T) {
	client := &DataClient{
		Config: common.Config{},
	}

	// Test with SQLite (always available)
	t.Run("SQLite configuration", func(t *testing.T) {
		config := orm.DatabaseConfig{
			Driver: orm.DriverSqlite,
			SQLite: orm.SQLiteConfig{
				Path: ":memory:",
			},
		}

		request := &model.QueryDataRequest{
			DatabaseConfig: config,
			Queries: []model.Query{
				{
					ID:      "test",
					SQL:     "SELECT sqlite_version() as version",
					Timeout: 30,
				},
			},
		}

		response := client.QueryData(context.Background(), request)
		result, exists := response.Results["test"]
		if !exists {
			t.Errorf("Expected result not found")
		} else if result.Error != "" {
			t.Errorf("SQLite query failed: %s", result.Error)
		}
	})

	// Test other configurations (may not be available in test environment)
	configurations := []struct {
		name   string
		config orm.DatabaseConfig
	}{
		{
			name: "PostgreSQL configuration",
			config: orm.DatabaseConfig{
				Driver: orm.DriverPostgreSQL,
				PostgreSQL: orm.PostgreSQLConfig{
					Host:              "localhost",
					Port:              5432,
					Username:          "test",
					Password:          "test",
					Database:          "test",
					MaxOpenConnection: 10,
					MaxLifetime:       3600,
				},
			},
		},
		{
			name: "ClickHouse configuration",
			config: orm.DatabaseConfig{
				Driver: orm.DriverClickHouse,
				ClickHouse: orm.ClickHouseConfig{
					Host:              "localhost",
					Port:              9000,
					Username:          "default",
					Password:          "",
					Database:          "default",
					MaxOpenConnection: 10,
					MaxLifetime:       3600,
				},
			},
		},
	}

	for _, tc := range configurations {
		t.Run(tc.name, func(t *testing.T) {
			request := &model.QueryDataRequest{
				DatabaseConfig: tc.config,
				Queries: []model.Query{
					{
						ID:      "connectivity_test",
						SQL:     "SELECT 1 as test",
						Timeout: 30,
					},
				},
			}

			response := client.QueryData(context.Background(), request)
			result, exists := response.Results["connectivity_test"]
			if !exists {
				t.Errorf("Expected result not found")
			} else {
				// For these tests, we just check that the client doesn't panic
				// Connection errors are expected if the databases aren't available
				t.Logf("%s result: error=%s", tc.name, result.Error)
			}
		})
	}
}
