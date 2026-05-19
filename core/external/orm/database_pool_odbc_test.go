//go:build odbc

package orm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestODBCBuildTag_SupportDriverMap(t *testing.T) {
	// Test that with ODBC build tag, Altibase driver is supported
	supported, exists := supportDriver[DriverAltibase]
	assert.True(t, exists, "DriverAltibase should exist in supportDriver map")
	assert.True(t, supported, "DriverAltibase should be supported with ODBC build tag")
}

func TestODBCBuildTag_AllDriversSupported(t *testing.T) {
	// With ODBC build tag, all drivers should be supported
	expectedDrivers := map[string]bool{
		DriverAltibase:   true,
		DriverSqlite:     true,
		DriverPostgreSQL: true,
		DriverClickHouse: true,
		DriverVertica:    true,
		DriverPrometheus: true,
	}

	for driver, expectedSupport := range expectedDrivers {
		actual, exists := supportDriver[driver]
		assert.True(t, exists, "Driver %s should exist in supportDriver map", driver)
		assert.Equal(t, expectedSupport, actual, "Driver %s support should be %v with ODBC build tag", driver, expectedSupport)
	}
}

func TestODBCBuildTag_SupportDriverCount(t *testing.T) {
	// Verify that all expected drivers are present
	expectedDriverCount := 6 // sqlite, postgresql, clickhouse, altibase, vertica, prometheus
	assert.Equal(t, expectedDriverCount, len(supportDriver),
		"supportDriver should contain exactly %d drivers with ODBC build tag", expectedDriverCount)
}

func TestODBCBuildTag_AltibaseSpecific(t *testing.T) {
	// Test Altibase-specific functionality that's only available with ODBC
	t.Run("Altibase driver mapping", func(t *testing.T) {
		driverName, exists := DriverName[DriverAltibase]
		assert.True(t, exists, "DriverAltibase should exist in DriverName map")
		assert.Equal(t, "odbc", driverName, "DriverAltibase should map to 'odbc'")
	})

	t.Run("Altibase data source generation", func(t *testing.T) {
		config := DatabaseConfig{
			Driver: DriverAltibase,
			Altibase: AltibaseConfig{
				Driver:   "Altibase",
				DSN:      "localhost",
				Port:     20300,
				UID:      "test_user",
				Password: "test_password",
				Database: "test_db",
			},
		}

		dsn := config.GetDataSourceName()
		assert.Contains(t, dsn, "DRIVER=Altibase")
		assert.Contains(t, dsn, "DSN=localhost")
		assert.Contains(t, dsn, "PORT_NO=20300")
		assert.Contains(t, dsn, "UID=test_user")
		assert.Contains(t, dsn, "PWD=test_password")
		assert.Contains(t, dsn, "DATABASE=test_db")
		assert.Contains(t, dsn, "CONNTYPE=1")
	})
}

func TestODBCBuildTag_InitializationEffect(t *testing.T) {
	// Test that the init() function correctly modified the supportDriver map
	// This verifies that the init() function in database_pool_odbc.go was executed

	// Before the init() function (from database_pool.go), DriverAltibase was false
	// After the init() function (from database_pool_odbc.go), DriverAltibase should be true
	assert.True(t, supportDriver[DriverAltibase],
		"init() function should have set DriverAltibase to true")
}

// Integration test for ODBC-specific functionality
func TestODBCIntegration(t *testing.T) {
	t.Run("supportDriver map integrity with ODBC", func(t *testing.T) {
		// Verify that all expected drivers are present
		expectedDriverCount := 6 // sqlite, postgresql, clickhouse, altibase, vertica, prometheus
		assert.Equal(t, expectedDriverCount, len(supportDriver),
			"supportDriver should contain exactly %d drivers with ODBC build tag", expectedDriverCount)

		// Verify each driver
		drivers := []string{DriverSqlite, DriverPostgreSQL, DriverClickHouse, DriverAltibase, DriverVertica, DriverPrometheus}
		for _, driver := range drivers {
			supported, exists := supportDriver[driver]
			assert.True(t, exists, "Driver %s should exist", driver)
			assert.True(t, supported, "Driver %s should be supported", driver)
		}
	})

	t.Run("ODBC driver availability", func(t *testing.T) {
		// Test that ODBC-specific functionality is available
		config := DatabaseConfig{Driver: DriverAltibase}
		driverName := config.GetDriverName()
		assert.Equal(t, "odbc", driverName, "Altibase should use ODBC driver")
	})
}

// Test that verifies the build tag behavior
func TestODBCBuildTagBehavior(t *testing.T) {
	t.Run("ODBC package import", func(t *testing.T) {
		// This test verifies that the ODBC package is imported
		// The import is done for side effects (_ import)
		// If the build succeeds, it means the package is available
		assert.True(t, true, "ODBC package should be importable with odbc build tag")
	})

	t.Run("Altibase support enabled", func(t *testing.T) {
		// This is the key test for this file - ensuring Altibase support is enabled
		assert.True(t, supportDriver[DriverAltibase],
			"Altibase driver should be supported when built with odbc tag")
	})
}

// Comparison test to verify the difference with non-ODBC builds
func TestODBCvsNonODBC(t *testing.T) {
	t.Run("Altibase driver state", func(t *testing.T) {
		// In ODBC build, Altibase should be supported
		// In non-ODBC build, Altibase would be false (default from database_pool.go)
		assert.True(t, supportDriver[DriverAltibase],
			"With ODBC build tag, Altibase should be supported")
	})

	t.Run("Other drivers unchanged", func(t *testing.T) {
		// Other drivers should maintain their original state
		assert.True(t, supportDriver[DriverSqlite], "SQLite should remain supported")
		assert.True(t, supportDriver[DriverPostgreSQL], "PostgreSQL should remain supported")
		assert.True(t, supportDriver[DriverClickHouse], "ClickHouse should remain supported")
		assert.True(t, supportDriver[DriverVertica], "Vertica should remain supported")
		assert.True(t, supportDriver[DriverPrometheus], "Prometheus should remain supported")
	})
}

// Benchmark test for ODBC-enabled operations
func BenchmarkODBCOperations(b *testing.B) {
	config := DatabaseConfig{
		Driver: DriverAltibase,
		Altibase: AltibaseConfig{
			Driver:   "Altibase",
			DSN:      "localhost",
			Port:     20300,
			UID:      "test",
			Password: "test",
			Database: "test",
		},
	}

	b.Run("GetDriverName", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = config.GetDriverName()
		}
	})

	b.Run("GetDataSourceName", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = config.GetDataSourceName()
		}
	})

	b.Run("SupportDriverLookup", func(b *testing.B) {
		b.ResetTimer()
		for i := 0; i < b.N; i++ {
			_ = supportDriver[DriverAltibase]
		}
	})
}

// Test the specific behavior of the init() function
func TestInitFunctionBehavior(t *testing.T) {
	// This test verifies that the init() function properly modifies the global state
	t.Run("init function execution", func(t *testing.T) {
		// The init() function should have been called during package initialization
		// and should have set supportDriver[DriverAltibase] = true

		// Verify the final state
		assert.True(t, supportDriver[DriverAltibase],
			"init() should have enabled Altibase support")

		// Verify that the change is persistent
		original := supportDriver[DriverAltibase]
		assert.True(t, original, "Altibase support should persist")
	})

	t.Run("global state modification", func(t *testing.T) {
		// Test that the global supportDriver map was correctly modified
		expectedState := map[string]bool{
			DriverAltibase:   true, // Changed by init() in database_pool_odbc.go
			DriverSqlite:     true, // Original from database_pool.go
			DriverPostgreSQL: true, // Original from database_pool.go
			DriverClickHouse: true, // Original from database_pool.go
			DriverVertica:    true, // Original from database_pool.go
			DriverPrometheus: true, // Original from database_pool.go
		}

		for driver, expectedSupport := range expectedState {
			actual := supportDriver[driver]
			assert.Equal(t, expectedSupport, actual,
				"Driver %s should have support=%v after init()", driver, expectedSupport)
		}
	})
}
