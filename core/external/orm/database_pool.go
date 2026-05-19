package orm

import (
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "github.com/ClickHouse/clickhouse-go/v2"
	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	prometheus_api "github.com/prometheus/client_golang/api"
	prometheus_api_v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	_ "github.com/vertica/vertica-sql-go"
	_ "modernc.org/sqlite"
	"ntels.com/pharos/core/external"
)

var supportDriver = map[string]bool{
	DriverAltibase:      false,
	DriverSqlite:        true,
	DriverPostgreSQL:    true,
	DriverClickHouse:    true,
	DriverVertica:       true,
	DriverPrometheus:    true,
	DriverElasticsearch: true,
}

type databasePool struct {
	mutex                sync.Mutex
	dbs                  map[string]*sqlx.DB
	prometheusApis       map[string]prometheus_api_v1.API
	elasticsearchClients map[string]*ElasticsearchClient
}

func (databasePool *databasePool) GetDB(databaseConfig DatabaseConfig) (*sqlx.DB, error) {
	if !supportDriver[databaseConfig.Driver] {
		return nil, external.ErrorNotSupported
	}

	key := databaseConfig.GetDataSourceName()

	// Fast path: return cached connection without slow work
	databasePool.mutex.Lock()
	if databasePool.dbs == nil {
		databasePool.dbs = map[string]*sqlx.DB{}
	}
	if db, exist := databasePool.dbs[key]; exist {
		databasePool.mutex.Unlock()
		return db, nil
	}
	databasePool.mutex.Unlock()

	// Slow path: establish connection WITHOUT holding mutex.
	// sqlx.Connect calls db.Ping() which may block for TCP timeout (~30s)
	// when the host is unreachable. Releasing the mutex here prevents
	// blocking unrelated database operations (e.g. pharos internal SQLite).
	db, err := databasePool.getSQLxDB(databaseConfig)
	if err != nil {
		return nil, err
	}

	switch databaseConfig.Driver {
	case DriverSqlite:
		db.SetMaxOpenConns(1) // "Update error: database is locked (5) (SQLITE_BUSY)"
	case DriverPostgreSQL:
		db.SetMaxOpenConns(databaseConfig.PostgreSQL.MaxOpenConnection)
		db.SetMaxIdleConns(databaseConfig.PostgreSQL.MaxOpenConnection)
		db.SetConnMaxLifetime(time.Duration(databaseConfig.PostgreSQL.MaxLifetime) * time.Second)
	case DriverClickHouse:
		db.SetMaxOpenConns(databaseConfig.ClickHouse.MaxOpenConnection)
		db.SetMaxIdleConns(databaseConfig.ClickHouse.MaxOpenConnection)
		db.SetConnMaxLifetime(time.Duration(databaseConfig.ClickHouse.MaxLifetime) * time.Second)
	case DriverAltibase:
		db.SetMaxOpenConns(databaseConfig.Altibase.MaxOpenConnection)
		db.SetMaxIdleConns(databaseConfig.Altibase.MaxOpenConnection)
		db.SetConnMaxLifetime(time.Duration(databaseConfig.Altibase.MaxLifetime) * time.Second)
	case DriverVertica:
		db.SetMaxOpenConns(databaseConfig.Vertica.MaxOpenConnection)
		db.SetMaxIdleConns(databaseConfig.Vertica.MaxOpenConnection)
		db.SetConnMaxLifetime(time.Duration(databaseConfig.Vertica.MaxLifetime) * time.Second)
	}

	// Re-acquire mutex to store in pool. If another goroutine connected
	// concurrently, use the existing connection and discard ours.
	databasePool.mutex.Lock()
	defer databasePool.mutex.Unlock()
	if existing, ok := databasePool.dbs[key]; ok {
		_ = db.Close()
		return existing, nil
	}
	databasePool.dbs[key] = db
	return db, nil
}

func (databasePool *databasePool) GetElasticsearchClient(databaseConfig DatabaseConfig) (*ElasticsearchClient, error) {
	if !supportDriver[databaseConfig.Driver] {
		return nil, external.ErrorNotSupported
	}
	if databaseConfig.Driver != DriverElasticsearch {
		return nil, external.ErrorNotSupported
	}

	key := databaseConfig.GetDataSourceName()

	// Fast path: return cached client
	databasePool.mutex.Lock()
	if databasePool.elasticsearchClients == nil {
		databasePool.elasticsearchClients = map[string]*ElasticsearchClient{}
	}
	if client, exist := databasePool.elasticsearchClients[key]; exist {
		databasePool.mutex.Unlock()
		return client, nil
	}
	databasePool.mutex.Unlock()

	// Slow path: establish connection WITHOUT holding mutex
	elasticsearchClient, err := NewElasticsearchClient(databaseConfig.Elasticsearch)
	if err != nil {
		return nil, err
	}

	// Re-acquire mutex to store in pool
	databasePool.mutex.Lock()
	defer databasePool.mutex.Unlock()
	if existing, ok := databasePool.elasticsearchClients[key]; ok {
		return existing, nil
	}
	databasePool.elasticsearchClients[key] = elasticsearchClient
	return elasticsearchClient, nil
}

func (databasePool *databasePool) GetPrometheusAPI(databaseConfig DatabaseConfig) (prometheus_api_v1.API, error) {
	if !supportDriver[databaseConfig.Driver] {
		return nil, external.ErrorNotSupported
	}
	if databaseConfig.Driver != DriverPrometheus {
		return nil, external.ErrorNotSupported
	}

	key := databaseConfig.GetDataSourceName()

	// Fast path: return cached API
	databasePool.mutex.Lock()
	if databasePool.prometheusApis == nil {
		databasePool.prometheusApis = map[string]prometheus_api_v1.API{}
	}
	if api, exist := databasePool.prometheusApis[key]; exist {
		databasePool.mutex.Unlock()
		return api, nil
	}
	databasePool.mutex.Unlock()

	// Slow path: establish connection WITHOUT holding mutex
	client, err := prometheus_api.NewClient(prometheus_api.Config{
		Address: databaseConfig.Prometheus.Address,
	})
	if err != nil {
		return nil, err
	}

	api := prometheus_api_v1.NewAPI(client)

	// Re-acquire mutex to store in pool
	databasePool.mutex.Lock()
	defer databasePool.mutex.Unlock()
	if existing, ok := databasePool.prometheusApis[key]; ok {
		return existing, nil
	}
	databasePool.prometheusApis[key] = api
	return api, nil
}

func (databasePool *databasePool) Remove(databaseConfig DatabaseConfig) error {
	databasePool.mutex.Lock()
	defer databasePool.mutex.Unlock()

	switch databaseConfig.Driver {
	case DriverPrometheus:
		if databasePool.prometheusApis == nil {
			return nil
		}
		if _, exist := databasePool.prometheusApis[databaseConfig.GetDataSourceName()]; !exist {
			return nil
		}

		delete(databasePool.prometheusApis, databaseConfig.GetDataSourceName())
		return nil
	case DriverElasticsearch:
		if databasePool.elasticsearchClients == nil {
			return nil
		}
		if _, exist := databasePool.elasticsearchClients[databaseConfig.GetDataSourceName()]; !exist {
			return nil
		}

		delete(databasePool.elasticsearchClients, databaseConfig.GetDataSourceName())
		return nil
	default:
		if databasePool.dbs == nil {
			return nil
		}
		if _, exist := databasePool.dbs[databaseConfig.GetDataSourceName()]; !exist {
			return nil
		}

		db := databasePool.dbs[databaseConfig.GetDataSourceName()]
		delete(databasePool.dbs, databaseConfig.GetDataSourceName())

		return db.Close()
	}
}

func (databasePool *databasePool) RemoveAll() {
	databasePool.mutex.Lock()
	defer databasePool.mutex.Unlock()

	for _, db := range databasePool.dbs {
		if err := db.Close(); err != nil {
			slog.Error("database error", "error", err.Error())
		}
	}
	clear(databasePool.dbs)

	clear(databasePool.prometheusApis)
	clear(databasePool.elasticsearchClients)
}

func (databasePool *databasePool) getSQLxDB(databaseConfig DatabaseConfig) (*sqlx.DB, error) {
	switch databaseConfig.Driver {
	case DriverSqlite:
		path := filepath.Dir(databaseConfig.SQLite.Path)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0o750); err != nil {
				return nil, err
			}
		}
	}

	return sqlx.Connect(databaseConfig.GetDriverName(), databaseConfig.GetDataSourceName())
}
