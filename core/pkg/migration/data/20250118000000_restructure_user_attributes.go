package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"maps"
	"strings"
)

func init() {
	Register(&Migration{
		Name: "restructure_user_attributes",
		Up:   upRestructureUserAttributes,
		Down: downRestructureUserAttributes,
	})
}

// UserAttributesLegacy represents the old flat structure
type UserAttributesLegacy map[string]any

// UserAttributesNew represents the new nested structure
// Following Keycloak pattern: all roles and attributes in single "roles" map
type UserAttributesNew struct {
	Roles map[string]bool `json:"roles"`
	Info  map[string]any  `json:"info"`
}

// upRestructureUserAttributes migrates user attributes from flat to nested structure
// Old format: {"role:super_admin": true, "attr:temporary_user": false, "userinfo": {...}}
// New format: {"roles": {"super_admin": true}, "attrs": {"temporary_user": false}, "info": {...}}
func upRestructureUserAttributes(ctx context.Context, db *sql.DB) error {
	// Detect driver (SQLite vs PostgreSQL)
	driver := getDriverFromDB(db)

	// Query all users with attributes
	// Load into memory to avoid database lock (SQLite) or transaction conflicts (PostgreSQL)
	query := `SELECT username, extra FROM users WHERE extra IS NOT NULL AND extra != ''`
	rows, err := db.QueryContext(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to query users: %w", err)
	}

	type userRow struct {
		username  string
		extraJSON string
	}
	var userRows []userRow

	for rows.Next() {
		var username string
		var extraJSON string
		if err := rows.Scan(&username, &extraJSON); err != nil {
			_ = rows.Close()
			return fmt.Errorf("failed to scan row: %w", err)
		}
		userRows = append(userRows, userRow{username, extraJSON})
	}

	if err := rows.Err(); err != nil {
		_ = rows.Close()
		return fmt.Errorf("error iterating rows: %w", err)
	}
	_ = rows.Close() // Close before UPDATE queries

	stats := &AttributesMigrationStats{}

	// Process each user
	for _, row := range userRows {
		stats.Total++

		// Parse old attributes
		var oldAttrs UserAttributesLegacy
		if err := json.Unmarshal([]byte(row.extraJSON), &oldAttrs); err != nil {
			fmt.Printf("Warning: failed to parse attributes for user %s: %v\n", row.username, err)
			stats.Skipped++
			continue
		}

		// Check if already in new format
		if hasNewFormat(oldAttrs) {
			fmt.Printf("User %s already has new format, skipping\n", row.username)
			stats.AlreadyMigrated++
			continue
		}

		// Convert to new format
		newAttrs := convertLegacyToStructured(oldAttrs)

		// Marshal new attributes
		newJSON, err := json.Marshal(newAttrs)
		if err != nil {
			return fmt.Errorf("failed to marshal new attributes for user %s: %w", row.username, err)
		}

		// Update user
		var updateQuery string
		if driver == "postgres" {
			updateQuery = `UPDATE users SET extra = $1 WHERE username = $2`
		} else {
			updateQuery = `UPDATE users SET extra = ? WHERE username = ?`
		}

		if _, err := db.ExecContext(ctx, updateQuery, string(newJSON), row.username); err != nil {
			return fmt.Errorf("failed to update user %s: %w", row.username, err)
		}

		stats.Migrated++
	}

	// Print statistics
	fmt.Printf("20250118 User Attributes Restructure Statistics:\n")
	fmt.Printf("  Total users processed: %d\n", stats.Total)
	fmt.Printf("  Successfully migrated: %d\n", stats.Migrated)
	fmt.Printf("  Already in new format: %d\n", stats.AlreadyMigrated)
	fmt.Printf("  Skipped (parse errors): %d\n", stats.Skipped)

	return nil
}

func downRestructureUserAttributes(_ context.Context, _ *sql.DB) error {
	return fmt.Errorf("rollback not supported for user attributes migration - backup recommended before migration")
}

// AttributesMigrationStats tracks migration statistics
type AttributesMigrationStats struct {
	Total           int
	Migrated        int
	AlreadyMigrated int
	Skipped         int
}

// hasNewFormat checks if attributes already use the new nested structure
func hasNewFormat(attrs UserAttributesLegacy) bool {
	_, hasRoles := attrs["roles"]
	_, hasInfo := attrs["info"]
	return hasRoles || hasInfo
}

// convertLegacyToStructured converts flat attributes to nested structure
// Following Keycloak pattern: keep prefixes intact in roles map
func convertLegacyToStructured(legacy UserAttributesLegacy) UserAttributesNew {
	newAttrs := UserAttributesNew{
		Roles: make(map[string]bool),
		Info:  make(map[string]any),
	}

	for key, value := range legacy {
		switch {
		case strings.HasPrefix(key, "role:"), strings.HasPrefix(key, "attr:"):
			// Keep the full key with prefix in roles map
			newAttrs.Roles[key] = toBool(value)

		case strings.HasSuffix(key, "_role"):
			// Composite role names without prefix (e.g., esn_admin_role, samsung_admin_role)
			// These should go to roles map
			newAttrs.Roles[key] = toBool(value)

		case key == "userinfo":
			// Move userinfo contents to info
			if infoMap, ok := value.(map[string]any); ok {
				maps.Copy(newAttrs.Info, infoMap)
			}

		default:
			// Unknown keys - preserve in info
			newAttrs.Info[key] = value
		}
	}

	return newAttrs
}

// toBool converts interface{} to bool
func toBool(value any) bool {
	switch v := value.(type) {
	case bool:
		return v
	case string:
		return v == "true" || v == "1"
	case float64:
		return v != 0
	case int:
		return v != 0
	default:
		return false
	}
}

// getDriverFromDB attempts to detect the driver from DB connection
func getDriverFromDB(db *sql.DB) string {
	// Query to check PostgreSQL-specific syntax
	_, err := db.Query("SELECT version()")
	if err == nil {
		return "postgres"
	}
	return "sqlite"
}
