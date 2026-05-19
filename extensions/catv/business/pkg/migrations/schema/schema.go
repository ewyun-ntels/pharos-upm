package schema

import (
	"embed"

	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	core_schema "ntels.com/pharos/core/pkg/migration/schema"
	master_statistics "ntels.com/pharos/extensions/catv/business/pkg/migrations/schema/master/statistics"
	_ "ntels.com/pharos/extensions/catv/business/pkg/migrations/schema/master/statistics/clickhouse"
)

//go:embed master/*
var masterFS embed.FS

func Load(config common.Config) {
	funcs := []core_schema.AddMigrationInfoFuncType{
		func() *core_schema.MigrationInfo {
			if config.Catv.Migration.Database.Driver != orm.DriverClickHouse {
				return nil
			}

			return &core_schema.MigrationInfo{
				Fs:             masterFS,
				Dir:            "master/statistics/clickhouse",
				Migrations:     master_statistics.Migrations[orm.DriverClickHouse],
				TableName:      master_statistics.TableName,
				DatabaseConfig: config.Catv.Migration.Database,
			}
		},
	}

	for _, f := range funcs {
		core_schema.AddMigrationInfoFunc(f)
	}
}
