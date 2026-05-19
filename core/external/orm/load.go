package orm

import (
	"fmt"
)

// Pool role names to separate connection pools per domain
const (
	PoolDefault    = "default"
	PoolStatistics = "statistics"
	PoolService    = "service"
)

var DatabasePool = databasePool{}
var StatisticsDatabasePool = databasePool{}
var ServiceDatabasePool = databasePool{}

// Load keeps backward compatibility: uses the default pool
func Load(config DatabaseConfig) error {
	return LoadWithRole(PoolDefault, config)
}

// LoadWithRole initializes a connection in a specific pool role
func LoadWithRole(role string, config DatabaseConfig) error {
	var pool *databasePool
	switch role {
	case PoolDefault:
		pool = &DatabasePool
	case PoolStatistics:
		pool = &StatisticsDatabasePool
	case PoolService:
		pool = &ServiceDatabasePool
	default:
		return fmt.Errorf("unknown pool role: %s", role)
	}

	if _, err := pool.GetDB(config); err != nil {
		return err
	}
	return nil
}

func Unload() {
	DatabasePool.RemoveAll()
	StatisticsDatabasePool.RemoveAll()
	ServiceDatabasePool.RemoveAll()
}
