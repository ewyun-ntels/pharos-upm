package server

import (
	"log/slog"
	"time"

	"ntels.com/pharos/core/external/config_loader"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/migration"
	"ntels.com/pharos/core/pkg/pools/connection"
	"ntels.com/pharos/core/pkg/pools/workers"
	"ntels.com/pharos/core/pkg/slogrotate"
)

var config common.Config
var GlobalWorkerPool *workers.WorkerPool
var GlobalConnectionPool *connection.ConnectionPool

func load(configFiles []string) error {
	//configPaths := []string{}
	//
	//for _, configFile := range configFiles {
	//	configPaths = append(configPaths, filepath.Dir(configFile))
	//}
	//configPath = filepath.Dir(configFile)

	cfg, err := config_loader.Load(configFiles)
	if err != nil {
		slog.Error("config_loader load failed", "error", err)
		return err
	}
	config = cfg

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
		return err
	}

	if maxIdleTime, err := time.ParseDuration(config.Pool.Connection.MaxIdleTime); err != nil {
		slog.Error("invalid connection pool max idle time", "error", err)
		return err
	} else {
		GlobalConnectionPool = connection.NewConnectionPool(maxIdleTime, config.Pool.Connection.MaxSize, config.Pool.Connection.ShardCount)
	}

	GlobalWorkerPool = workers.NewWorkerPool(workers.WorkerPoolOptions{
		WorkerCount:   config.Pool.Workers.WorkerCount,
		JobQueueSize:  config.Pool.Workers.JobQueueSize,
		EnableMetrics: config.Pool.Workers.EnableMetrics,
	})

	if err := loadExtensions(config); err != nil {
		slog.Error("loadExtensions failed", "error", err)
		return err
	}

	err = migration.Load(config)
	if err != nil {
		slog.Error("migration load failed", "error", err)
		return err
	}

	// Initialize pools per role to avoid cross-domain sharing even for identical configs
	// Primary (default) DB
	if err = orm.LoadWithRole(orm.PoolDefault, config.Database); err != nil {
		slog.Error("orm load failed", "role", orm.PoolDefault, "error", err)
		return err
	}
	// Statistics DB (optional)
	if config.Statistics.Database.Driver != "" {
		if err = orm.LoadWithRole(orm.PoolStatistics, config.Statistics.Database); err != nil {
			slog.Error("orm load failed", "role", orm.PoolStatistics, "error", err)
			return err
		}
	}

	// Start worker pool after all initialization is complete
	GlobalWorkerPool.Start()

	return nil
}

func unload() {
	unloadExtensions()

	GlobalWorkerPool.Stop()
	GlobalConnectionPool.Close()

	orm.Unload()
}
