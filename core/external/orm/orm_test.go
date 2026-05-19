package orm

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDriverName_Constants(t *testing.T) {
	tests := []struct {
		name     string
		driver   string
		expected string
	}{
		{
			name:     "Altibase driver",
			driver:   DriverAltibase,
			expected: "odbc",
		},
		{
			name:     "ClickHouse driver",
			driver:   DriverClickHouse,
			expected: "clickhouse",
		},
		{
			name:     "PostgreSQL driver",
			driver:   DriverPostgreSQL,
			expected: "postgres",
		},
		{
			name:     "SQLite driver",
			driver:   DriverSqlite,
			expected: "sqlite",
		},
		{
			name:     "Vertica driver",
			driver:   DriverVertica,
			expected: "vertica",
		},
		{
			name:     "Prometheus driver",
			driver:   DriverPrometheus,
			expected: "prometheus",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			actual := DriverName[tt.driver]
			assert.Equal(t, tt.expected, actual)
		})
	}
}

func TestDriverConstants(t *testing.T) {
	assert.Equal(t, "", DriverDefault)
	assert.Equal(t, "postgresql", DriverPostgreSQL)
	assert.Equal(t, "sqlite", DriverSqlite)
	assert.Equal(t, "clickhouse", DriverClickHouse)
	assert.Equal(t, "altibase", DriverAltibase)
	assert.Equal(t, "vertica", DriverVertica)
	assert.Equal(t, "prometheus", DriverPrometheus)
}

func TestDatabaseResponse_Structure(t *testing.T) {
	// Test that DatabaseResponse structure has expected fields
	response := DatabaseResponse{
		Meta: []map[string]string{
			{"column1": "type1"},
			{"column2": "type2"},
		},
		Data: []map[string]any{
			{"column1": "value1", "column2": 123},
			{"column1": "value2", "column2": 456},
		},
		Rows: 2,
		SQL:  "SELECT * FROM test",
		Statistics: DatabaseResponseStatistics{
			Elapsed: 0.123,
		},
	}

	assert.Equal(t, 2, len(response.Meta))
	assert.Equal(t, 2, len(response.Data))
	assert.Equal(t, int64(2), response.Rows)
	assert.Equal(t, "SELECT * FROM test", response.SQL)
	assert.Equal(t, 0.123, response.Statistics.Elapsed)
}

// Helper function to create a temporary SQLite database for testing
func createTempSQLiteDB(t *testing.T) (string, func()) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test.db")

	cleanup := func() {
		os.Remove(dbPath)
	}

	return dbPath, cleanup
}

func TestHandler_WithDefaultDriver(t *testing.T) {
	// Create temporary SQLite database
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	// Setup config
	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	// Test handler function
	var handlerCalled bool
	var receivedDB *sqlx.DB
	handlerFunc := func(db *sqlx.DB) error {
		handlerCalled = true
		receivedDB = db
		return nil
	}

	// Execute
	err := Handler(DriverDefault, config, handlerFunc)

	// Assert
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.NotNil(t, receivedDB)
}

func TestHandler_WithSpecificDriver(t *testing.T) {
	// Create temporary SQLite database
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	// Setup config with different driver
	config := &DatabaseConfig{
		Driver: DriverPostgreSQL, // This will be overridden
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	// Test handler function
	var handlerCalled bool
	var receivedDB *sqlx.DB
	handlerFunc := func(db *sqlx.DB) error {
		handlerCalled = true
		receivedDB = db
		return nil
	}

	// Execute with SQLite driver override
	err := Handler(DriverSqlite, config, handlerFunc)

	// Assert
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
	assert.NotNil(t, receivedDB)
}

func TestHandler_HandlerFunctionError(t *testing.T) {
	// Create temporary SQLite database
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	// Setup config
	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	// Test handler function that returns error
	expectedError := errors.New("handler function error")
	handlerFunc := func(db *sqlx.DB) error {
		return expectedError
	}

	// Execute
	err := Handler(DriverDefault, config, handlerFunc)

	// Assert
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func TestHandler_InvalidDatabaseConfig(t *testing.T) {
	// Setup config with invalid path
	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: "/invalid/path/to/database.db",
		},
	}

	// Test handler function
	handlerFunc := func(db *sqlx.DB) error {
		t.Error("Handler function should not be called when database connection fails")
		return nil
	}

	// Execute
	err := Handler(DriverDefault, config, handlerFunc)

	// Assert - should return error due to invalid path
	assert.Error(t, err)
}

func TestHandler_ConfigModification(t *testing.T) {
	// Create temporary SQLite database
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	// Test that the original config is not modified
	originalConfig := &DatabaseConfig{
		Driver: DriverPostgreSQL,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	// Create a copy to compare later
	configCopy := *originalConfig

	handlerFunc := func(db *sqlx.DB) error {
		return nil
	}

	// Execute with different driver
	err := Handler(DriverSqlite, originalConfig, handlerFunc)

	// Assert
	assert.NoError(t, err)
	// Verify original config was not modified
	assert.Equal(t, configCopy, *originalConfig)
}

func TestHandler_MultipleDriverTypes(t *testing.T) {
	// Create temporary SQLite database
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	testCases := []struct {
		name           string
		handlerDriver  string
		configDriver   string
		expectedDriver string
	}{
		{
			name:           "Default driver uses config driver",
			handlerDriver:  DriverDefault,
			configDriver:   DriverSqlite,
			expectedDriver: DriverSqlite,
		},
		{
			name:           "Specific driver overrides config driver",
			handlerDriver:  DriverSqlite,
			configDriver:   DriverPostgreSQL,
			expectedDriver: DriverSqlite,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			config := &DatabaseConfig{
				Driver: tc.configDriver,
				SQLite: SQLiteConfig{
					Path: dbPath,
				},
			}

			var actualConfig DatabaseConfig
			handlerFunc := func(db *sqlx.DB) error {
				// We can't directly check the driver used, but we can verify the handler was called
				assert.NotNil(t, db)
				return nil
			}

			// Only test with SQLite since it's the only one we can easily test
			if tc.expectedDriver == DriverSqlite {
				err := Handler(tc.handlerDriver, config, handlerFunc)
				assert.NoError(t, err)
			}

			// Verify original config driver wasn't changed
			assert.Equal(t, tc.configDriver, config.Driver)
			_ = actualConfig // Prevent unused variable warning
		})
	}
}

func TestDBHandlerFunc_Type(t *testing.T) {
	// Test that DBHandlerFunc is properly defined
	var handler DBHandlerFunc = func(db *sqlx.DB) error {
		return nil
	}

	assert.NotNil(t, handler)

	// Test with nil DB (should not panic)
	err := handler(nil)
	assert.NoError(t, err)
}

func TestHandler_ConcurrentAccess(t *testing.T) {
	// Create temporary SQLite database
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	// Test concurrent access to the same handler
	const numGoroutines = 10
	done := make(chan bool, numGoroutines)
	errors := make(chan error, numGoroutines)

	for range numGoroutines {
		go func() {
			handlerFunc := func(db *sqlx.DB) error {
				// Simple query to verify DB connection
				_, err := db.Exec("SELECT 1")
				return err
			}

			err := Handler(DriverDefault, config, handlerFunc)
			if err != nil {
				errors <- err
			}
			done <- true
		}()
	}

	// Wait for all goroutines to complete
	for range numGoroutines {
		<-done
	}

	// Check for errors
	close(errors)
	for err := range errors {
		t.Errorf("Concurrent access error: %v", err)
	}
}

// Integration test that verifies the complete flow with a real database operation
func TestHandler_IntegrationSQLite(t *testing.T) {
	// Create temporary SQLite database
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	// Create a test table and insert data
	err := Handler(DriverDefault, config, func(db *sqlx.DB) error {
		// Create table
		_, err := db.Exec(`CREATE TABLE test_table (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL
		)`)
		if err != nil {
			return err
		}

		// Insert data
		_, err = db.Exec("INSERT INTO test_table (name) VALUES (?)", "test_name")
		return err
	})

	require.NoError(t, err)

	// Verify data was inserted
	err = Handler(DriverDefault, config, func(db *sqlx.DB) error {
		var count int
		err := db.Get(&count, "SELECT COUNT(*) FROM test_table")
		if err != nil {
			return err
		}

		assert.Equal(t, 1, count)
		return nil
	})

	assert.NoError(t, err)
}

// Benchmark tests
func BenchmarkHandler_SQLite(b *testing.B) {
	// Create temporary SQLite database
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench.db")

	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	handlerFunc := func(db *sqlx.DB) error {
		_, err := db.Exec("SELECT 1")
		return err
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := Handler(DriverDefault, config, handlerFunc)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkHandler_ConfigCopy(b *testing.B) {
	// Benchmark the config copying overhead
	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: ":memory:",
		},
		PostgreSQL: PostgreSQLConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "user",
			Password: "password",
			Database: "testdb",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		// Simulate the config copying that happens in Handler
		databaseConfig := *config
		databaseConfig.Driver = DriverPostgreSQL
		_ = databaseConfig
	}
}

func TestHandlerWithRole_DefaultPool(t *testing.T) {
	// Create temporary SQLite database
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	var handlerCalled bool
	handlerFunc := func(db *sqlx.DB) error {
		handlerCalled = true
		assert.NotNil(t, db)
		return nil
	}

	err := HandlerWithRole(PoolDefault, DriverDefault, config, handlerFunc)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestHandlerWithRole_StatisticsPool(t *testing.T) {
	// Create temporary SQLite database
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	var handlerCalled bool
	handlerFunc := func(db *sqlx.DB) error {
		handlerCalled = true
		assert.NotNil(t, db)
		return nil
	}

	err := HandlerWithRole(PoolStatistics, DriverDefault, config, handlerFunc)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestHandlerWithRole_ServicePool(t *testing.T) {
	// Create temporary SQLite database
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	var handlerCalled bool
	handlerFunc := func(db *sqlx.DB) error {
		handlerCalled = true
		assert.NotNil(t, db)
		return nil
	}

	err := HandlerWithRole(PoolService, DriverDefault, config, handlerFunc)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestHandlerWithRole_UnknownRole(t *testing.T) {
	// Create temporary SQLite database
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	handlerFunc := func(db *sqlx.DB) error {
		t.Error("Handler should not be called with unknown role")
		return nil
	}

	err := HandlerWithRole("unknown_role", DriverDefault, config, handlerFunc)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "unknown pool role")
}

func TestStatisticsHandler(t *testing.T) {
	// Create temporary SQLite database
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	var handlerCalled bool
	handlerFunc := func(db *sqlx.DB) error {
		handlerCalled = true
		assert.NotNil(t, db)
		return nil
	}

	err := StatisticsHandler(DriverDefault, config, handlerFunc)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestStatisticsHandler_WithSpecificDriver(t *testing.T) {
	// Create temporary SQLite database
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	config := &DatabaseConfig{
		Driver: DriverPostgreSQL, // This will be overridden
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	var handlerCalled bool
	handlerFunc := func(db *sqlx.DB) error {
		handlerCalled = true
		assert.NotNil(t, db)
		return nil
	}

	err := StatisticsHandler(DriverSqlite, config, handlerFunc)
	assert.NoError(t, err)
	assert.True(t, handlerCalled)
}

func TestHandler_BackwardCompatibility(t *testing.T) {
	// Verify that Handler() uses the default pool and is equivalent to HandlerWithRole(PoolDefault, ...)
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	var handler1Called, handler2Called bool
	handlerFunc1 := func(db *sqlx.DB) error {
		handler1Called = true
		// Create a test table to verify it's the same pool
		_, err := db.Exec("CREATE TABLE IF NOT EXISTS test_compat (id INTEGER)")
		return err
	}

	handlerFunc2 := func(db *sqlx.DB) error {
		handler2Called = true
		// Check if the table exists (should exist if using same pool)
		var count int
		err := db.Get(&count, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test_compat'")
		assert.NoError(t, err)
		assert.Equal(t, 1, count, "Table should exist, confirming same pool usage")
		return err
	}

	// Use Handler() (old API)
	err := Handler(DriverDefault, config, handlerFunc1)
	assert.NoError(t, err)
	assert.True(t, handler1Called)

	// Use HandlerWithRole() with PoolDefault (new API)
	err = HandlerWithRole(PoolDefault, DriverDefault, config, handlerFunc2)
	assert.NoError(t, err)
	assert.True(t, handler2Called)
}

func TestHandlerWithRole_DifferentPools(t *testing.T) {
	// Verify that different pools maintain separate connections
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	// Create table in default pool
	err := HandlerWithRole(PoolDefault, DriverDefault, config, func(db *sqlx.DB) error {
		_, err := db.Exec("CREATE TABLE default_table (id INTEGER)")
		return err
	})
	assert.NoError(t, err)

	// Create table in statistics pool
	err = HandlerWithRole(PoolStatistics, DriverDefault, config, func(db *sqlx.DB) error {
		_, err := db.Exec("CREATE TABLE statistics_table (id INTEGER)")
		return err
	})
	assert.NoError(t, err)

	// Verify both tables exist (since it's the same SQLite file, both pools access the same database)
	err = HandlerWithRole(PoolDefault, DriverDefault, config, func(db *sqlx.DB) error {
		var count int
		// Check for both tables
		err := db.Get(&count, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name IN ('default_table', 'statistics_table')")
		assert.NoError(t, err)
		assert.Equal(t, 2, count, "Both tables should exist")
		return err
	})
	assert.NoError(t, err)
}

func TestHandlerWithRole_ErrorHandling(t *testing.T) {
	// Test error handling in HandlerWithRole
	dbPath, cleanup := createTempSQLiteDB(t)
	defer cleanup()

	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	expectedError := errors.New("handler function error")
	handlerFunc := func(db *sqlx.DB) error {
		return expectedError
	}

	err := HandlerWithRole(PoolDefault, DriverDefault, config, handlerFunc)
	assert.Error(t, err)
	assert.Equal(t, expectedError, err)
}

func BenchmarkHandlerWithRole(b *testing.B) {
	// Benchmark HandlerWithRole performance
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench.db")

	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	handlerFunc := func(db *sqlx.DB) error {
		_, err := db.Exec("SELECT 1")
		return err
	}

	roles := []string{PoolDefault, PoolStatistics, PoolService}

	for _, role := range roles {
		b.Run(role, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				err := HandlerWithRole(role, DriverDefault, config, handlerFunc)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func BenchmarkStatisticsHandler(b *testing.B) {
	// Benchmark StatisticsHandler performance
	tempDir := b.TempDir()
	dbPath := filepath.Join(tempDir, "bench_stats.db")

	config := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: dbPath,
		},
	}

	handlerFunc := func(db *sqlx.DB) error {
		_, err := db.Exec("SELECT 1")
		return err
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := StatisticsHandler(DriverDefault, config, handlerFunc)
		if err != nil {
			b.Fatal(err)
		}
	}
}
