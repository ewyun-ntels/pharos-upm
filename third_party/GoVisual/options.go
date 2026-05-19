package GoVisual

import (
	"database/sql"
	"fmt"
	"path/filepath"

	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/third_party/GoVisual/internal/store"
)

type Config struct {
	MaxRequests int

	LogRequestBody bool

	LogResponseBody bool

	IgnorePaths []string

	ServiceName string

	ServiceVersion string

	OTelEndpoint string

	// Storage configuration
	StorageType store.StorageType

	// Connection string for database stores
	ConnectionString string

	// TableName for SQL database stores
	TableName string

	// TTL for Redis store in seconds
	RedisTTL int

	// Existing database connection for SQLite
	ExistingDB *sql.DB

	// ClickHouse specific configuration
	ClickHouseConf *common.Config
}

// Option is a function that modifies the configuration
type Option func(*Config)

// WithMaxRequests sets the maximum number of requests to store
func WithMaxRequests(max int) Option {
	return func(c *Config) {
		c.MaxRequests = max
	}
}

// WithRequestBodyLogging enables or disables request body logging
func WithRequestBodyLogging(enabled bool) Option {
	return func(c *Config) {
		c.LogRequestBody = enabled
	}
}

// WithResponseBodyLogging enables or disables response body logging
func WithResponseBodyLogging(enabled bool) Option {
	return func(c *Config) {
		c.LogResponseBody = enabled
	}
}

// WithIgnorePaths sets the path patterns to ignore
func WithIgnorePaths(patterns ...string) Option {
	return func(c *Config) {
		c.IgnorePaths = append(c.IgnorePaths, patterns...)
	}
}

// WithServiceName sets the service name for OpenTelemetry
func WithServiceName(name string) Option {
	return func(c *Config) {
		c.ServiceName = name
	}
}

// WithServiceVersion sets the service version for OpenTelemetry
func WithServiceVersion(version string) Option {
	return func(c *Config) {
		c.ServiceVersion = version
	}
}

// WithOTelEndpoint sets the OTLP endpoint for exporting telemetry data
func WithOTelEndpoint(endpoint string) Option {
	return func(c *Config) {
		c.OTelEndpoint = endpoint
	}
}

// WithMemoryStorage configures the application to use in-memory storage
func WithMemoryStorage() Option {
	return func(c *Config) {
		c.StorageType = store.StorageTypeMemory
	}
}

// WithPostgresStorage configures the application to use PostgreSQL storage
func WithPostgresStorage(connStr string, tableName string) Option {
	return func(c *Config) {
		c.StorageType = store.StorageTypePostgres
		c.ConnectionString = connStr
		c.TableName = tableName
	}
}

// WithSQLiteStorage configures the application to use SQLite storage
func WithSQLiteStorage(dbPath string, tableName string) Option {
	return func(c *Config) {
		c.StorageType = store.StorageTypeSQLite
		c.ConnectionString = dbPath
		c.TableName = tableName
	}
}

// WithSQLiteStorageDB configures the application to use SQLite storage with an existing database connection
func WithSQLiteStorageDB(db *sql.DB, tableName string) Option {
	return func(c *Config) {
		c.StorageType = store.StorageTypeSQLiteWithDB
		c.ExistingDB = db
		c.TableName = tableName
	}
}

// WithRedisStorage configures the application to use Redis storage
func WithRedisStorage(connStr string, ttlSeconds int) Option {
	return func(c *Config) {
		c.StorageType = store.StorageTypeRedis
		c.ConnectionString = connStr
		c.RedisTTL = ttlSeconds
	}
}

// WithMongoDBStorage configures the application to use MongoDB storage
func WithMongoDBStorage(uri, databaseName, collectionName string) Option {
	return func(c *Config) {
		c.StorageType = store.StorageTypeMongoDB
		c.ConnectionString = uri
		c.TableName = fmt.Sprintf("%s.%s", databaseName, collectionName)
	}
}

// ShouldIgnorePath checks if a path should be ignored based on the configured patterns
// ShouldIgnorePath checks if a path should be ignored based on the configured patterns
func (c *Config) ShouldIgnorePath(path string) bool {
	// Then check against provided ignore patterns
	for _, pattern := range c.IgnorePaths {
		matched, err := filepath.Match(pattern, path)
		if err == nil && matched {
			return true
		}

		// Special handling for path groups with trailing slash
		if len(pattern) > 0 && pattern[len(pattern)-1] == '/' {
			// If pattern ends with /, check if path starts with pattern
			if len(path) >= len(pattern) && path[:len(pattern)] == pattern {
				return true
			}
		}
	}

	return false
}

// defaultConfig returns the default configuration
func defaultConfig() *Config {
	return &Config{
		MaxRequests:     100,
		LogRequestBody:  false,
		LogResponseBody: false,
		IgnorePaths:     []string{},
		ServiceName:     "tarzan-history",
		ServiceVersion:  "dev",
		OTelEndpoint:    "localhost:4317",
		StorageType:     store.StorageTypeMemory,
		TableName:       "history_requests",
		RedisTTL:        86400, // 24 hours
	}
}

// WithClickHouseStorage configures the application to use ClickHouse storage
func WithClickHouseStorage(config *common.Config) Option {
	return func(c *Config) {
		c.StorageType = store.StorageTypeClickHouse
		c.ClickHouseConf = config
	}
}
