package command

import (
	"log/slog"

	"ntels.com/pharos/core/external/config_loader"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/migration"
	"ntels.com/pharos/core/pkg/slogrotate"
)

func Load(configFiles []string) (common.Config, error) {
	config, err := config_loader.Load(configFiles)
	if err != nil {
		slog.Error("config_loader load failed", "error", err)
		return common.Config{}, err
	}

	var logLevel slog.Level
	switch config.Logger.Level {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		slog.Warn("invalid log level in config, defaulting to INFO", "level", config.Logger.Level)
		logLevel = slog.LevelInfo
	}
	_, err = slogrotate.New(&config).SetupGlobal(&slog.HandlerOptions{
		Level:     logLevel,
		AddSource: true,
	}, config.Logger.IsJson, config.Logger.IsConsole)
	if err != nil {
		slog.Error("slogrotate setup failed", "error", err)
		return common.Config{}, err
	}

	if err := loadExtensions(config); err != nil {
		slog.Error("loadExtensions failed", "error", err)
		return common.Config{}, err
	}

	err = migration.Load(config)
	if err != nil {
		slog.Error("migration load failed", "error", err)
		return common.Config{}, err
	}

	// Initialize pools per role to avoid cross-domain sharing even for identical configs
	// Primary (default) DB
	if err = orm.LoadWithRole(orm.PoolDefault, config.Database); err != nil {
		slog.Error("orm load failed", "role", orm.PoolDefault, "error", err)
		return common.Config{}, err
	}
	// Statistics DB (optional)
	if config.Catv.Database.Driver != "" {
		if err = orm.LoadWithRole(orm.PoolStatistics, config.Catv.Database); err != nil {
			slog.Error("orm load failed", "role", orm.PoolStatistics, "error", err)
			return common.Config{}, err
		}
	}

	return config, nil
}

func Unload() {
	unloadExtensions()

	orm.Unload()
}
