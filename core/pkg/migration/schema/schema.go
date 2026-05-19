package schema

import (
	"embed"
	"log/slog"

	"ntels.com/pharos/core/internal/migration"
	"ntels.com/pharos/core/pkg/common"
	master_base "ntels.com/pharos/core/pkg/migration/schema/master/base"
	_ "ntels.com/pharos/core/pkg/migration/schema/master/base/postgresql"
	_ "ntels.com/pharos/core/pkg/migration/schema/master/base/sqlite"
	master_statistics "ntels.com/pharos/core/pkg/migration/schema/master/statistics"
	_ "ntels.com/pharos/core/pkg/migration/schema/master/statistics/clickhouse"
	_ "ntels.com/pharos/core/pkg/migration/schema/master/statistics/postgresql"
	_ "ntels.com/pharos/core/pkg/migration/schema/master/statistics/sqlite"
)

//go:embed master/*
var masterFS embed.FS

func Up(config common.Config) error {
	var infos []MigrationInfo

	infos = append(infos, MigrationInfo{
		Fs:             masterFS,
		Dir:            "master/base/" + config.Database.Driver,
		Migrations:     master_base.Migrations[config.Database.Driver],
		TableName:      migration.CoreDBVersionTableName,
		DatabaseConfig: config.Database,
	})

	infos = append(infos, MigrationInfo{
		Fs:             masterFS,
		Dir:            "master/statistics/" + config.Statistics.Database.Driver,
		Migrations:     master_statistics.Migrations[config.Statistics.Database.Driver],
		TableName:      master_statistics.TableName,
		DatabaseConfig: config.Statistics.Database,
	})

	for _, fn := range addMigrationInfoFuncs {
		if info := fn(); info != nil {
			infos = append(infos, *info)
		}
	}

	for _, info := range infos {
		if err := info.up(); err != nil {
			slog.Error("migration up failed", "dir", info.Dir, "table", info.TableName, "error", err)
			return err
		}
	}

	return nil
}

func UpMasterBase(config common.Config) error {
	info := MigrationInfo{
		Fs:             masterFS,
		Dir:            "master/base/" + config.Database.Driver,
		Migrations:     master_base.Migrations[config.Database.Driver],
		TableName:      migration.CoreDBVersionTableName,
		DatabaseConfig: config.Database,
	}

	return info.up()
}

func UpMasterStatistics(config common.Config) error {
	if config.Statistics.Database.Driver == "" {
		return nil
	}
	info := MigrationInfo{
		Fs:             masterFS,
		Dir:            "master/statistics/" + config.Statistics.Database.Driver,
		Migrations:     master_statistics.Migrations[config.Statistics.Database.Driver],
		TableName:      master_statistics.TableName,
		DatabaseConfig: config.Statistics.Database,
	}
	return info.up()
}
