package migration

import (
	"log/slog"

	"ntels.com/pharos/core/internal/repositories/factory"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/migration/data"
	"ntels.com/pharos/core/pkg/migration/schema"
)

// Load runs all migrations (schema and data)
func Load(cfg common.Config) error {
	// Check if this is a fresh installation BEFORE schema migration runs
	isFreshInstall, err := IsFreshInstall(cfg)
	if err != nil {
		slog.Warn("Failed to check fresh install status, assuming existing installation", "error", err)
		isFreshInstall = false
	}

	// Run schema migrations (creates tables)
	if err := schema.Up(cfg); err != nil {
		return err
	}

	// Run data migrations with fresh install flag
	if err := data.RunWithFreshInstallFlag(cfg, isFreshInstall); err != nil {
		return err
	}

	return nil
}

// IsFreshInstall checks if this is a fresh installation by checking goose_db_version table existence
func IsFreshInstall(cfg common.Config) (bool, error) {
	repo, err := factory.NewDataMigrationRepository(factory.RepositoryOptions{
		DatabaseConfig: cfg.Database,
	})
	if err != nil {
		return false, err
	}

	isFresh, err := repo.IsFreshInstall()
	if err != nil {
		return false, err
	}

	if isFresh {
		slog.Info("Fresh installation detected", "reason", "no core_db_version table")
	} else {
		slog.Info("Existing installation detected", "reason", "core_db_version table exists")
	}

	return isFresh, nil
}
