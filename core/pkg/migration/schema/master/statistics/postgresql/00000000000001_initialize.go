package postgresql

import (
	"context"
	"database/sql"
	"runtime"

	"github.com/pressly/goose/v3"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/migration/schema/master/statistics"
)

var driver = orm.DriverPostgreSQL

func init() {
	_, filename, _, _ := runtime.Caller(0)
	version, err := goose.NumericComponent(filename)
	if err != nil {
		panic(err)
	}

	up := &goose.GoFunc{
		RunTx: nil,
		RunDB: func(ctx context.Context, db *sql.DB) error { return nil },
	}
	down := &goose.GoFunc{
		RunTx: nil,
		RunDB: func(ctx context.Context, db *sql.DB) error { return nil },
	}

	migration := goose.NewGoMigration(version, up, down)
	migration.Source = filename
	statistics.Migrations[driver] = append(statistics.Migrations[driver], migration)
}
