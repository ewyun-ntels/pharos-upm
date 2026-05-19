// Package data contains data migration scripts that modify existing data structure.
//
// Data migrations run AFTER schema migrations and application startup.
// Each migration file should:
//  1. Follow naming convention: YYYYMMDDHHMMSS_description.go
//  2. Register itself in init() function using Register()
//  3. Implement up and down migration functions
//
// Example:
//
//	func init() {
//	    Register(&Migration{
//	        Name:    "your_migration_name",  // Version auto-detected from filename!
//	        Up:      upYourMigration,
//	        Down:    downYourMigration,
//	    })
//	}
//
//	func upYourMigration(ctx context.Context, tx *sql.Tx) error {
//	    // Migration logic here
//	    return nil
//	}
//
//	func downYourMigration(ctx context.Context, tx *sql.Tx) error {
//	    // Rollback logic here
//	    return nil
//	}
//
// VERSION AUTO-DETECTION:
// The Register() function automatically extracts the version from your filename.
// Example: 20251017120000_your_migration.go → Version: 20251017120000
//
// You can manually override the version if needed (not recommended):
//
//	Register(&Migration{
//	    Version: 20251017120000,  // Manual override
//	    Name:    "your_migration_name",
//	    Up:      upYourMigration,
//	    Down:    downYourMigration,
//	})
//
// All migration files are automatically registered when this package is imported.
// No external dependencies (like goose) required!
package data
