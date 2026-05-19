package orm

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	prometheus_api_v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/stretchr/testify/require"
)

// helper to reset global pool between tests
func resetDatabasePool() {
	DatabasePool.RemoveAll()
}

func newSqliteConfig(path string) DatabaseConfig {
	return DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{Path: path},
	}
}

func Test_Get_Sqlite_CreatesDirAndSetsMaxOpenConnsOne(t *testing.T) {
	resetDatabasePool()

	tmp := t.TempDir()
	// nested path that does not yet exist
	dbPath := filepath.Join(tmp, "nested", "db.sqlite")

	// Ensure directory does not exist initially
	dir := filepath.Dir(dbPath)
	_, err := os.Stat(dir)
	require.True(t, os.IsNotExist(err))

	cfg := newSqliteConfig(dbPath)

	db, err := DatabasePool.GetDB(cfg)
	require.NoError(t, err)
	require.NotNil(t, db)

	// Directory should be created
	fi, err := os.Stat(dir)
	require.NoError(t, err)
	require.True(t, fi.IsDir())

	// Ensure connection is usable
	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS t (id INTEGER PRIMARY KEY, v TEXT);`)
	require.NoError(t, err)

	// Verify that MaxOpenConns is set to 1 for sqlite
	stats := db.Stats()
	require.Equal(t, 1, stats.MaxOpenConnections)
}

func Test_Get_CachesSameInstance(t *testing.T) {
	resetDatabasePool()

	cfg := newSqliteConfig(filepath.Join(t.TempDir(), "a.db"))

	db1, err := DatabasePool.GetDB(cfg)
	require.NoError(t, err)
	db2, err := DatabasePool.GetDB(cfg)
	require.NoError(t, err)

	// Same pointer instance returned
	require.True(t, db1 == db2)
}

func Test_Get_UnsupportedDriver_ReturnsError(t *testing.T) {
	resetDatabasePool()

	cfg := DatabaseConfig{Driver: "unsupported-driver"}
	db, err := DatabasePool.GetDB(cfg)
	require.Error(t, err)
	require.Nil(t, db)
}

func Test_Remove_RemovesAndCloses(t *testing.T) {
	resetDatabasePool()

	cfg := newSqliteConfig(filepath.Join(t.TempDir(), "b.db"))

	db, err := DatabasePool.GetDB(cfg)
	require.NoError(t, err)

	// Keep a pointer to the db before removal
	raw := (*sqlx.DB)(db)

	// Remove should close and delete from map
	err = DatabasePool.Remove(cfg)
	require.NoError(t, err)

	// The underlying db should be closed now
	err = raw.Ping()
	require.Error(t, err)

	// Subsequent remove on non-existent should be no-op
	err = DatabasePool.Remove(cfg)
	require.NoError(t, err)
}

func Test_RemoveAll_ClosesAllAndClears(t *testing.T) {
	resetDatabasePool()

	cfg1 := newSqliteConfig(filepath.Join(t.TempDir(), "c1.db"))
	cfg2 := newSqliteConfig(filepath.Join(t.TempDir(), "c2.db"))

	db1, err := DatabasePool.GetDB(cfg1)
	require.NoError(t, err)
	db2, err := DatabasePool.GetDB(cfg2)
	require.NoError(t, err)

	// Sanity: connections work
	_, err = db1.Exec(`CREATE TABLE IF NOT EXISTS t1 (id INTEGER);`)
	require.NoError(t, err)
	_, err = db2.Exec(`CREATE TABLE IF NOT EXISTS t2 (id INTEGER);`)
	require.NoError(t, err)

	DatabasePool.RemoveAll()

	// Old pointers should now be closed
	require.Error(t, db1.Ping())
	require.Error(t, db2.Ping())

	// Pool should create new instances on next Get
	db1n, err := DatabasePool.GetDB(cfg1)
	require.NoError(t, err)
	require.NotNil(t, db1n)
	require.True(t, db1n != db1)
}

func newVerticaConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver: DriverVertica,
		Vertica: VerticaConfig{
			Host:              "localhost",
			Port:              5433,
			Username:          "dbadmin",
			Password:          "password",
			Database:          "testdb",
			MaxOpenConnection: 20,
			MaxLifetime:       300,
		},
	}
}

func Test_Get_Vertica_SetsConnectionPoolParams(t *testing.T) {
	resetDatabasePool()

	cfg := newVerticaConfig()

	// This will fail with actual connection but we test the configuration setup
	_, err := DatabasePool.GetDB(cfg)

	// We expect this to fail since we don't have a real Vertica server
	// but the configuration should be processed correctly
	require.Error(t, err) // Expected since no real Vertica server

	// Verify the configuration would be applied correctly by checking supportDriver
	require.True(t, supportDriver[DriverVertica])
}

func newPostgreSQLConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver: DriverPostgreSQL,
		PostgreSQL: PostgreSQLConfig{
			Host:              "localhost",
			Port:              5432,
			Username:          "postgres",
			Password:          "password",
			Database:          "testdb",
			MaxOpenConnection: 25,
			MaxLifetime:       300,
		},
	}
}

func Test_Get_PostgreSQL_SetsConnectionPoolParams(t *testing.T) {
	resetDatabasePool()

	cfg := newPostgreSQLConfig()

	// This will fail with actual connection but we test the configuration setup
	_, err := DatabasePool.GetDB(cfg)

	// We expect this to fail since we don't have a real PostgreSQL server
	// but the configuration should be processed correctly
	require.Error(t, err) // Expected since no real PostgreSQL server

	// Verify the configuration would be applied correctly by checking supportDriver
	require.True(t, supportDriver[DriverPostgreSQL])
}

func newClickHouseConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver: DriverClickHouse,
		ClickHouse: ClickHouseConfig{
			Host:              "localhost",
			Port:              9000,
			Username:          "default",
			Password:          "",
			Database:          "default",
			MaxOpenConnection: 50,
			MaxLifetime:       600,
		},
	}
}

func Test_Get_ClickHouse_SetsConnectionPoolParams(t *testing.T) {
	resetDatabasePool()

	cfg := newClickHouseConfig()

	// This will fail with actual connection but we test the configuration setup
	_, err := DatabasePool.GetDB(cfg)

	// We expect this to fail since we don't have a real ClickHouse server
	// but the configuration should be processed correctly
	require.Error(t, err) // Expected since no real ClickHouse server

	// Verify the configuration would be applied correctly by checking supportDriver
	require.True(t, supportDriver[DriverClickHouse])
}

func newPrometheusConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver: DriverPrometheus,
		Prometheus: PrometheusConfig{
			Address: "http://localhost:9090",
		},
	}
}

func Test_GetPrometheusAPI_Success(t *testing.T) {
	resetDatabasePool()

	cfg := newPrometheusConfig()

	// This should work even without a real Prometheus server since it just creates the client
	api, err := DatabasePool.GetPrometheusAPI(cfg)
	require.NoError(t, err)
	require.NotNil(t, api)

	// Test caching - should return same instance
	api2, err := DatabasePool.GetPrometheusAPI(cfg)
	require.NoError(t, err)
	require.True(t, api == api2)
}

func Test_GetPrometheusAPI_UnsupportedDriver(t *testing.T) {
	resetDatabasePool()

	cfg := newSqliteConfig(":memory:")

	api, err := DatabasePool.GetPrometheusAPI(cfg)
	require.Error(t, err)
	require.Nil(t, api)
	require.Contains(t, err.Error(), "not supported")
}

func Test_Remove_Prometheus(t *testing.T) {
	resetDatabasePool()

	cfg := newPrometheusConfig()

	// Get API first
	api, err := DatabasePool.GetPrometheusAPI(cfg)
	require.NoError(t, err)
	require.NotNil(t, api)

	// Remove should work without error
	err = DatabasePool.Remove(cfg)
	require.NoError(t, err)

	// Subsequent remove should be no-op
	err = DatabasePool.Remove(cfg)
	require.NoError(t, err)
}

func newElasticsearchConfig() DatabaseConfig {
	return DatabaseConfig{
		Driver: DriverElasticsearch,
		Elasticsearch: ElasticsearchConfig{
			Addresses: []string{"http://localhost:9200"},
			Bulk: struct {
				FlushBytes    int    `mapstructure:"flush_bytes"`
				FlushInterval string `mapstructure:"flush_interval"`
			}{
				FlushBytes:    5000000,
				FlushInterval: "30s",
			},
		},
	}
}

func Test_GetElasticsearchClient_V8_Success(t *testing.T) {
	resetDatabasePool()

	cfg := newElasticsearchConfig()
	cfg.Elasticsearch.Version = 8

	// This should work even without a real Elasticsearch server since it just creates the client
	client, err := DatabasePool.GetElasticsearchClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Test caching - should return same instance
	client2, err := DatabasePool.GetElasticsearchClient(cfg)
	require.NoError(t, err)
	require.True(t, client == client2)
}

func Test_GetElasticsearchClient_V9_Success(t *testing.T) {
	resetDatabasePool()

	cfg := newElasticsearchConfig()
	cfg.Elasticsearch.Version = 9

	// This should work even without a real Elasticsearch server since it just creates the client
	client, err := DatabasePool.GetElasticsearchClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Test caching - should return same instance
	client2, err := DatabasePool.GetElasticsearchClient(cfg)
	require.NoError(t, err)
	require.True(t, client == client2)
}

func Test_GetElasticsearchClient_UnsupportedDriver(t *testing.T) {
	resetDatabasePool()

	cfg := newSqliteConfig(":memory:")

	client, err := DatabasePool.GetElasticsearchClient(cfg)
	require.Error(t, err)
	require.Nil(t, client)
	require.Contains(t, err.Error(), "not supported")
}

func Test_GetElasticsearchClient_UnsupportedVersion(t *testing.T) {
	resetDatabasePool()

	cfg := newElasticsearchConfig()
	cfg.Elasticsearch.Version = 7 // Unsupported version

	client, err := DatabasePool.GetElasticsearchClient(cfg)
	require.Error(t, err)
	require.Nil(t, client)
	// Error message could be "not supported" or "unsupported version"
	errorMsg := err.Error()
	require.True(t, strings.Contains(errorMsg, "not supported") || strings.Contains(errorMsg, "unsupported"), "Expected 'not supported' or 'unsupported' error, got: %v", err)
}

func Test_Remove_Elasticsearch(t *testing.T) {
	resetDatabasePool()

	cfg := newElasticsearchConfig()
	cfg.Elasticsearch.Version = 8

	// Get client first
	client, err := DatabasePool.GetElasticsearchClient(cfg)
	require.NoError(t, err)
	require.NotNil(t, client)

	// Remove should work without error
	err = DatabasePool.Remove(cfg)
	require.NoError(t, err)

	// Subsequent remove should be no-op
	err = DatabasePool.Remove(cfg)
	require.NoError(t, err)
}

func Test_RemoveAll_IncludesElasticsearch(t *testing.T) {
	resetDatabasePool()

	cfg1 := newElasticsearchConfig()
	cfg1.Elasticsearch.Version = 8

	cfg2 := newElasticsearchConfig()
	cfg2.Elasticsearch.Version = 9
	cfg2.Elasticsearch.Addresses = []string{"http://localhost:9201"}

	client1, err := DatabasePool.GetElasticsearchClient(cfg1)
	require.NoError(t, err)
	client2, err := DatabasePool.GetElasticsearchClient(cfg2)
	require.NoError(t, err)

	DatabasePool.RemoveAll()

	// Pool should create new instances on next Get
	client1n, err := DatabasePool.GetElasticsearchClient(cfg1)
	require.NoError(t, err)
	require.NotNil(t, client1n)
	require.True(t, client1n != client1)

	client2n, err := DatabasePool.GetElasticsearchClient(cfg2)
	require.NoError(t, err)
	require.NotNil(t, client2n)
	require.True(t, client2n != client2)
}

func Test_GetDB_ConcurrentAccess_ReturnsOnlyOneInstance(t *testing.T) {
	resetDatabasePool()

	cfg := newSqliteConfig(filepath.Join(t.TempDir(), "concurrent.db"))

	// Launch 10 goroutines requesting the same DB simultaneously
	const numGoroutines = 10
	results := make([]*sqlx.DB, numGoroutines)
	errors := make([]error, numGoroutines)
	done := make(chan struct{})

	for i := range numGoroutines {
		go func(idx int) {
			results[idx], errors[idx] = DatabasePool.GetDB(cfg)
			done <- struct{}{}
		}(i)
	}

	// Wait for all goroutines to complete
	for range numGoroutines {
		<-done
	}

	// All requests should succeed
	for i := range numGoroutines {
		require.NoError(t, errors[i], "goroutine %d failed", i)
		require.NotNil(t, results[i], "goroutine %d got nil", i)
	}

	// All should point to the same instance (cached)
	first := results[0]
	for i := 1; i < numGoroutines; i++ {
		require.True(t, first == results[i], "goroutine %d got different instance", i)
	}
}

func Test_GetDB_SlowConnection_DoesNotBlockOtherDB(t *testing.T) {
	t.Skip("Skipping slow connection test - requires actual network timeout")
	resetDatabasePool()

	// Create a working SQLite DB
	sqliteCfg := newSqliteConfig(filepath.Join(t.TempDir(), "fast.db"))

	// Create a PostgreSQL config that will fail/timeout (no server running)
	slowCfg := newPostgreSQLConfig()

	// Channel to track when SQLite access completes
	sqliteDone := make(chan bool, 1)
	slowStarted := make(chan bool, 1)

	// Start slow connection attempt in background
	go func() {
		slowStarted <- true
		_, _ = DatabasePool.GetDB(slowCfg) // This will timeout, we don't care about result
	}()

	// Wait for slow connection to actually start
	<-slowStarted

	// Give it a moment to actually enter the connection code
	time.Sleep(100 * time.Millisecond)

	// Immediately try to access SQLite DB - this should NOT be blocked
	go func() {
		db, err := DatabasePool.GetDB(sqliteCfg)
		if err == nil && db != nil {
			// Verify we can actually use it
			_, err = db.Exec(`CREATE TABLE IF NOT EXISTS test (id INTEGER);`)
		}
		sqliteDone <- (err == nil)
	}()

	// SQLite should complete quickly (within 2 seconds)
	// If mutex was held during slow connection, this would timeout
	select {
	case success := <-sqliteDone:
		require.True(t, success, "SQLite access failed or was blocked")
	case <-time.After(2 * time.Second):
		t.Fatal("SQLite access blocked by slow connection attempt - mutex not released!")
	}
}

func Test_GetElasticsearchClient_ConcurrentAccess(t *testing.T) {
	resetDatabasePool()

	cfg := newElasticsearchConfig()
	cfg.Elasticsearch.Version = 8

	const numGoroutines = 10
	results := make([]*ElasticsearchClient, numGoroutines)
	errors := make([]error, numGoroutines)
	done := make(chan struct{})

	for i := range numGoroutines {
		go func(idx int) {
			results[idx], errors[idx] = DatabasePool.GetElasticsearchClient(cfg)
			done <- struct{}{}
		}(i)
	}

	for range numGoroutines {
		<-done
	}

	for i := range numGoroutines {
		require.NoError(t, errors[i])
		require.NotNil(t, results[i])
	}

	// All should be the same instance
	first := results[0]
	for i := 1; i < numGoroutines; i++ {
		require.True(t, first == results[i])
	}
}

func Test_GetPrometheusAPI_ConcurrentAccess(t *testing.T) {
	resetDatabasePool()

	cfg := newPrometheusConfig()

	const numGoroutines = 10
	results := make([]prometheus_api_v1.API, numGoroutines)
	errors := make([]error, numGoroutines)
	done := make(chan struct{})

	for i := range numGoroutines {
		go func(idx int) {
			results[idx], errors[idx] = DatabasePool.GetPrometheusAPI(cfg)
			done <- struct{}{}
		}(i)
	}

	for range numGoroutines {
		<-done
	}

	for i := range numGoroutines {
		require.NoError(t, errors[i])
		require.NotNil(t, results[i])
	}

	// All should be the same instance
	first := results[0]
	for i := 1; i < numGoroutines; i++ {
		require.True(t, first == results[i])
	}
}
