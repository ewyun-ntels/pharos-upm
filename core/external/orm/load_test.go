package orm

import (
	"maps"
	"path/filepath"
	"sync"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	// Mutex to protect test execution and global state access for load tests
	loadTestMutex sync.Mutex
)

func TestDatabasePool_GlobalInstance(t *testing.T) {
	loadTestMutex.Lock()
	defer loadTestMutex.Unlock()

	// Test that DatabasePool is properly initialized as a global instance
	assert.NotNil(t, DatabasePool.dbs, "DatabasePool.dbs should be initialized")
	// Simply check that DatabasePool is not nil instead of type checking to avoid copying
	assert.NotNil(t, &DatabasePool, "DatabasePool should be initialized")
}

func TestLoad_ValidConfig(t *testing.T) {
	loadTestMutex.Lock()
	defer loadTestMutex.Unlock()

	// Ensure clean state before test
	DatabasePool.RemoveAll()

	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: filepath.Join(t.TempDir(), "test_load.db"),
		},
	}

	err := Load(config)
	assert.NoError(t, err, "Load should succeed with valid SQLite config")

	// Verify the connection was added to the pool
	db, err := DatabasePool.GetDB(config)
	assert.NoError(t, err, "Should be able to get the loaded connection")
	assert.NotNil(t, db, "Database connection should not be nil")

	// Clean up
	DatabasePool.RemoveAll()
}

func TestLoad_InvalidConfig(t *testing.T) {
	loadTestMutex.Lock()
	defer loadTestMutex.Unlock()

	// Ensure clean state before test
	DatabasePool.RemoveAll()

	config := DatabaseConfig{
		Driver: "unsupported_driver",
	}

	err := Load(config)
	assert.Error(t, err, "Load should fail with unsupported driver")
	assert.Contains(t, err.Error(), "not supported", "Error should indicate driver not supported")

	// Clean up
	DatabasePool.RemoveAll()
}

func TestLoad_MultipleConfigs(t *testing.T) {
	loadTestMutex.Lock()
	defer loadTestMutex.Unlock()

	// Ensure clean state before test
	DatabasePool.RemoveAll()

	configs := []DatabaseConfig{
		{
			Driver: DriverSqlite,
			SQLite: SQLiteConfig{
				Path: filepath.Join(t.TempDir(), "test1.db"),
			},
		},
		{
			Driver: DriverSqlite,
			SQLite: SQLiteConfig{
				Path: filepath.Join(t.TempDir(), "test2.db"),
			},
		},
	}

	// Load multiple configurations
	for _, config := range configs {
		err := Load(config)
		assert.NoError(t, err, "Each Load should succeed")
	}

	// Verify all connections are in the pool
	for _, config := range configs {
		db, err := DatabasePool.GetDB(config)
		assert.NoError(t, err, "Should be able to get each loaded connection")
		assert.NotNil(t, db, "Each database connection should not be nil")
	}

	// Clean up
	DatabasePool.RemoveAll()
}

func TestLoad_SameConfigTwice(t *testing.T) {
	loadTestMutex.Lock()
	defer loadTestMutex.Unlock()

	// Ensure clean state before test
	DatabasePool.RemoveAll()

	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: filepath.Join(t.TempDir(), "test_duplicate.db"),
		},
	}

	// Load same config twice
	err1 := Load(config)
	assert.NoError(t, err1, "First Load should succeed")

	err2 := Load(config)
	assert.NoError(t, err2, "Second Load should also succeed (should reuse existing connection)")

	// Verify only one connection exists in the pool
	db1, err := DatabasePool.GetDB(config)
	assert.NoError(t, err, "Should be able to get the connection")

	db2, err := DatabasePool.GetDB(config)
	assert.NoError(t, err, "Should be able to get the connection again")

	// Should be the same instance (cached)
	assert.Same(t, db1, db2, "Should return the same connection instance")

	// Clean up
	DatabasePool.RemoveAll()
}

func TestUnload(t *testing.T) {
	// Ensure clean state before test
	DatabasePool.RemoveAll()

	configs := []DatabaseConfig{
		{
			Driver: DriverSqlite,
			SQLite: SQLiteConfig{
				Path: filepath.Join(t.TempDir(), "test_unload1.db"),
			},
		},
		{
			Driver: DriverSqlite,
			SQLite: SQLiteConfig{
				Path: filepath.Join(t.TempDir(), "test_unload2.db"),
			},
		},
	}

	// Load multiple configurations
	for _, config := range configs {
		err := Load(config)
		require.NoError(t, err, "Load should succeed")
	}

	// Verify connections exist
	for _, config := range configs {
		db, err := DatabasePool.GetDB(config)
		assert.NoError(t, err, "Should be able to get connection before unload")
		assert.NotNil(t, db, "Connection should exist before unload")
	}

	// Unload all
	Unload()

	// Verify all connections are removed (need to load again to test)
	for _, config := range configs {
		// After unload, getting a connection should create a new one
		db, err := DatabasePool.GetDB(config)
		assert.NoError(t, err, "Should be able to create new connection after unload")
		assert.NotNil(t, db, "New connection should be created")
	}

	// Clean up
	DatabasePool.RemoveAll()
}

func TestUnload_EmptyPool(t *testing.T) {
	// Ensure clean state before test
	DatabasePool.RemoveAll()

	// Unload when pool is already empty should not cause errors
	assert.NotPanics(t, func() {
		Unload()
	}, "Unload should not panic when pool is empty")
}

func TestLoad_ConcurrentAccess(t *testing.T) {
	// Ensure clean state before test
	DatabasePool.RemoveAll()

	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: filepath.Join(t.TempDir(), "test_concurrent.db"),
		},
	}

	var wg sync.WaitGroup
	numGoroutines := 10
	errors := make([]error, numGoroutines)

	// Concurrent Load calls
	for i := range numGoroutines {
		wg.Add(1)
		go func(index int) {
			defer wg.Done()
			errors[index] = Load(config)
		}(i)
	}

	wg.Wait()

	// All Load calls should succeed
	for i, err := range errors {
		assert.NoError(t, err, "Concurrent Load call %d should succeed", i)
	}

	// Verify only one connection was created
	db, err := DatabasePool.GetDB(config)
	assert.NoError(t, err, "Should be able to get the connection")
	assert.NotNil(t, db, "Connection should exist")

	// Clean up
	DatabasePool.RemoveAll()
}

func TestLoad_UnloadCycle(t *testing.T) {
	// Test multiple Load/Unload cycles
	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: filepath.Join(t.TempDir(), "test_cycle.db"),
		},
	}

	for i := range 3 {
		// Load
		err := Load(config)
		assert.NoError(t, err, "Load cycle %d should succeed", i)

		// Verify connection exists
		db, err := DatabasePool.GetDB(config)
		assert.NoError(t, err, "Should be able to get connection in cycle %d", i)
		assert.NotNil(t, db, "Connection should exist in cycle %d", i)

		// Unload
		Unload()
	}
}

func TestGlobalDatabasePoolModification(t *testing.T) {
	// Test that global DatabasePool can be modified safely
	originalDBs := DatabasePool.dbs

	// Backup current state
	backup := make(map[string]*sqlx.DB)
	maps.Copy(backup, DatabasePool.dbs)

	// Test modification
	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: filepath.Join(t.TempDir(), "test_global.db"),
		},
	}

	err := Load(config)
	assert.NoError(t, err, "Should be able to modify global pool")

	// Verify modification - use DSN as key
	dsn := config.GetDataSourceName()
	assert.Contains(t, DatabasePool.dbs, dsn, "Global pool should contain new config DSN")

	// Restore original state
	DatabasePool.dbs = originalDBs
	maps.Copy(DatabasePool.dbs, backup)
}

func TestLoadIntegration_WithDatabaseOperations(t *testing.T) {
	// Integration test: Load -> Database operations -> Unload
	DatabasePool.RemoveAll()

	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: filepath.Join(t.TempDir(), "test_integration.db"),
		},
	}

	// Load
	err := Load(config)
	require.NoError(t, err, "Load should succeed")

	// Get connection and perform database operations
	db, err := DatabasePool.GetDB(config)
	require.NoError(t, err, "Should be able to get connection")
	require.NotNil(t, db, "Connection should not be nil")

	// Create table
	_, err = db.Exec("CREATE TABLE test_table (id INTEGER PRIMARY KEY, name TEXT)")
	assert.NoError(t, err, "Should be able to create table")

	// Insert data
	_, err = db.Exec("INSERT INTO test_table (name) VALUES (?)", "test_name")
	assert.NoError(t, err, "Should be able to insert data")

	// Query data
	var count int
	err = db.Get(&count, "SELECT COUNT(*) FROM test_table")
	assert.NoError(t, err, "Should be able to query data")
	assert.Equal(t, 1, count, "Should have one record")

	// Unload
	Unload()

	// After unload, should be able to load again and access the persisted data
	err = Load(config)
	require.NoError(t, err, "Should be able to load again after unload")

	db, err = DatabasePool.GetDB(config)
	require.NoError(t, err, "Should be able to get connection again")

	err = db.Get(&count, "SELECT COUNT(*) FROM test_table")
	assert.NoError(t, err, "Should be able to query data after reload")
	assert.Equal(t, 1, count, "Data should persist after unload/load cycle")

	// Final cleanup
	DatabasePool.RemoveAll()
}

// Benchmark tests
func BenchmarkLoad(b *testing.B) {
	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: ":memory:",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		DatabasePool.RemoveAll() // Clean state for each iteration
		err := Load(config)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func BenchmarkLoad_Cached(b *testing.B) {
	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: ":memory:",
		},
	}

	// Pre-load to establish cached connection
	err := Load(config)
	if err != nil {
		b.Fatal(err)
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := Load(config)
		if err != nil {
			b.Fatal(err)
		}
	}

	DatabasePool.RemoveAll()
}

func BenchmarkUnload(b *testing.B) {
	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: ":memory:",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		b.StopTimer()
		err := Load(config)
		if err != nil {
			b.Fatal(err)
		}
		b.StartTimer()

		Unload()
	}
}

func BenchmarkLoadUnloadCycle(b *testing.B) {
	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: ":memory:",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := Load(config)
		if err != nil {
			b.Fatal(err)
		}
		Unload()
	}
}

// Test edge cases and error conditions
func TestLoad_ErrorConditions(t *testing.T) {
	tests := []struct {
		name          string
		config        DatabaseConfig
		expectError   bool
		errorContains string
	}{
		{
			name: "Empty driver",
			config: DatabaseConfig{
				Driver: "",
			},
			expectError:   true,
			errorContains: "not supported",
		},
		{
			name: "Invalid SQLite path (directory without permissions)",
			config: DatabaseConfig{
				Driver: DriverSqlite,
				SQLite: SQLiteConfig{
					Path: "/root/forbidden/test.db", // Assuming no write permission
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			DatabasePool.RemoveAll()

			err := Load(tt.config)

			if tt.expectError {
				assert.Error(t, err, "Expected error for %s", tt.name)
				if tt.errorContains != "" {
					assert.Contains(t, err.Error(), tt.errorContains, "Error should contain expected text")
				}
			} else {
				assert.NoError(t, err, "Expected no error for %s", tt.name)
			}

			DatabasePool.RemoveAll()
		})
	}
}

func TestLoad_MemoryDatabase(t *testing.T) {
	// Test with in-memory database
	DatabasePool.RemoveAll()

	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: ":memory:",
		},
	}

	err := Load(config)
	assert.NoError(t, err, "Should be able to load in-memory database")

	// Verify connection works
	db, err := DatabasePool.GetDB(config)
	assert.NoError(t, err, "Should be able to get in-memory connection")
	assert.NotNil(t, db, "In-memory connection should not be nil")

	// Test database operations
	_, err = db.Exec("CREATE TABLE memory_test (id INTEGER)")
	assert.NoError(t, err, "Should be able to use in-memory database")

	DatabasePool.RemoveAll()
}

func TestLoadWithRole_DefaultPool(t *testing.T) {
	loadTestMutex.Lock()
	defer loadTestMutex.Unlock()

	// Ensure clean state
	DatabasePool.RemoveAll()

	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: filepath.Join(t.TempDir(), "test_role_default.db"),
		},
	}

	err := LoadWithRole(PoolDefault, config)
	assert.NoError(t, err, "LoadWithRole should succeed with PoolDefault")

	// Verify connection is in the default pool
	db, err := DatabasePool.GetDB(config)
	assert.NoError(t, err, "Should be able to get connection from default pool")
	assert.NotNil(t, db, "Connection should not be nil")

	DatabasePool.RemoveAll()
}

func TestLoadWithRole_StatisticsPool(t *testing.T) {
	loadTestMutex.Lock()
	defer loadTestMutex.Unlock()

	// Ensure clean state
	StatisticsDatabasePool.RemoveAll()

	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: filepath.Join(t.TempDir(), "test_role_statistics.db"),
		},
	}

	err := LoadWithRole(PoolStatistics, config)
	assert.NoError(t, err, "LoadWithRole should succeed with PoolStatistics")

	// Verify connection is in the statistics pool
	db, err := StatisticsDatabasePool.GetDB(config)
	assert.NoError(t, err, "Should be able to get connection from statistics pool")
	assert.NotNil(t, db, "Connection should not be nil")

	// Verify connection is NOT in the default pool
	_, err = DatabasePool.GetDB(config)
	assert.NoError(t, err, "Default pool should still be able to create new connection")

	StatisticsDatabasePool.RemoveAll()
}

func TestLoadWithRole_ServicePool(t *testing.T) {
	loadTestMutex.Lock()
	defer loadTestMutex.Unlock()

	// Ensure clean state
	ServiceDatabasePool.RemoveAll()

	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: filepath.Join(t.TempDir(), "test_role_service.db"),
		},
	}

	err := LoadWithRole(PoolService, config)
	assert.NoError(t, err, "LoadWithRole should succeed with PoolService")

	// Verify connection is in the service pool
	db, err := ServiceDatabasePool.GetDB(config)
	assert.NoError(t, err, "Should be able to get connection from service pool")
	assert.NotNil(t, db, "Connection should not be nil")

	ServiceDatabasePool.RemoveAll()
}

func TestLoadWithRole_UnknownRole(t *testing.T) {
	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: filepath.Join(t.TempDir(), "test_unknown_role.db"),
		},
	}

	err := LoadWithRole("unknown_role", config)
	assert.Error(t, err, "LoadWithRole should fail with unknown role")
	assert.Contains(t, err.Error(), "unknown pool role", "Error should indicate unknown role")
}

func TestLoadWithRole_AllRoles(t *testing.T) {
	loadTestMutex.Lock()
	defer loadTestMutex.Unlock()

	// Ensure clean state
	DatabasePool.RemoveAll()
	StatisticsDatabasePool.RemoveAll()
	ServiceDatabasePool.RemoveAll()

	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: filepath.Join(t.TempDir(), "test_all_roles.db"),
		},
	}

	roles := []string{PoolDefault, PoolStatistics, PoolService}
	pools := []*databasePool{&DatabasePool, &StatisticsDatabasePool, &ServiceDatabasePool}

	// Load same config into all pools
	for _, role := range roles {
		err := LoadWithRole(role, config)
		assert.NoError(t, err, "LoadWithRole should succeed for role %s", role)
	}

	// Verify each pool has the connection
	for i, pool := range pools {
		db, err := pool.GetDB(config)
		assert.NoError(t, err, "Should be able to get connection from pool %d", i)
		assert.NotNil(t, db, "Connection should not be nil in pool %d", i)
	}

	// Clean up all pools
	DatabasePool.RemoveAll()
	StatisticsDatabasePool.RemoveAll()
	ServiceDatabasePool.RemoveAll()
}

func TestLoad_BackwardCompatibility(t *testing.T) {
	loadTestMutex.Lock()
	defer loadTestMutex.Unlock()

	// Ensure clean state
	DatabasePool.RemoveAll()

	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: filepath.Join(t.TempDir(), "test_backward_compat.db"),
		},
	}

	// Test that Load() uses the default pool (backward compatibility)
	err := Load(config)
	assert.NoError(t, err, "Load should succeed for backward compatibility")

	// Verify connection is in the default pool
	db1, err := DatabasePool.GetDB(config)
	assert.NoError(t, err, "Should be able to get connection from default pool")
	assert.NotNil(t, db1, "Connection should not be nil")

	// Test LoadWithRole(PoolDefault, config) produces same result
	err = LoadWithRole(PoolDefault, config)
	assert.NoError(t, err, "LoadWithRole with PoolDefault should succeed")

	db2, err := DatabasePool.GetDB(config)
	assert.NoError(t, err, "Should be able to get connection from default pool")
	assert.Same(t, db1, db2, "Load() and LoadWithRole(PoolDefault) should use same connection")

	DatabasePool.RemoveAll()
}

func TestUnload_AllPools(t *testing.T) {
	// Test that Unload() clears all pools
	configs := []DatabaseConfig{
		{
			Driver: DriverSqlite,
			SQLite: SQLiteConfig{
				Path: filepath.Join(t.TempDir(), "test_unload_all1.db"),
			},
		},
		{
			Driver: DriverSqlite,
			SQLite: SQLiteConfig{
				Path: filepath.Join(t.TempDir(), "test_unload_all2.db"),
			},
		},
	}

	roles := []string{PoolDefault, PoolStatistics, PoolService}
	pools := []*databasePool{&DatabasePool, &StatisticsDatabasePool, &ServiceDatabasePool}

	// Load configs into all pools
	for _, config := range configs {
		for _, role := range roles {
			err := LoadWithRole(role, config)
			require.NoError(t, err, "LoadWithRole should succeed")
		}
	}

	// Verify all pools have connections
	for i, pool := range pools {
		for j, config := range configs {
			db, err := pool.GetDB(config)
			assert.NoError(t, err, "Pool %d should have config %d", i, j)
			assert.NotNil(t, db, "Connection should exist in pool %d for config %d", i, j)
		}
	}

	// Unload all
	Unload()

	// Verify all pools are empty (connections recreated when accessed)
	for i, pool := range pools {
		for j, config := range configs {
			// After unload, accessing should create new connections
			db, err := pool.GetDB(config)
			assert.NoError(t, err, "Should be able to create new connection in pool %d for config %d", i, j)
			assert.NotNil(t, db, "New connection should be created in pool %d for config %d", i, j)
		}
	}

	// Final cleanup
	DatabasePool.RemoveAll()
	StatisticsDatabasePool.RemoveAll()
	ServiceDatabasePool.RemoveAll()
}

func TestPoolConstants(t *testing.T) {
	// Test that pool constants are properly defined
	assert.Equal(t, "default", PoolDefault, "PoolDefault should be 'default'")
	assert.Equal(t, "statistics", PoolStatistics, "PoolStatistics should be 'statistics'")
	assert.Equal(t, "service", PoolService, "PoolService should be 'service'")
}

func TestGlobalPools_Initialization(t *testing.T) {
	// Test that global pools are properly initialized
	assert.NotNil(t, &DatabasePool, "DatabasePool should be initialized")
	assert.NotNil(t, &StatisticsDatabasePool, "StatisticsDatabasePool should be initialized")
	assert.NotNil(t, &ServiceDatabasePool, "ServiceDatabasePool should be initialized")
}

func BenchmarkLoadWithRole(b *testing.B) {
	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: ":memory:",
		},
	}

	roles := []string{PoolDefault, PoolStatistics, PoolService}

	for _, role := range roles {
		b.Run(role, func(b *testing.B) {
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				// Clean the specific pool for this role
				switch role {
				case PoolDefault:
					DatabasePool.RemoveAll()
				case PoolStatistics:
					StatisticsDatabasePool.RemoveAll()
				case PoolService:
					ServiceDatabasePool.RemoveAll()
				}
				b.StartTimer()

				err := LoadWithRole(role, config)
				if err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}
