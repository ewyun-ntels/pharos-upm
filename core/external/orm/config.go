package orm

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/prometheus/common/model"
	// ewyun-20260514: QueryRange를 위해 prometheus v1 API 타입 추가
	prometheus_api_v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"ntels.com/pharos/core/external"
)

// ewyun-20260514: Prometheus QueryRange context key 및 시간 범위 타입
type prometheusTimeRangeKey struct{}

// PrometheusTimeRange 대시보드 시간 범위를 Prometheus QueryRange에 전달하기 위한 구조체
type PrometheusTimeRange struct {
	StartTime int64 // Unix 초
	EndTime   int64 // Unix 초
	Step      int   // 초 단위 step
}

// WithPrometheusTimeRange context에 Prometheus 시간 범위를 주입
func WithPrometheusTimeRange(ctx context.Context, tr PrometheusTimeRange) context.Context {
	return context.WithValue(ctx, prometheusTimeRangeKey{}, tr)
}

func getPrometheusTimeRange(ctx context.Context) (PrometheusTimeRange, bool) {
	tr, ok := ctx.Value(prometheusTimeRangeKey{}).(PrometheusTimeRange)
	return tr, ok
}

type SQLiteConfig struct {
	Path string `mapstructure:"path"`

	Backup struct {
		Use       bool   `mapstructure:"use"`
		Schedule  string `mapstructure:"schedule"`
		Directory string `mapstructure:"directory"`
		TTL       string `mapstructure:"ttl"`
	} `mapstructure:"backup"`
}

type PostgreSQLConfig struct {
	Host              string `mapstructure:"host"`
	Port              int    `mapstructure:"port"`
	Username          string `mapstructure:"username"`
	Password          string `mapstructure:"password"`
	Database          string `mapstructure:"database"`
	Version           string `mapstructure:"version"`
	MaxOpenConnection int    `mapstructure:"max_open_connection" json:"max_open_connection"`
	MaxLifetime       int    `mapstructure:"max_lifetime" json:"max_lifetime"`
}

func (pgConfig *PostgreSQLConfig) GetMajorVersion() int {
	if pgConfig.Version == "" {
		return 0
	}

	if major, err := strconv.Atoi(strings.SplitN(pgConfig.Version, ".", 2)[0]); err != nil {
		return 0
	} else {
		return major
	}
}

type ClickHouseConfig struct {
	Host              string `mapstructure:"host"`
	Port              int    `mapstructure:"port"`
	Username          string `mapstructure:"username"`
	Password          string `mapstructure:"password"`
	Database          string `mapstructure:"database"`
	MaxOpenConnection int    `mapstructure:"max_open_connection" json:"max_open_connection"`
	MaxLifetime       int    `mapstructure:"max_lifetime" json:"max_lifetime"`
}

type AltibaseConfig struct {
	Driver            string `mapstructure:"driver"`
	DSN               string `mapstructure:"dsn"`
	Port              int    `mapstructure:"port"`
	UID               string `mapstructure:"uid"`
	Password          string `mapstructure:"password"`
	Database          string `mapstructure:"database"`
	MaxOpenConnection int    `mapstructure:"max_open_connection" json:"max_open_connection"`
	MaxLifetime       int    `mapstructure:"max_lifetime" json:"max_lifetime"`
}

type VerticaConfig struct {
	Host              string `mapstructure:"host"`
	Port              int    `mapstructure:"port"`
	Username          string `mapstructure:"username"`
	Password          string `mapstructure:"password"`
	Database          string `mapstructure:"database"`
	MaxOpenConnection int    `mapstructure:"max_open_connection" json:"max_open_connection"`
	MaxLifetime       int    `mapstructure:"max_lifetime" json:"max_lifetime"`
}

type PrometheusConfig struct {
	Address string `mapstructure:"address"`
}

type ElasticsearchConfig struct {
	Version   int      `mapstructure:"version"`
	Addresses []string `mapstructure:"addresses"`
	Bulk      struct {
		FlushBytes    int    `mapstructure:"flush_bytes"`
		FlushInterval string `mapstructure:"flush_interval"`
	} `mapstructure:"bulk"`
}

type DatabaseConfig struct {
	Driver string `mapstructure:"driver"`

	Migration struct {
		ClickHouse struct {
			ClusterName string `mapstructure:"cluster_name"`
		} `mapstructure:"clickhouse"`
	} `mapstructure:"migration"`

	SQLite        SQLiteConfig        `mapstructure:"sqlite"`
	PostgreSQL    PostgreSQLConfig    `mapstructure:"postgresql"`
	ClickHouse    ClickHouseConfig    `mapstructure:"clickhouse"`
	Altibase      AltibaseConfig      `mapstructure:"altibase"`
	Vertica       VerticaConfig       `mapstructure:"vertica"`
	Prometheus    PrometheusConfig    `mapstructure:"prometheus"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
}

func (databaseConfig *DatabaseConfig) GetDriverName() string {
	if _, exist := DriverName[databaseConfig.Driver]; !exist {
		return databaseConfig.Driver
	}

	return DriverName[databaseConfig.Driver]
}

func (databaseConfig *DatabaseConfig) GetDataSourceName() string {
	switch databaseConfig.Driver {
	case DriverSqlite:
		return databaseConfig.SQLite.Path
	case DriverPostgreSQL:
		return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable application_name=pharos",
			databaseConfig.PostgreSQL.Host,
			databaseConfig.PostgreSQL.Port,
			databaseConfig.PostgreSQL.Username,
			databaseConfig.PostgreSQL.Password,
			databaseConfig.PostgreSQL.Database,
		)
	case DriverClickHouse:
		return fmt.Sprintf("clickhouse://%s:%s@%s:%d/%s",
			databaseConfig.ClickHouse.Username,
			databaseConfig.ClickHouse.Password,
			databaseConfig.ClickHouse.Host,
			databaseConfig.ClickHouse.Port,
			databaseConfig.ClickHouse.Database,
		)
	case DriverAltibase:
		return fmt.Sprintf("DRIVER=%s;DSN=%s;CONNTYPE=1;PORT_NO=%d;UID=%s;PWD=%s;DATABASE=%s",
			databaseConfig.Altibase.Driver,
			databaseConfig.Altibase.DSN,
			databaseConfig.Altibase.Port,
			databaseConfig.Altibase.UID,
			databaseConfig.Altibase.Password,
			databaseConfig.Altibase.Database,
		)
	case DriverVertica:
		return fmt.Sprintf("vertica://%s:%s@%s:%d/%s",
			databaseConfig.Vertica.Username,
			databaseConfig.Vertica.Password,
			databaseConfig.Vertica.Host,
			databaseConfig.Vertica.Port,
			databaseConfig.Vertica.Database,
		)
	case DriverPrometheus:
		return databaseConfig.Prometheus.Address
	case DriverElasticsearch:
		return strings.Join(databaseConfig.Elasticsearch.Addresses, ",")
	default:
		return "not implemented database driver"
	}
}

func (databaseConfig *DatabaseConfig) GetDatabaseResponse(ctx context.Context, query string, timeout int) (DatabaseResponse, error) {
	if len(query) == 0 {
		return DatabaseResponse{}, external.ErrorInvalidQuery
	}

	switch databaseConfig.Driver {
	case DriverPrometheus:
		return databaseConfig.makeDatabaseResponseForPrometheus(ctx, query, timeout)
	case DriverElasticsearch:
		return databaseConfig.makeDatabaseResponseForElasticsearch(ctx, query, timeout)
	default:
		return databaseConfig.makeDatabaseResponseForSqlx(ctx, query, timeout)
	}
}

func (databaseConfig *DatabaseConfig) makeDatabaseResponseForSqlx(ctx context.Context, query string, timeout int) (DatabaseResponse, error) {
	response := DatabaseResponse{}

	db, err := DatabasePool.GetDB(*databaseConfig)
	if err != nil {
		return response, err
	}

	convertData := func(columnTypes []*sql.ColumnType, data map[string]any) {
		switch databaseConfig.Driver {
		case DriverAltibase:
			for _, columnType := range columnTypes {
				if _, exist := data[columnType.Name()]; !exist {
					break
				}

				switch columnType.DatabaseTypeName() {
				case "ODBC_SQL_CHAR":
					fallthrough
				case "ODBC_SQL_VARCHAR":
					fallthrough
				case "ODBC_SQL_LONGVARCHAR":
					fallthrough
				case "ODBC_SQL_WCHAR":
					fallthrough
				case "ODBC_SQL_WVARCHAR":
					fallthrough
				case "ODBC_SQL_WLONGVARCHAR":
					switch v := data[columnType.Name()].(type) {
					case []uint8:
						data[columnType.Name()] = string(v)
					default:
						slog.Info("Need to check if type conversion is needed",
							"name", columnType.Name(),
							"database_type", columnType.DatabaseTypeName(),
							"type", v,
							"value", data[columnType.Name()],
							"query", query)
					}
				}
			}
		}
	}

	handler := func(db *sqlx.DB) error {
		start := time.Now()

		queryCtx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
		defer cancel()

		rows, err := db.QueryxContext(queryCtx, query)
		if errors.Is(err, context.DeadlineExceeded) {
			return context.DeadlineExceeded
		} else if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		columnTypes, err := rows.ColumnTypes()
		if err != nil {
			return err
		}

		for _, columnType := range columnTypes {
			response.Meta = append(response.Meta,
				map[string]string{
					"name": columnType.Name(),
				})
		}

		for rows.Next() {
			data := make(map[string]any)
			if err := rows.MapScan(data); err != nil {
				return err
			}

			convertData(columnTypes, data)

			response.Data = append(response.Data, data)
		}

		if err := rows.Err(); err != nil {
			return err
		}

		response.Rows = int64(len(response.Data))
		response.SQL = query
		response.Statistics.Elapsed = time.Since(start).Seconds()

		return nil
	}

	if err := handler(db); err != nil {
		return response, err
	}

	return response, nil
}

func (databaseConfig *DatabaseConfig) makeDatabaseResponseForPrometheus(ctx context.Context, query string, timeout int) (DatabaseResponse, error) {
	response := DatabaseResponse{}

	v1api, err := DatabasePool.GetPrometheusAPI(*databaseConfig)
	if err != nil {
		return response, err
	}

	ctx, cancel := context.WithTimeout(ctx, time.Duration(timeout)*time.Second)
	defer cancel()

	start := time.Now()

	// context에 시간 범위가 있으면 QueryRange, 없으면 기존 Query(instant) 사용
	var (
		result   model.Value
		warnings prometheus_api_v1.Warnings
	)

	if tr, ok := getPrometheusTimeRange(ctx); ok {
		r := prometheus_api_v1.Range{
			Start: time.Unix(tr.StartTime, 0),
			End:   time.Unix(tr.EndTime, 0),
			Step:  time.Duration(tr.Step) * time.Second,
		}
		result, warnings, err = v1api.QueryRange(ctx, query, r)
	} else {
		result, warnings, err = v1api.Query(ctx, query, time.Now())
	}

	if err != nil {
		return response, err
	}
	if len(warnings) > 0 {
		slog.Warn("prometheus query warnings", "warnings", warnings)
	}

	response.Statistics.Elapsed = time.Since(start).Seconds()
	response.SQL = query

	switch v := result.(type) {
	case *model.Scalar:
		response.Rows = 1
		response.Meta = []map[string]string{
			{
				"name": "value",
			},
			{
				"name": "timestamp",
			},
		}
		response.Data = append(response.Data, map[string]any{
			"value":     v.Value,
			"timestamp": v.Timestamp.Time().UTC(),
		})
	case model.Vector:
		if len(v) > 0 {
			response.Meta = []map[string]string{
				{
					"name": "value",
				},
				{
					"name": "timestamp",
				},
			}

			for labelName := range v[0].Metric {
				response.Meta = append(response.Meta, map[string]string{
					"name": string(labelName),
				})
			}
		}

		for _, sample := range v {
			data := map[string]any{
				"value":     sample.Value,
				"timestamp": sample.Timestamp.Time().UTC(),
			}

			for labelName, labelValue := range sample.Metric {
				data[string(labelName)] = string(labelValue)
			}

			response.Data = append(response.Data, data)
		}

		response.Rows = int64(len(response.Data))
	case model.Matrix:
		if len(v) > 0 {
			response.Meta = []map[string]string{
				{
					"name": "value",
				},
				{
					"name": "timestamp",
				},
			}
			for labelName := range v[0].Metric {
				response.Meta = append(response.Meta, map[string]string{
					"name": string(labelName),
				})
			}
		}

		for _, stream := range v {
			for _, sample := range stream.Values {
				data := map[string]any{
					"value":     sample.Value,
					"timestamp": sample.Timestamp.Time().UTC(),
				}

				for labelName, labelValue := range stream.Metric {
					data[string(labelName)] = string(labelValue)
				}

				response.Data = append(response.Data, data)
			}
		}

		response.Rows = int64(len(response.Data))
	default:
		return response, fmt.Errorf("not supported result type: %T", v)
	}

	return response, nil
}

func (databaseConfig *DatabaseConfig) makeDatabaseResponseForElasticsearch(ctx context.Context, query string, timeout int) (DatabaseResponse, error) {
	client, err := DatabasePool.GetElasticsearchClient(*databaseConfig)
	if err != nil {
		return DatabaseResponse{}, err
	}

	return client.QueryForESQL(ctx, query, timeout)
}
