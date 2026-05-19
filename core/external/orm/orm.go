package orm

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

const (
	DriverDefault       = ""
	DriverPostgreSQL    = "postgresql"
	DriverSqlite        = "sqlite"
	DriverClickHouse    = "clickhouse"
	DriverAltibase      = "altibase"
	DriverVertica       = "vertica"
	DriverPrometheus    = "prometheus"
	DriverElasticsearch = "elasticsearch"
)

var DriverName = map[string]string{
	DriverAltibase:      "odbc",
	DriverClickHouse:    "clickhouse",
	DriverPostgreSQL:    "postgres",
	DriverSqlite:        "sqlite",
	DriverVertica:       "vertica",
	DriverPrometheus:    "prometheus",
	DriverElasticsearch: "elasticsearch",
}

type DatabaseResponseStatistics struct {
	Elapsed float64 `json:"elapsed"`
}

type DatabaseResponse struct {
	Meta       []map[string]string        `json:"meta"`
	Data       []map[string]any           `json:"data"`
	Rows       int64                      `json:"rows"`
	SQL        string                     `json:"sql,omitempty"`
	Statistics DatabaseResponseStatistics `json:"statistics"`
}

type DBHandlerFunc func(db *sqlx.DB) error

// Handler driver가 "" 일경우 config에 설정되어 있는 driver 사용 (default pool)
func Handler(driver string, config *DatabaseConfig, handler DBHandlerFunc) error {
	return HandlerWithRole(PoolDefault, driver, config, handler)
}

// HandlerWithRole allows selecting a specific pool role (default/statistics/service)
func HandlerWithRole(role string, driver string, config *DatabaseConfig, handler DBHandlerFunc) error {
	setDriver := driver
	if setDriver == DriverDefault {
		setDriver = config.Driver
	}

	databaseConfig := *config
	databaseConfig.Driver = setDriver

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

	db, err := pool.GetDB(databaseConfig)
	if err != nil {
		return err
	}
	return handler(db)
}

// StatisticsHandler is a convenience wrapper for the statistics pool
func StatisticsHandler(driver string, config *DatabaseConfig, handler DBHandlerFunc) error {
	return HandlerWithRole(PoolStatistics, driver, config, handler)
}

// ServiceHandler is a convenience wrapper for the agent service pool
//func ServiceHandler(driver string, config *DatabaseConfig, handler DBHandlerFunc) error {
//	return HandlerWithRole(PoolService, driver, config, handler)
//}
