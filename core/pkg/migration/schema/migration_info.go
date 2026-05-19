package schema

import (
	"context"
	"database/sql"
	"embed"
	"errors"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/pressly/goose/v3"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/external/orm"
)

var addMigrationInfoFuncs []AddMigrationInfoFuncType

type AddMigrationInfoFuncType func() *MigrationInfo

func AddMigrationInfoFunc(fn AddMigrationInfoFuncType) {
	addMigrationInfoFuncs = append(addMigrationInfoFuncs, fn)
}

type MigrationInfo struct {
	Fs  embed.FS
	Dir string

	Migrations goose.Migrations
	TableName  string

	DatabaseConfig orm.DatabaseConfig
}

func (migrationInfo *MigrationInfo) up() error {
	dataSourceName := migrationInfo.DatabaseConfig.GetDataSourceName()

	switch migrationInfo.DatabaseConfig.Driver {
	case orm.DriverSqlite:
		path := filepath.Dir(dataSourceName)
		if _, err := os.Stat(path); os.IsNotExist(err) {
			if err := os.MkdirAll(path, 0o750); err != nil {
				return err
			}
		}
	}

	db, err := sql.Open(migrationInfo.DatabaseConfig.GetDriverName(), dataSourceName)
	if err != nil {
		return err
	}
	defer func() { _ = db.Close() }()

	fsys, err := fs.Sub(migrationInfo.Fs, migrationInfo.Dir)
	if err != nil {
		return err
	}

	var dialect goose.Dialect
	switch migrationInfo.DatabaseConfig.Driver {
	case orm.DriverSqlite:
		dialect = goose.DialectSQLite3
	default:
		dialect = goose.Dialect(migrationInfo.DatabaseConfig.GetDriverName())
	}

	// When ClickHouse cluster is configured, use a custom store that creates the
	// version table with ON CLUSTER + ReplicatedMergeTree so that the migration
	// state is replicated across all nodes and does not go out of sync.
	clusterName := migrationInfo.DatabaseConfig.Migration.ClickHouse.ClusterName
	useCluster := migrationInfo.DatabaseConfig.Driver == orm.DriverClickHouse && clusterName != ""

	var providerOpts []goose.ProviderOption
	providerOpts = append(providerOpts,
		goose.WithDisableGlobalRegistry(true),
		goose.WithGoMigrations(migrationInfo.Migrations...),
		goose.WithIsolateDDL(true),
		goose.WithSlog(slog.Default()),
	)

	if useCluster {
		store, err := newClickhouseClusterStore(migrationInfo.TableName, migrationInfo.DatabaseConfig.ClickHouse.Database, clusterName)
		if err != nil {
			return err
		}
		dialect = goose.DialectCustom
		providerOpts = append(providerOpts, goose.WithStore(store))
	} else {
		providerOpts = append(providerOpts, goose.WithTableName(migrationInfo.TableName))
	}

	provider, err := goose.NewProvider(
		dialect,
		db,
		fsys,
		providerOpts...)
	if err != nil {
		return err
	}

	var ctx = context.Background()

	if err := provider.Ping(ctx); err != nil {
		return err
	}

	if _, err := provider.Up(ctx); err != nil && !errors.Is(err, external.ErrorMigrationCompleted) {
		return err
	}

	if _, err := provider.Status(ctx); err != nil {
		return err
	}

	return nil
}

func OpenDBWithDriver(driver string, dbstring string) (*sql.DB, error) {
	if driver != orm.DriverName[orm.DriverPostgreSQL] {
		return goose.OpenDBWithDriver(driver, dbstring)
	}

	if err := goose.SetDialect(driver); err != nil {
		return nil, err
	}

	return sql.Open(driver, dbstring)
}
