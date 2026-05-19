package sqlite

import (
	"context"
	"database/sql"
	"math"
	"runtime"

	"github.com/jmoiron/sqlx"
	"github.com/pressly/goose/v3"
	"github.com/pressly/goose/v3/database"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/migration"
	"ntels.com/pharos/core/pkg/migration/schema/master/base"
)

var driver = orm.DriverSqlite

const oldTableName = "goose_db_version"

func init() {
	_, filename, _, _ := runtime.Caller(0)
	version, err := goose.NumericComponent(filename)
	if err != nil {
		panic(err)
	}

	up := &goose.GoFunc{
		RunTx: nil,
		RunDB: func(ctx context.Context, db *sql.DB) error {
			handler := func(db *sqlx.DB) error {
				count := 0
				if err := db.Get(&count, `SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='`+oldTableName+`';`); err != nil {
					return err
				} else if count == 0 {
					return nil
				}

				oldStore, err := database.NewStore(database.DialectSQLite3, oldTableName)
				if err != nil {
					return err
				}

				newStore, err := database.NewStore(database.DialectSQLite3, migration.CoreDBVersionTableName)
				if err != nil {
					return err
				}

				if err := newStore.Insert(ctx, db.DB, database.InsertRequest{Version: version}); err != nil {
					return err
				}

				oldVersions := map[int64]*database.ListMigrationsResult{}
				listMigrationsResults, err := oldStore.ListMigrations(ctx, db)
				if err != nil {
					return err
				}
				for _, listMigrationsResult := range listMigrationsResults {
					oldVersions[listMigrationsResult.Version] = listMigrationsResult
				}

				migrations, err := goose.CollectMigrations(".", 0, math.MaxInt64)
				if err != nil {
					return err
				}
				for _, m := range migrations {
					if m.Version <= version {
						continue
					}

					if _, exist := oldVersions[m.Version]; !exist {
						continue
					}

					if err := newStore.Insert(ctx, db.DB, database.InsertRequest{Version: m.Version}); err != nil {
						return err
					}
				}

				return external.ErrorMigrationCompleted
			}

			return handler(sqlx.NewDb(db, orm.DriverName[driver]))
		},
	}

	down := &goose.GoFunc{
		RunTx: nil,
		RunDB: func(ctx context.Context, db *sql.DB) error {
			return nil
		},
	}

	newGoMigration := goose.NewGoMigration(version, up, down)
	newGoMigration.Source = filename
	base.Migrations[driver] = append(base.Migrations[driver], newGoMigration)
}
