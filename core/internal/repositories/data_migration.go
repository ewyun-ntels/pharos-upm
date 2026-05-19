package repositories

import "database/sql"

// DataMigrationRepository handles data migration version tracking
type DataMigrationRepository interface {
	// CreateMigrationTable creates the migration version table if not exists
	CreateMigrationTable() error

	// GetAppliedMigrations returns a map of applied migration versions
	GetAppliedMigrations() (map[int64]bool, error)

	// RecordMigration records a migration as applied
	RecordMigration(tx *sql.Tx, version int64) error

	// IsFreshInstall checks if this is a fresh installation
	// Returns true if core_db_version table doesn't exist (never ran schema migrations)
	IsFreshInstall() (bool, error)

	// MarkAllAsApplied marks all given versions as applied without running them
	// Used for fresh installations where schema is already up-to-date
	MarkAllAsApplied(versions []int64) error
}
