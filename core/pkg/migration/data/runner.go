package data

import (
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strconv"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/factory"
	"ntels.com/pharos/core/pkg/common"
)

// Migration represents a single data migration
type Migration struct {
	Version int64
	Name    string
	Up      func(ctx context.Context, db *sql.DB) error
	Down    func(ctx context.Context, db *sql.DB) error
}

// migrations holds all registered migrations
var migrations []*Migration

// versionPattern extracts version from filename (e.g., 20251017000000_name.go → 20251017000000)
var versionPattern = regexp.MustCompile(`^(\d{14})_`)

// Register registers a new migration (called from init() in migration files)
// It automatically extracts the version from the caller's filename
func Register(m *Migration) {
	// Auto-detect version from filename if not set
	if m.Version == 0 {
		_, file, _, ok := runtime.Caller(1)
		if !ok {
			panic("failed to get caller info for migration registration")
		}

		filename := filepath.Base(file)
		matches := versionPattern.FindStringSubmatch(filename)
		if len(matches) < 2 {
			panic(fmt.Sprintf("migration filename %s does not match pattern YYYYMMDDHHMMSS_name.go", filename))
		}

		version, err := strconv.ParseInt(matches[1], 10, 64)
		if err != nil {
			panic(fmt.Sprintf("failed to parse version from filename %s: %v", filename, err))
		}
		m.Version = version
	}

	migrations = append(migrations, m)
}

// RunWithFreshInstallFlag executes all pending data migrations with pre-determined fresh install flag
func RunWithFreshInstallFlag(config common.Config, isFreshInstall bool) error {
	// Create data migration repository
	dataMigrationRepo, err := factory.NewDataMigrationRepository(factory.RepositoryOptions{
		DatabaseConfig: config.Database,
	})
	if err != nil {
		return fmt.Errorf("failed to create data migration repository: %w", err)
	}

	// Create migration table
	if err := dataMigrationRepo.CreateMigrationTable(); err != nil {
		return fmt.Errorf("failed to create migration table: %w", err)
	}

	// Sort migrations by version
	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	if isFreshInstall {
		slog.Info("Fresh installation confirmed - marking all data migrations as applied without execution")

		// Collect all versions
		versions := make([]int64, len(migrations))
		for i, m := range migrations {
			versions[i] = m.Version
		}

		// Mark all as applied (without running them)
		if err := dataMigrationRepo.MarkAllAsApplied(versions); err != nil {
			return fmt.Errorf("failed to mark migrations as applied: %w", err)
		}

		slog.Info("All data migrations marked as applied", "count", len(versions))

		if err := printStatus(dataMigrationRepo); err != nil {
			slog.Warn("Failed to print migration status", "error", err)
		}

		return nil
	}

	// Not fresh install - run migrations normally
	applied, err := dataMigrationRepo.GetAppliedMigrations()
	if err != nil {
		return fmt.Errorf("failed to get applied migrations: %w", err)
	}

	// Execute each migration with its own orm.Handler call
	ctx := context.Background()
	for _, m := range migrations {
		if applied[m.Version] {
			slog.Info("Migration already applied", "version", m.Version, "name", m.Name)
			continue
		}

		slog.Info("Applying migration", "version", m.Version, "name", m.Name)

		// Each migration gets its own DB connection from pool
		// Migration decides whether to use transaction or not
		err := orm.Handler("", &config.Database, func(db *sqlx.DB) error {
			sqlDB := db.DB

			// Execute migration (migration handles its own transaction if needed)
			if err := m.Up(ctx, sqlDB); err != nil {
				return fmt.Errorf("migration failed: %w", err)
			}

			// Record migration success in a transaction
			tx, err := sqlDB.Begin()
			if err != nil {
				return fmt.Errorf("failed to begin transaction for recording: %w", err)
			}

			if err := dataMigrationRepo.RecordMigration(tx, m.Version); err != nil {
				_ = tx.Rollback()
				return fmt.Errorf("failed to record migration: %w", err)
			}

			if err := tx.Commit(); err != nil {
				return fmt.Errorf("failed to commit migration record: %w", err)
			}

			return nil
		})

		if err != nil {
			return fmt.Errorf("migration %d (%s) failed: %w", m.Version, m.Name, err)
		}

		slog.Info("Migration applied successfully", "version", m.Version, "name", m.Name)
	}

	if err := printStatus(dataMigrationRepo); err != nil {
		slog.Warn("Failed to print migration status", "error", err)
	}

	return nil
}

func printStatus(repo repositories.DataMigrationRepository) error {
	applied, err := repo.GetAppliedMigrations()
	if err != nil {
		return err
	}

	sort.Slice(migrations, func(i, j int) bool {
		return migrations[i].Version < migrations[j].Version
	})

	slog.Info("Data Migration Status:")
	slog.Info("    =======================================")
	for _, m := range migrations {
		status := "Pending"
		if applied[m.Version] {
			status = "Applied"
		}
		slog.Info(fmt.Sprintf("    [%s] %d - %s", status, m.Version, m.Name))
	}
	slog.Info("    =======================================")
	return nil
}
