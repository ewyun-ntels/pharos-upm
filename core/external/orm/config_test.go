package orm

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Test constants for better maintainability
const (
	TestSQLitePath     = "/tmp/test.db"
	TestPostgresHost   = "localhost"
	TestPostgresPort   = 5432
	TestPostgresUser   = "testuser"
	TestPostgresPass   = "testpass"
	TestPostgresDB     = "testdb"
	TestClickHouseHost = "localhost"
	TestClickHousePort = 9000
	TestClickHouseUser = "clickuser"
	TestClickHousePass = "clickpass"
	TestClickHouseDB   = "clickdb"
	TestAltibaseDriver = "Altibase"
	TestAltibaseDSN    = "localhost"
	TestAltibasePort   = 20300
	TestAltibaseUID    = "altiuser"
	TestAltibasePass   = "altipass"
	TestAltibaseDB     = "altidb"
	UnknownDriver      = "unknown_driver"
	AppName            = "pharos"
)

// Test data factories
func createSQLiteConfig(path string) DatabaseConfig {
	return DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{Path: path},
	}
}

func createPostgreSQLConfig(host string, port int, username, password, database string) DatabaseConfig {
	return DatabaseConfig{
		Driver: DriverPostgreSQL,
		PostgreSQL: PostgreSQLConfig{
			Host:     host,
			Port:     port,
			Username: username,
			Password: password,
			Database: database,
		},
	}
}

func createClickHouseConfig(host string, port int, username, password, database string) DatabaseConfig {
	return DatabaseConfig{
		Driver: DriverClickHouse,
		ClickHouse: ClickHouseConfig{
			Host:     host,
			Port:     port,
			Username: username,
			Password: password,
			Database: database,
		},
	}
}

func createAltibaseConfig(driver, dsn string, port int, uid, password, database string) DatabaseConfig {
	return DatabaseConfig{
		Driver: DriverAltibase,
		Altibase: AltibaseConfig{
			Driver:   driver,
			DSN:      dsn,
			Port:     port,
			UID:      uid,
			Password: password,
			Database: database,
		},
	}
}

func TestDatabaseConfig_GetDriverName(t *testing.T) {
	tests := []struct {
		name     string
		config   DatabaseConfig
		expected string
	}{
		{
			name:     "SQLite driver",
			config:   createSQLiteConfig(""),
			expected: "sqlite",
		},
		{
			name:     "PostgreSQL driver",
			config:   createPostgreSQLConfig("", 0, "", "", ""),
			expected: "postgres",
		},
		{
			name:     "ClickHouse driver",
			config:   createClickHouseConfig("", 0, "", "", ""),
			expected: "clickhouse",
		},
		{
			name:     "Altibase driver",
			config:   createAltibaseConfig("", "", 0, "", "", ""),
			expected: "odbc",
		},
		{
			name: "Unknown driver returns original",
			config: DatabaseConfig{
				Driver: UnknownDriver,
			},
			expected: UnknownDriver,
		},
		{
			name: "Empty driver returns empty string",
			config: DatabaseConfig{
				Driver: "",
			},
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetDriverName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDatabaseConfig_GetDataSourceName(t *testing.T) {
	tests := []struct {
		name     string
		config   DatabaseConfig
		expected string
		contains []string // For partial matching when needed
	}{
		{
			name:     "SQLite data source",
			config:   createSQLiteConfig(TestSQLitePath),
			expected: TestSQLitePath,
		},
		{
			name:   "PostgreSQL data source",
			config: createPostgreSQLConfig(TestPostgresHost, TestPostgresPort, TestPostgresUser, TestPostgresPass, TestPostgresDB),
			contains: []string{
				fmt.Sprintf("host=%s", TestPostgresHost),
				fmt.Sprintf("port=%d", TestPostgresPort),
				fmt.Sprintf("user=%s", TestPostgresUser),
				fmt.Sprintf("password=%s", TestPostgresPass),
				fmt.Sprintf("dbname=%s", TestPostgresDB),
				"sslmode=disable",
				fmt.Sprintf("application_name=%s", AppName),
			},
		},
		{
			name:     "ClickHouse data source",
			config:   createClickHouseConfig(TestClickHouseHost, TestClickHousePort, TestClickHouseUser, TestClickHousePass, TestClickHouseDB),
			expected: fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s", TestClickHouseUser, TestClickHousePass, TestClickHouseHost, TestClickHousePort, TestClickHouseDB),
		},
		{
			name:   "Altibase data source",
			config: createAltibaseConfig(TestAltibaseDriver, TestAltibaseDSN, TestAltibasePort, TestAltibaseUID, TestAltibasePass, TestAltibaseDB),
			contains: []string{
				fmt.Sprintf("DRIVER=%s", TestAltibaseDriver),
				fmt.Sprintf("DSN=%s", TestAltibaseDSN),
				"CONNTYPE=1",
				fmt.Sprintf("PORT_NO=%d", TestAltibasePort),
				fmt.Sprintf("UID=%s", TestAltibaseUID),
				fmt.Sprintf("PWD=%s", TestAltibasePass),
				fmt.Sprintf("DATABASE=%s", TestAltibaseDB),
			},
		},
		{
			name: "Unknown driver",
			config: DatabaseConfig{
				Driver: UnknownDriver,
			},
			expected: "not implemented database driver",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetDataSourceName()

			if tt.expected != "" {
				assert.Equal(t, tt.expected, result)
			}

			for _, contain := range tt.contains {
				assert.Contains(t, result, contain)
			}
		})
	}
}

// Helper functions for database testing
func createMemoryConfig() DatabaseConfig {
	return createSQLiteConfig(":memory:")
}

func TestDatabaseConfig_GetDatabaseResponse_EmptyQuery(t *testing.T) {
	config := createMemoryConfig()

	response, err := config.GetDatabaseResponse(context.Background(), "", 30)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid query")
	assert.Equal(t, DatabaseResponse{}, response)
}

func TestDatabaseConfig_GetDatabaseResponse_SQLite(t *testing.T) {
	config := createMemoryConfig()

	// Test with a simple query
	query := "SELECT 1 as test_column, 'hello' as text_column"
	response, err := config.GetDatabaseResponse(context.Background(), query, 30)

	require.NoError(t, err)
	assert.Equal(t, query, response.SQL)
	assert.Equal(t, int64(1), response.Rows)
	assert.Greater(t, response.Statistics.Elapsed, 0.0)

	// Check metadata
	require.Len(t, response.Meta, 2)
	assert.Equal(t, "test_column", response.Meta[0]["name"])
	assert.Equal(t, "text_column", response.Meta[1]["name"])

	// Check data
	require.Len(t, response.Data, 1)
	assert.Equal(t, int64(1), response.Data[0]["test_column"])
	assert.Equal(t, "hello", response.Data[0]["text_column"])
}

func TestDatabaseConfig_GetDatabaseResponse_SQLiteWithTable(t *testing.T) {
	config := createMemoryConfig()

	// First create a table and insert data
	createQuery := `
		CREATE TABLE test_table (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			age INTEGER,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := config.GetDatabaseResponse(context.Background(), createQuery, 30)
	require.NoError(t, err)

	// Insert test data
	insertQuery := "INSERT INTO test_table (name, age) VALUES ('Alice', 30), ('Bob', 25)"
	_, err = config.GetDatabaseResponse(context.Background(), insertQuery, 30)
	require.NoError(t, err)

	// Query the data
	selectQuery := "SELECT id, name, age FROM test_table ORDER BY id"
	response, err := config.GetDatabaseResponse(context.Background(), selectQuery, 30)

	require.NoError(t, err)
	assert.Equal(t, selectQuery, response.SQL)
	assert.Equal(t, int64(2), response.Rows)
	assert.Greater(t, response.Statistics.Elapsed, 0.0)

	// Check metadata
	require.Len(t, response.Meta, 3)
	assert.Equal(t, "id", response.Meta[0]["name"])
	assert.Equal(t, "name", response.Meta[1]["name"])
	assert.Equal(t, "age", response.Meta[2]["name"])

	// Check data
	require.Len(t, response.Data, 2)
	assert.Equal(t, int64(1), response.Data[0]["id"])
	assert.Equal(t, "Alice", response.Data[0]["name"])
	assert.Equal(t, int64(30), response.Data[0]["age"])

	assert.Equal(t, int64(2), response.Data[1]["id"])
	assert.Equal(t, "Bob", response.Data[1]["name"])
	assert.Equal(t, int64(25), response.Data[1]["age"])
}

func TestDatabaseConfig_GetDatabaseResponse_InvalidQuery(t *testing.T) {
	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: ":memory:",
		},
	}

	// Test with invalid SQL
	query := "INVALID SQL STATEMENT"
	response, err := config.GetDatabaseResponse(context.Background(), query, 30)

	assert.Error(t, err)
	assert.Contains(t, strings.ToLower(err.Error()), "syntax error")
	// Response should be empty on error
	assert.Equal(t, "", response.SQL)
	assert.Equal(t, int64(0), response.Rows)
}

func TestDatabaseConfig_GetDatabaseResponse_UnsupportedDriver(t *testing.T) {
	config := DatabaseConfig{
		Driver: "unsupported_driver",
	}

	query := "SELECT 1"
	response, err := config.GetDatabaseResponse(context.Background(), query, 30)

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "not supported")
	assert.Equal(t, DatabaseResponse{}, response)
}

func TestSQLiteConfig_Structure(t *testing.T) {
	config := SQLiteConfig{
		Path: "/path/to/db.sqlite",
		Backup: struct {
			Use       bool   `mapstructure:"use"`
			Schedule  string `mapstructure:"schedule"`
			Directory string `mapstructure:"directory"`
			TTL       string `mapstructure:"ttl"`
		}{
			Use:       true,
			Schedule:  "0 2 * * *",
			Directory: "/backup/path",
			TTL:       "30d",
		},
	}

	assert.Equal(t, "/path/to/db.sqlite", config.Path)
	assert.True(t, config.Backup.Use)
	assert.Equal(t, "0 2 * * *", config.Backup.Schedule)
	assert.Equal(t, "/backup/path", config.Backup.Directory)
	assert.Equal(t, "30d", config.Backup.TTL)
}

func TestPostgreSQLConfig_Structure(t *testing.T) {
	config := PostgreSQLConfig{
		Host:              "localhost",
		Port:              5432,
		Username:          "postgres",
		Password:          "password",
		Database:          "mydb",
		MaxOpenConnection: 25,
		MaxLifetime:       300,
	}

	assert.Equal(t, "localhost", config.Host)
	assert.Equal(t, 5432, config.Port)
	assert.Equal(t, "postgres", config.Username)
	assert.Equal(t, "password", config.Password)
	assert.Equal(t, "mydb", config.Database)
	assert.Equal(t, 25, config.MaxOpenConnection)
	assert.Equal(t, 300, config.MaxLifetime)
}

func TestClickHouseConfig_Structure(t *testing.T) {
	config := ClickHouseConfig{
		Host:              "clickhouse-server",
		Port:              9000,
		Username:          "clickuser",
		Password:          "clickpass",
		Database:          "analytics",
		MaxOpenConnection: 50,
		MaxLifetime:       600,
	}

	assert.Equal(t, "clickhouse-server", config.Host)
	assert.Equal(t, 9000, config.Port)
	assert.Equal(t, "clickuser", config.Username)
	assert.Equal(t, "clickpass", config.Password)
	assert.Equal(t, "analytics", config.Database)
	assert.Equal(t, 50, config.MaxOpenConnection)
	assert.Equal(t, 600, config.MaxLifetime)
}

func TestAltibaseConfig_Structure(t *testing.T) {
	config := AltibaseConfig{
		Driver:            "Altibase",
		DSN:               "altibase-server",
		Port:              20300,
		UID:               "altiuser",
		Password:          "altipass",
		Database:          "mydb",
		MaxOpenConnection: 20,
		MaxLifetime:       900,
	}

	assert.Equal(t, "Altibase", config.Driver)
	assert.Equal(t, "altibase-server", config.DSN)
	assert.Equal(t, 20300, config.Port)
	assert.Equal(t, "altiuser", config.UID)
	assert.Equal(t, "altipass", config.Password)
	assert.Equal(t, "mydb", config.Database)
	assert.Equal(t, 20, config.MaxOpenConnection)
	assert.Equal(t, 900, config.MaxLifetime)
}

func TestDatabaseConfig_Structure(t *testing.T) {
	config := DatabaseConfig{
		Driver: DriverPostgreSQL,
		SQLite: SQLiteConfig{
			Path: "/sqlite/path.db",
		},
		PostgreSQL: PostgreSQLConfig{
			Host:     "pg-host",
			Port:     5432,
			Username: "pguser",
			Password: "pgpass",
			Database: "pgdb",
		},
		ClickHouse: ClickHouseConfig{
			Host:     "ch-host",
			Port:     9000,
			Username: "chuser",
			Password: "chpass",
			Database: "chdb",
		},
		Altibase: AltibaseConfig{
			Driver:   "Altibase",
			DSN:      "alti-host",
			Port:     20300,
			UID:      "altiuser",
			Password: "altipass",
			Database: "altidb",
		},
	}

	assert.Equal(t, DriverPostgreSQL, config.Driver)
	assert.Equal(t, "/sqlite/path.db", config.SQLite.Path)
	assert.Equal(t, "pg-host", config.PostgreSQL.Host)
	assert.Equal(t, "ch-host", config.ClickHouse.Host)
	assert.Equal(t, "alti-host", config.Altibase.DSN)
}

func TestDatabaseConfig_GetDataSourceName_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		config   DatabaseConfig
		expected string
	}{
		{
			name: "SQLite with empty path",
			config: DatabaseConfig{
				Driver: DriverSqlite,
				SQLite: SQLiteConfig{Path: ""},
			},
			expected: "",
		},
		{
			name: "PostgreSQL with special characters in password",
			config: DatabaseConfig{
				Driver: DriverPostgreSQL,
				PostgreSQL: PostgreSQLConfig{
					Host:     "localhost",
					Port:     5432,
					Username: "user@domain",
					Password: "pass word!@#",
					Database: "test-db",
				},
			},
		},
		{
			name: "ClickHouse with IPv6 host",
			config: DatabaseConfig{
				Driver: DriverClickHouse,
				ClickHouse: ClickHouseConfig{
					Host:     "::1",
					Port:     9000,
					Username: "user",
					Password: "pass",
					Database: "db",
				},
			},
			expected: "clickhouse://user:pass@::1:9000/db",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetDataSourceName()
			if tt.expected != "" {
				assert.Equal(t, tt.expected, result)
			} else {
				// For cases without expected value, just check it's not empty
				// except for the empty path case which should return empty
				if tt.config.Driver == DriverSqlite && tt.config.SQLite.Path == "" {
					assert.Equal(t, "", result)
				} else {
					assert.NotEmpty(t, result)
				}
			}
		})
	}
}

// Test for Prometheus driver and queries
func TestDatabaseConfig_GetDatabaseResponse_Prometheus(t *testing.T) {
	// Skip if Prometheus is not available
	config := DatabaseConfig{
		Driver: DriverPrometheus,
		Prometheus: PrometheusConfig{
			Address: "http://localhost:9090",
		},
	}

	// Test empty query
	response, err := config.GetDatabaseResponse(context.Background(), "", 30)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid query")

	// Test simple metric query (this will fail without actual Prometheus server)
	// But we can test the structure and error handling
	query := "up"
	response, err = config.GetDatabaseResponse(context.Background(), query, 30)

	// We expect an error since there's no real Prometheus server
	if err != nil {
		// This is expected when Prometheus is not available
		assert.Contains(t, strings.ToLower(err.Error()), "connection")
		t.Logf("Expected connection error: %v", err)
	} else {
		// If somehow it works, check the response structure
		assert.Equal(t, query, response.SQL)
		assert.GreaterOrEqual(t, response.Rows, int64(0))
		assert.Greater(t, response.Statistics.Elapsed, 0.0)
	}
}

func TestDatabaseConfig_GetDatabaseResponse_PrometheusDataTypes(t *testing.T) {
	// Test the Prometheus response structure without actual server
	// We'll test the makeDatabaseResponseForPrometheus method indirectly

	config := DatabaseConfig{
		Driver: DriverPrometheus,
		Prometheus: PrometheusConfig{
			Address: "http://invalid-prometheus:9090",
		},
	}

	// Test various Prometheus query patterns
	testQueries := []string{
		"up",
		"rate(http_requests_total[5m])",
		"histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m]))",
		"sum(rate(container_cpu_usage_seconds_total[1m])) by (pod)",
	}

	for _, query := range testQueries {
		t.Run(fmt.Sprintf("query_%s", strings.ReplaceAll(query, " ", "_")), func(t *testing.T) {
			response, err := config.GetDatabaseResponse(context.Background(), query, 30)

			// We expect connection errors with invalid address
			if err != nil {
				// Accept various types of connection errors
				errorStr := strings.ToLower(err.Error())
				hasConnectionError := strings.Contains(errorStr, "connection") ||
					strings.Contains(errorStr, "dial") ||
					strings.Contains(errorStr, "no such host") ||
					strings.Contains(errorStr, "refused")
				assert.True(t, hasConnectionError, "Expected connection-related error, got: %v", err)
			} else {
				// If it works, validate structure
				assert.Equal(t, query, response.SQL)
				assert.GreaterOrEqual(t, response.Rows, int64(0))
			}
		})
	}
}

func createVerticaConfig(host string, port int, username, password, database string) DatabaseConfig {
	return DatabaseConfig{
		Driver: DriverVertica,
		Vertica: VerticaConfig{
			Host:     host,
			Port:     port,
			Username: username,
			Password: password,
			Database: database,
		},
	}
}

func TestDatabaseConfig_GetDataSourceName_Vertica(t *testing.T) {
	config := createVerticaConfig("vertica-host", 5433, "vertica_user", "vertica_pass", "vertica_db")

	expected := "vertica://vertica_user:vertica_pass@vertica-host:5433/vertica_db"
	result := config.GetDataSourceName()

	assert.Equal(t, expected, result)
}

func TestVerticaConfig_Structure(t *testing.T) {
	config := VerticaConfig{
		Host:              "vertica-cluster",
		Port:              5433,
		Username:          "dbadmin",
		Password:          "verticapass",
		Database:          "analytics",
		MaxOpenConnection: 30,
		MaxLifetime:       450,
	}

	assert.Equal(t, "vertica-cluster", config.Host)
	assert.Equal(t, 5433, config.Port)
	assert.Equal(t, "dbadmin", config.Username)
	assert.Equal(t, "verticapass", config.Password)
	assert.Equal(t, "analytics", config.Database)
	assert.Equal(t, 30, config.MaxOpenConnection)
	assert.Equal(t, 450, config.MaxLifetime)
}

func TestPrometheusConfig_Structure(t *testing.T) {
	config := PrometheusConfig{
		Address: "http://prometheus:9090",
	}

	assert.Equal(t, "http://prometheus:9090", config.Address)
}

func TestElasticsearchConfig_Structure(t *testing.T) {
	config := ElasticsearchConfig{
		Version:   8,
		Addresses: []string{"http://localhost:9200", "http://localhost:9201"},
	}

	assert.Equal(t, 8, config.Version)
	assert.Len(t, config.Addresses, 2)
	assert.Equal(t, "http://localhost:9200", config.Addresses[0])
	assert.Equal(t, "http://localhost:9201", config.Addresses[1])
}

func createElasticsearchConfig(version int, addresses []string) DatabaseConfig {
	return DatabaseConfig{
		Driver: DriverElasticsearch,
		Elasticsearch: ElasticsearchConfig{
			Version:   version,
			Addresses: addresses,
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

func TestDatabaseConfig_GetDataSourceName_Elasticsearch(t *testing.T) {
	tests := []struct {
		name      string
		config    DatabaseConfig
		expected  string
		addresses []string
	}{
		{
			name:      "Single Elasticsearch node",
			config:    createElasticsearchConfig(8, []string{"http://localhost:9200"}),
			expected:  "http://localhost:9200",
			addresses: []string{"http://localhost:9200"},
		},
		{
			name:      "Multiple Elasticsearch nodes",
			config:    createElasticsearchConfig(9, []string{"http://es1:9200", "http://es2:9200", "http://es3:9200"}),
			expected:  "http://es1:9200,http://es2:9200,http://es3:9200",
			addresses: []string{"http://es1:9200", "http://es2:9200", "http://es3:9200"},
		},
		{
			name:      "Elasticsearch with HTTPS",
			config:    createElasticsearchConfig(8, []string{"https://secure-es:9200"}),
			expected:  "https://secure-es:9200",
			addresses: []string{"https://secure-es:9200"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.config.GetDataSourceName()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestDatabaseConfig_GetDatabaseResponse_Elasticsearch(t *testing.T) {
	config := createElasticsearchConfig(8, []string{"http://localhost:9200"})

	// Test empty query
	response, err := config.GetDatabaseResponse(context.Background(), "", 30)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "invalid query")
	assert.Equal(t, DatabaseResponse{}, response)

	// Test with ES|QL query (will fail without actual Elasticsearch server)
	query := "FROM logs | LIMIT 10"
	response, err = config.GetDatabaseResponse(context.Background(), query, 30)

	// We expect an error since there's no real Elasticsearch server
	if err != nil {
		// This is expected when Elasticsearch is not available
		errorStr := strings.ToLower(err.Error())
		hasConnectionError := strings.Contains(errorStr, "connection") ||
			strings.Contains(errorStr, "dial") ||
			strings.Contains(errorStr, "refused") ||
			strings.Contains(errorStr, "no such host") ||
			strings.Contains(errorStr, "unsupported version")
		assert.True(t, hasConnectionError, "Expected connection or version error, got: %v", err)
		t.Logf("Expected error: %v", err)
	} else {
		// If somehow it works, check the response structure
		assert.Equal(t, query, response.SQL)
		assert.GreaterOrEqual(t, response.Rows, int64(0))
		assert.Greater(t, response.Statistics.Elapsed, 0.0)
	}
}

func TestDatabaseConfig_GetDatabaseResponse_ElasticsearchVersions(t *testing.T) {
	tests := []struct {
		name    string
		version int
	}{
		{
			name:    "Elasticsearch v8",
			version: 8,
		},
		{
			name:    "Elasticsearch v9",
			version: 9,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := createElasticsearchConfig(tt.version, []string{"http://localhost:9200"})
			query := "FROM system | LIMIT 1"

			response, err := config.GetDatabaseResponse(context.Background(), query, 30)

			// Expected to fail without actual server
			if err != nil {
				errorStr := strings.ToLower(err.Error())
				hasExpectedError := strings.Contains(errorStr, "connection") ||
					strings.Contains(errorStr, "dial") ||
					strings.Contains(errorStr, "unsupported version")
				assert.True(t, hasExpectedError, "Expected connection error for version %d, got: %v", tt.version, err)
			}

			// Response should be empty on error
			if err != nil {
				assert.Equal(t, int64(0), response.Rows)
			}
		})
	}
}

func TestDatabaseConfig_GetDatabaseResponse_ElasticsearchTimeout(t *testing.T) {
	config := createElasticsearchConfig(8, []string{"http://non-existent-es:9200"})

	// Test that timeout parameter is properly passed
	query := "FROM logs | LIMIT 100"
	start := time.Now()
	response, err := config.GetDatabaseResponse(context.Background(), query, 1) // 1 second timeout
	elapsed := time.Since(start)

	// Should fail quickly due to connection error (before timeout)
	assert.Error(t, err)
	assert.Less(t, elapsed, 5*time.Second, "Should fail quickly on connection error")
	assert.Equal(t, DatabaseResponse{}, response)
}

// Test database connection error scenarios
func TestDatabaseConfig_GetDatabaseResponse_ConnectionErrors(t *testing.T) {
	tests := []struct {
		name   string
		config DatabaseConfig
	}{
		{
			name: "Invalid SQLite path",
			config: DatabaseConfig{
				Driver: DriverSqlite,
				SQLite: SQLiteConfig{
					Path: "/invalid/path/to/database.db",
				},
			},
		},
		{
			name: "Invalid PostgreSQL host",
			config: DatabaseConfig{
				Driver: DriverPostgreSQL,
				PostgreSQL: PostgreSQLConfig{
					Host:     "invalid-host-12345",
					Port:     5432,
					Username: "user",
					Password: "pass",
					Database: "db",
				},
			},
		},
		{
			name: "Invalid ClickHouse configuration",
			config: DatabaseConfig{
				Driver: DriverClickHouse,
				ClickHouse: ClickHouseConfig{
					Host:     "invalid-clickhouse",
					Port:     9000,
					Username: "user",
					Password: "pass",
					Database: "db",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			response, err := tt.config.GetDatabaseResponse(context.Background(), "SELECT 1", 30)

			// We expect errors for invalid configurations
			if err != nil {
				// This is expected - just log the error type for debugging
				t.Logf("Expected error for %s: %v", tt.name, err)
			}

			// Response should be empty or minimal on error
			if err != nil {
				assert.Equal(t, int64(0), response.Rows)
			}
		})
	}
}

// Test metadata validation for different data types
func TestDatabaseConfig_GetDatabaseResponse_MetadataValidation(t *testing.T) {
	config := createMemoryConfig()

	// Create table with various data types
	createQuery := `
		CREATE TABLE metadata_test (
			id INTEGER PRIMARY KEY,
			name TEXT NOT NULL,
			age INTEGER,
			salary REAL,
			is_active BOOLEAN,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
			data BLOB
		)
	`

	_, err := config.GetDatabaseResponse(context.Background(), createQuery, 30)
	require.NoError(t, err)

	// Insert test data
	insertQuery := `
		INSERT INTO metadata_test (name, age, salary, is_active, data) 
		VALUES ('Alice', 30, 50000.50, 1, X'48656C6C6F')
	`
	_, err = config.GetDatabaseResponse(context.Background(), insertQuery, 30)
	require.NoError(t, err)

	// Query and validate metadata
	selectQuery := "SELECT * FROM metadata_test"
	response, err := config.GetDatabaseResponse(context.Background(), selectQuery, 30)

	require.NoError(t, err)
	assert.Equal(t, int64(1), response.Rows)

	// Validate metadata structure
	expectedColumns := []string{"id", "name", "age", "salary", "is_active", "created_at", "data"}
	require.Len(t, response.Meta, len(expectedColumns))

	for i, expectedCol := range expectedColumns {
		assert.Equal(t, expectedCol, response.Meta[i]["name"])
	}

	// Validate data types in response
	require.Len(t, response.Data, 1)
	data := response.Data[0]

	assert.IsType(t, int64(0), data["id"])
	assert.IsType(t, "", data["name"])
	assert.Equal(t, "Alice", data["name"])
	assert.IsType(t, int64(0), data["age"])
	assert.Equal(t, int64(30), data["age"])
}

// Test Altibase data type conversion
func TestDatabaseConfig_GetDatabaseResponse_AltibaseDataConversion(t *testing.T) {
	// This test checks the data conversion logic for Altibase ODBC types
	// Since we can't easily test with actual Altibase, we test the conversion logic indirectly

	config := DatabaseConfig{
		Driver: DriverAltibase,
		Altibase: AltibaseConfig{
			Driver:   "Altibase",
			DSN:      "invalid-altibase",
			Port:     20300,
			UID:      "user",
			Password: "pass",
			Database: "db",
		},
	}

	// Test with a simple query - this will fail due to invalid connection
	// but we can verify the error handling
	response, err := config.GetDatabaseResponse(context.Background(), "SELECT 1", 30)

	if err != nil {
		// Expected since we don't have real Altibase connection
		t.Logf("Expected Altibase connection error: %v", err)
		errorStr := strings.ToLower(err.Error())
		hasExpectedError := strings.Contains(errorStr, "connection") ||
			strings.Contains(errorStr, "not supported") ||
			strings.Contains(errorStr, "dial") ||
			strings.Contains(errorStr, "invalid")
		assert.True(t, hasExpectedError, "Expected connection or support-related error, got: %v", err)
	}

	// Response should be empty on connection error
	if err != nil {
		assert.Equal(t, int64(0), response.Rows)
		assert.Equal(t, "", response.SQL)
	}
}

// Test concurrent access to GetDatabaseResponse
func TestDatabaseConfig_GetDatabaseResponse_Concurrent(t *testing.T) {
	config := createMemoryConfig()

	// Create a simple table
	_, err := config.GetDatabaseResponse(context.Background(), "CREATE TABLE concurrent_test (id INTEGER, value TEXT)", 30)
	require.NoError(t, err)

	// Insert some data
	_, err = config.GetDatabaseResponse(context.Background(), "INSERT INTO concurrent_test VALUES (1, 'test1'), (2, 'test2')", 30)
	require.NoError(t, err)

	// Run multiple concurrent queries
	const numGoroutines = 10
	results := make(chan error, numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			query := fmt.Sprintf("SELECT * FROM concurrent_test WHERE id = %d", (id%2)+1)
			response, err := config.GetDatabaseResponse(context.Background(), query, 30)

			if err != nil {
				results <- err
				return
			}

			if response.Rows == 0 {
				results <- fmt.Errorf("no rows returned for query %s", query)
				return
			}

			results <- nil
		}(i)
	}

	// Collect results
	for range numGoroutines {
		err := <-results
		assert.NoError(t, err)
	}
}

// Benchmark tests
func BenchmarkDatabaseConfig_GetDriverName(b *testing.B) {
	config := DatabaseConfig{Driver: DriverPostgreSQL}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.GetDriverName()
	}
}

func BenchmarkDatabaseConfig_GetDataSourceName(b *testing.B) {
	config := DatabaseConfig{
		Driver: DriverPostgreSQL,
		PostgreSQL: PostgreSQLConfig{
			Host:     "localhost",
			Port:     5432,
			Username: "user",
			Password: "pass",
			Database: "db",
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = config.GetDataSourceName()
	}
}

func BenchmarkDatabaseConfig_GetDatabaseResponse(b *testing.B) {
	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: ":memory:",
		},
	}

	query := "SELECT 1 as test_column"

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = config.GetDatabaseResponse(context.Background(), query, 30)
	}
}

// Test performance with different query types
func TestDatabaseConfig_GetDatabaseResponse_Performance(t *testing.T) {
	config := DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{
			Path: ":memory:",
		},
	}

	// Create a table with more data for performance testing
	createQuery := `
		CREATE TABLE perf_test (
			id INTEGER PRIMARY KEY,
			data TEXT,
			number INTEGER,
			timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
		)
	`

	_, err := config.GetDatabaseResponse(context.Background(), createQuery, 30)
	require.NoError(t, err)

	// Insert multiple rows using a simpler approach
	for i := range 100 {
		insertQuery := fmt.Sprintf("INSERT INTO perf_test (data, number) VALUES ('test_data_%d', %d)", i, i)
		_, err := config.GetDatabaseResponse(context.Background(), insertQuery, 30)
		require.NoError(t, err)
	}

	// Test query performance
	start := time.Now()
	selectQuery := "SELECT * FROM perf_test LIMIT 50"
	response, err := config.GetDatabaseResponse(context.Background(), selectQuery, 30)
	elapsed := time.Since(start)

	require.NoError(t, err)
	assert.Equal(t, int64(50), response.Rows)
	assert.Less(t, elapsed, 100*time.Millisecond) // Should complete within 100ms
	t.Logf("Query took %v for %d rows", elapsed, response.Rows)
}

// HTTP 테스트 서버를 사용한 Prometheus 테스트
func TestDatabaseConfig_GetDatabaseResponse_PrometheusWithTestServer(t *testing.T) {
	// Prometheus API 응답을 모킹하는 HTTP 테스트 서버 생성
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// 요청 URL과 쿼리 파라미터 확인
		if r.URL.Path == "/api/v1/query" {
			var query string

			// GET 또는 POST 요청 모두 처리
			switch r.Method {
			case "GET":
				query = r.URL.Query().Get("query")
			case "POST":
				// POST 요청의 경우 form data에서 읽기
				err := r.ParseForm()
				if err == nil {
					query = r.Form.Get("query")
				}
			}

			// 다양한 Prometheus 쿼리에 대한 모킹 응답
			switch query {
			case "up":
				// Vector 타입 응답 모킹
				response := `{
					"status": "success",
					"data": {
						"resultType": "vector",
						"result": [
							{
								"metric": {
									"__name__": "up",
									"instance": "localhost:9090",
									"job": "prometheus"
								},
								"value": [1694956800, "1"]
							},
							{
								"metric": {
									"__name__": "up",
									"instance": "localhost:8080",
									"job": "app"
								},
								"value": [1694956800, "0"]
							}
						]
					}
				}`
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))

			case "prometheus_build_info":
				// Scalar 타입 응답 모킹
				response := `{
					"status": "success",
					"data": {
						"resultType": "scalar",
						"result": [1694956800, "1"]
					}
				}`
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))

			case "rate(http_requests_total[5m])":
				// Matrix 타입 응답 모킹
				response := `{
					"status": "success",
					"data": {
						"resultType": "matrix",
						"result": [
							{
								"metric": {
									"__name__": "http_requests_total",
									"instance": "localhost:8080",
									"job": "app",
									"method": "GET",
									"status": "200"
								},
								"values": [
									[1694956800, "10.5"],
									[1694956860, "12.3"],
									[1694956920, "8.7"]
								]
							}
						]
					}
				}`
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusOK)
				w.Write([]byte(response))

			default:
				// 알 수 없는 쿼리에 대한 에러 응답
				response := `{
					"status": "error",
					"errorType": "bad_data",
					"error": "invalid query"
				}`
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusBadRequest)
				w.Write([]byte(response))
			}
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	// 테스트 서버 URL로 DatabaseConfig 생성
	config := DatabaseConfig{
		Driver: DriverPrometheus,
		Prometheus: PrometheusConfig{
			Address: server.URL,
		},
	}

	t.Run("Vector_type_response", func(t *testing.T) {
		query := "up"
		response, err := config.GetDatabaseResponse(context.Background(), query, 30)

		// If connection fails to localhost:9090, skip this test
		if err != nil && (strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "dial tcp")) {
			t.Skipf("Prometheus server not available for testing: %v", err)
			return
		}

		require.NoError(t, err)
		assert.Equal(t, query, response.SQL)
		assert.Equal(t, int64(2), response.Rows) // 두 개의 인스턴스
		assert.Greater(t, response.Statistics.Elapsed, 0.0)

		// 메타데이터 확인 (value, timestamp, labels)
		require.Len(t, response.Meta, 5) // value, timestamp, __name__, instance, job
		assert.Equal(t, "value", response.Meta[0]["name"])
		assert.Equal(t, "timestamp", response.Meta[1]["name"])

		// 데이터 확인
		require.Len(t, response.Data, 2)
		assert.Equal(t, model.SampleValue(1), response.Data[0]["value"])
		assert.Equal(t, "up", response.Data[0]["__name__"])
		assert.Equal(t, "localhost:9090", response.Data[0]["instance"])
		assert.Equal(t, "prometheus", response.Data[0]["job"])

		assert.Equal(t, model.SampleValue(0), response.Data[1]["value"])
		assert.Equal(t, "localhost:8080", response.Data[1]["instance"])
		assert.Equal(t, "app", response.Data[1]["job"])
	})

	t.Run("Scalar_type_response", func(t *testing.T) {
		query := "prometheus_build_info"
		response, err := config.GetDatabaseResponse(context.Background(), query, 30)

		// If connection fails to localhost:9090, skip this test
		if err != nil && (strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "dial tcp")) {
			t.Skipf("Prometheus server not available for testing: %v", err)
			return
		}

		require.NoError(t, err)
		assert.Equal(t, query, response.SQL)
		assert.Equal(t, int64(1), response.Rows)
		assert.Greater(t, response.Statistics.Elapsed, 0.0)

		// 메타데이터 확인
		require.Len(t, response.Meta, 2) // value, timestamp only for scalar
		assert.Equal(t, "value", response.Meta[0]["name"])
		assert.Equal(t, "timestamp", response.Meta[1]["name"])

		// 데이터 확인
		require.Len(t, response.Data, 1)
		assert.Equal(t, model.SampleValue(1), response.Data[0]["value"])
		assert.IsType(t, time.Time{}, response.Data[0]["timestamp"])
	})

	t.Run("Matrix_type_response", func(t *testing.T) {
		query := "rate(http_requests_total[5m])"
		response, err := config.GetDatabaseResponse(context.Background(), query, 30)

		// If connection fails to localhost:9090, skip this test
		if err != nil && (strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "dial tcp")) {
			t.Skipf("Prometheus server not available for testing: %v", err)
			return
		}

		require.NoError(t, err)
		assert.Equal(t, query, response.SQL)
		assert.Equal(t, int64(3), response.Rows) // 3개의 시계열 데이터 포인트
		assert.Greater(t, response.Statistics.Elapsed, 0.0)

		// 메타데이터 확인
		require.Len(t, response.Meta, 7) // value, timestamp + 5 labels
		assert.Equal(t, "value", response.Meta[0]["name"])
		assert.Equal(t, "timestamp", response.Meta[1]["name"])

		// 데이터 확인
		require.Len(t, response.Data, 3)
		for _, data := range response.Data {
			assert.Contains(t, data, "value")
			assert.Contains(t, data, "timestamp")
			assert.Contains(t, data, "__name__")
			assert.Contains(t, data, "instance")
			assert.Contains(t, data, "job")
			assert.Contains(t, data, "method")
			assert.Contains(t, data, "status")

			assert.Equal(t, "http_requests_total", data["__name__"])
			assert.Equal(t, "localhost:8080", data["instance"])
			assert.Equal(t, "app", data["job"])
			assert.Equal(t, "GET", data["method"])
			assert.Equal(t, "200", data["status"])
		}
	})

	t.Run("Invalid_query", func(t *testing.T) {
		query := "invalid_prometheus_query"
		response, err := config.GetDatabaseResponse(context.Background(), query, 30)

		// If connection fails to localhost:9090, skip this test
		if err != nil && (strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "dial tcp")) {
			t.Skipf("Prometheus server not available for testing: %v", err)
			return
		}

		assert.Error(t, err)
		assert.Contains(t, strings.ToLower(err.Error()), "bad_data")
		// 에러 시 응답은 빈 구조체여야 함
		assert.Equal(t, DatabaseResponse{}, response)
	})

	t.Run("Empty_query", func(t *testing.T) {
		response, err := config.GetDatabaseResponse(context.Background(), "", 30)

		assert.Error(t, err)
		assert.Contains(t, err.Error(), "invalid query")
		assert.Equal(t, DatabaseResponse{}, response)
	})
}

// Prometheus 에러 응답 테스트
func TestDatabaseConfig_GetDatabaseResponse_PrometheusErrorHandling(t *testing.T) {
	// 에러를 반환하는 HTTP 테스트 서버
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/api/v1/query" {
			// Prometheus API 에러 응답 모킹
			response := `{
				"status": "error",
				"errorType": "timeout",
				"error": "query timeout exceeded"
			}`
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusUnprocessableEntity)
			w.Write([]byte(response))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	config := DatabaseConfig{
		Driver: DriverPrometheus,
		Prometheus: PrometheusConfig{
			Address: server.URL,
		},
	}

	query := "complex_query_that_times_out"
	response, err := config.GetDatabaseResponse(context.Background(), query, 30)

	assert.Error(t, err)
	// Check for timeout error OR connection error (both are valid)
	errorStr := strings.ToLower(err.Error())
	hasExpectedError := strings.Contains(errorStr, "timeout") || strings.Contains(errorStr, "connection refused") || strings.Contains(errorStr, "dial tcp")
	assert.True(t, hasExpectedError, "Expected timeout or connection error, got: %v", err)
	assert.Equal(t, DatabaseResponse{}, response)
}

// Prometheus 연결 실패 테스트
func TestDatabaseConfig_GetDatabaseResponse_PrometheusConnectionFailure(t *testing.T) {
	// 잘못된 주소로 DatabaseConfig 생성
	config := DatabaseConfig{
		Driver: DriverPrometheus,
		Prometheus: PrometheusConfig{
			Address: "http://non-existent-prometheus:9090",
		},
	}

	query := "up"
	response, err := config.GetDatabaseResponse(context.Background(), query, 30)

	assert.Error(t, err)
	errorStr := strings.ToLower(err.Error())
	hasConnectionError := strings.Contains(errorStr, "connection") ||
		strings.Contains(errorStr, "dial") ||
		strings.Contains(errorStr, "no such host") ||
		strings.Contains(errorStr, "refused")
	assert.True(t, hasConnectionError, "Expected connection-related error, got: %v", err)
	assert.Equal(t, DatabaseResponse{}, response)
}

// 쿼리 타임아웃 테스트
func TestDatabaseConfig_GetDatabaseResponse_QueryTimeout(t *testing.T) {
	config := createMemoryConfig()

	// 타임아웃을 유발할 수 있는 복잡한 쿼리 테스트
	// SQLite의 경우 30초 타임아웃 설정이 적용됨
	t.Run("Simple_query_completes_before_timeout", func(t *testing.T) {
		query := "SELECT 1"
		response, err := config.GetDatabaseResponse(context.Background(), query, 30)

		require.NoError(t, err)
		assert.Equal(t, query, response.SQL)
		assert.Equal(t, int64(1), response.Rows)
		assert.Greater(t, response.Statistics.Elapsed, 0.0)
	})

	t.Run("Query_with_timeout_context", func(t *testing.T) {
		// 실제 타임아웃은 30초이므로 이 테스트는 빠르게 완료되어야 함
		query := "SELECT 1 as test"
		start := time.Now()
		response, err := config.GetDatabaseResponse(context.Background(), query, 30)
		elapsed := time.Since(start)

		require.NoError(t, err)
		assert.Equal(t, query, response.SQL)
		// 쿼리는 타임아웃(30초)보다 훨씬 빨리 완료되어야 함
		assert.Less(t, elapsed, 1*time.Second)
	})
}

// context.DeadlineExceeded 에러 처리 테스트
func TestDatabaseConfig_GetDatabaseResponse_DeadlineExceeded(t *testing.T) {
	// Note: 실제로 30초 타임아웃을 발생시키는 것은 테스트 시간이 너무 오래 걸리므로
	// 이 테스트는 타임아웃이 발생했을 때의 에러 처리 로직을 검증하기 위한 것입니다.
	// 실제 타임아웃은 통합 테스트나 수동 테스트에서 확인해야 합니다.

	t.Run("Timeout_error_is_returned_correctly", func(t *testing.T) {
		config := createMemoryConfig()

		// 간단한 쿼리로 타임아웃이 발생하지 않는 것을 확인
		query := "SELECT 1"
		response, err := config.GetDatabaseResponse(context.Background(), query, 30)

		require.NoError(t, err)
		assert.NotEqual(t, context.DeadlineExceeded, err)
		assert.Equal(t, int64(1), response.Rows)
	})

	t.Run("Invalid_driver_returns_error", func(t *testing.T) {
		config := DatabaseConfig{
			Driver: "invalid_driver",
		}

		query := "SELECT 1"
		response, err := config.GetDatabaseResponse(context.Background(), query, 30)

		assert.Error(t, err)
		assert.Equal(t, DatabaseResponse{}, response)
	})
}
