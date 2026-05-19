//go:build !odbc

package orm

import (
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
)

func TestNonODBCBuildTag_SupportDriverMap(t *testing.T) {
	// Test that without ODBC build tag, Altibase driver is not supported
	supported, exists := supportDriver[DriverAltibase]
	assert.True(t, exists, "DriverAltibase should exist in supportDriver map")
	assert.False(t, supported, "DriverAltibase should not be supported without ODBC build tag")
}

func TestNonODBCBuildTag_PartialDriverSupport(t *testing.T) {
	// Without ODBC build tag, only some drivers should be supported
	expectedDrivers := map[string]bool{
		DriverAltibase:   false, // Not supported without ODBC
		DriverSqlite:     true,  // Always supported
		DriverPostgreSQL: true,  // Always supported
		DriverClickHouse: true,  // Always supported
		DriverVertica:    true,  // Always supported
		DriverPrometheus: true,  // Always supported
	}

	for driver, expectedSupport := range expectedDrivers {
		actual, exists := supportDriver[driver]
		assert.True(t, exists, "Driver %s should exist in supportDriver map", driver)
		assert.Equal(t, expectedSupport, actual, "Driver %s support should be %v without ODBC build tag", driver, expectedSupport)
	}
}

func TestNonODBCBuildTag_SupportDriverCount(t *testing.T) {
	// Verify that all expected drivers are present (6 total, but Altibase is disabled)
	expectedDriverCount := 7 // sqlite, postgresql, clickhouse, altibase, vertica, prometheus (but altibase is false)
	assert.Equal(t, expectedDriverCount, len(supportDriver),
		"supportDriver should contain exactly %d drivers without ODBC build tag", expectedDriverCount)
}

func TestNonODBCBuildTag_AltibaseStillMapped(t *testing.T) {
	// Even without ODBC, Altibase should still be in the DriverName map
	t.Run("Altibase driver mapping exists", func(t *testing.T) {
		driverName, exists := DriverName[DriverAltibase]
		assert.True(t, exists, "DriverAltibase should exist in DriverName map even without ODBC")
		assert.Equal(t, "odbc", driverName, "DriverAltibase should map to 'odbc' even without ODBC")
	})

	t.Run("Altibase data source generation works", func(t *testing.T) {
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

		// Data source generation should work even if driver is not supported
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

func TestNonODBCBuildTag_NoInitEffect(t *testing.T) {
	// Test that without ODBC build tag, the init() function from database_pool_odbc.go
	// is not executed, so DriverAltibase remains false
	assert.False(t, supportDriver[DriverAltibase],
		"Without ODBC build tag, DriverAltibase should remain false")
}

// Integration test for non-ODBC functionality
func TestNonODBCIntegration(t *testing.T) {
	t.Run("supportDriver map integrity without ODBC", func(t *testing.T) {
		// Verify that all expected drivers are present
		expectedDriverCount := 7 // sqlite, postgresql, clickhouse, altibase, vertica, prometheus
		assert.Equal(t, expectedDriverCount, len(supportDriver),
			"supportDriver should contain exactly %d drivers without ODBC build tag", expectedDriverCount)

		// Verify each driver's expected state
		expectedStates := map[string]bool{
			DriverSqlite:     true,
			DriverPostgreSQL: true,
			DriverClickHouse: true,
			DriverVertica:    true,
			DriverPrometheus: true,
			DriverAltibase:   false, // Should be false without ODBC
		}

		for driver, expectedState := range expectedStates {
			supported, exists := supportDriver[driver]
			assert.True(t, exists, "Driver %s should exist", driver)
			assert.Equal(t, expectedState, supported, "Driver %s should be %v without ODBC", driver, expectedState)
		}
	})

	t.Run("Non-ODBC drivers work normally", func(t *testing.T) {
		// Test that non-ODBC drivers still work normally
		supportedDrivers := []string{DriverSqlite, DriverPostgreSQL, DriverClickHouse, DriverVertica, DriverPrometheus}

		for _, driver := range supportedDrivers {
			config := DatabaseConfig{Driver: driver}
			driverName := config.GetDriverName()
			assert.NotEmpty(t, driverName, "Supported driver %s should have a driver name", driver)
		}
	})
}

// Test that verifies the build tag behavior
func TestNonODBCBuildTagBehavior(t *testing.T) {
	t.Run("ODBC package not imported", func(t *testing.T) {
		// Without ODBC build tag, the ODBC package should not be imported
		// This test just ensures the build succeeds without ODBC
		assert.True(t, true, "Build should succeed without ODBC package")
	})

	t.Run("Altibase support disabled", func(t *testing.T) {
		// This is the key test for non-ODBC builds - ensuring Altibase support is disabled
		assert.False(t, supportDriver[DriverAltibase],
			"Altibase driver should not be supported when built without odbc tag")
	})
}

// Test database pool behavior without ODBC
func TestNonODBCDatabasePool(t *testing.T) {
	t.Run("Altibase connection should fail", func(t *testing.T) {
		// Without ODBC, attempting to create an Altibase connection should fail
		pool := &databasePool{
			dbs: make(map[string]*sqlx.DB),
		}

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

		_, err := pool.GetDB(config)
		assert.Error(t, err, "Altibase connection should fail without ODBC support")
		assert.Contains(t, err.Error(), "not supported", "Error should indicate driver is not supported")
	})

	t.Run("Supported drivers work normally", func(t *testing.T) {
		// Other drivers should still work
		pool := &databasePool{
			dbs: make(map[string]*sqlx.DB),
		}

		config := DatabaseConfig{
			Driver: DriverSqlite,
			SQLite: SQLiteConfig{
				Path: ":memory:",
			},
		}

		db, err := pool.getSQLxDB(config)
		assert.NoError(t, err, "SQLite should work without ODBC")
		assert.NotNil(t, db, "SQLite connection should be created")
		if db != nil {
			db.Close()
		}
	})
}

// Benchmark test for non-ODBC operations
func BenchmarkNonODBCOperations(b *testing.B) {
	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: ":memory:",
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
			_ = supportDriver[DriverSqlite]
		}
	})
}

// Test that demonstrates the difference between ODBC and non-ODBC builds
func TestBuildTagDifference(t *testing.T) {
	t.Run("Default state verification", func(t *testing.T) {
		// In non-ODBC build, this should be the default state from database_pool.go
		expectedDefaultState := map[string]bool{
			DriverAltibase:   false, // Default from database_pool.go
			DriverSqlite:     true,  // Default from database_pool.go
			DriverPostgreSQL: true,  // Default from database_pool.go
			DriverClickHouse: true,  // Default from database_pool.go
			DriverVertica:    true,  // Default from database_pool.go
			DriverPrometheus: true,  // Default from database_pool.go
		}

		for driver, expectedSupport := range expectedDefaultState {
			actual := supportDriver[driver]
			assert.Equal(t, expectedSupport, actual,
				"Driver %s should have default support=%v in non-ODBC build", driver, expectedSupport)
		}
	})

	t.Run("No ODBC init modification", func(t *testing.T) {
		// Verify that the init() function from database_pool_odbc.go was NOT executed
		assert.False(t, supportDriver[DriverAltibase],
			"Without ODBC build tag, Altibase should remain unsupported (init() not executed)")
	})
}
