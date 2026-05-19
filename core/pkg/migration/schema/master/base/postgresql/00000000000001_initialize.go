package postgresql

import (
	"context"
	"database/sql"
	"runtime"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/migration/schema/master/base"
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
		RunDB: func(ctx context.Context, db *sql.DB) error {
			return nil
		},
	}

	down := &goose.GoFunc{
		RunTx: nil,
		RunDB: func(ctx context.Context, db *sql.DB) error {
			handler := func(_ *sqlx.DB) error { return nil }

			return handler(sqlx.NewDb(db, orm.DriverName[driver]))
		},
	}

	migration := goose.NewGoMigration(version, up, down)
	migration.Source = filename
	base.Migrations[driver] = append(base.Migrations[driver], migration)
}
