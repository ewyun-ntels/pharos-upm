package data

import (
	"context"
	"database/sql"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	_ "modernc.org/sqlite"
)

func TestUpRestructureUserAttributes(t *testing.T) {
	// Create temp directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")

	// Open database
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		t.Fatalf("Failed to open database: %v", err)
	}
	defer func() { _ = db.Close() }()

	// Create users table
	createTableSQL := `
		CREATE TABLE users (
			username TEXT PRIMARY KEY,
			extra TEXT
		)
	`
	if _, err := db.Exec(createTableSQL); err != nil {
		t.Fatalf("Failed to create users table: %v", err)
	}

	tests := []struct {
		name          string
		username      string
		oldAttributes map[string]any
		expectedRoles map[string]bool
		expectedAttrs map[string]bool
		expectedInfo  map[string]any
		shouldMigrate bool
	}{
		{
			name:     "Full conversion with all types",
			username: "user1",
			oldAttributes: map[string]any{
				"role:super_admin":    true,
				"role:manage_user":    false,
				"attr:temporary_user": false,
				"userinfo": map[string]any{
					"email": "user1@example.com",
					"phone": "123-456-7890",
				},
			},
			expectedRoles: map[string]bool{
				"role:super_admin": true, // prefix 유지
				"role:manage_user": false,
			},
			expectedAttrs: map[string]bool{
				"temporary_user": false, // attr: prefix는 검증 시 자동 추가됨
			},
			expectedInfo: map[string]any{
				"email": "user1@example.com",
				"phone": "123-456-7890",
			},
			shouldMigrate: true,
		},
		{
			name:     "Only roles",
			username: "user2",
			oldAttributes: map[string]any{
				"role:super_admin": true,
			},
			expectedRoles: map[string]bool{
				"role:super_admin": true, // prefix 유지
			},
			expectedAttrs: map[string]bool{},
			expectedInfo:  map[string]any{},
			shouldMigrate: true,
		},
		{
			name:     "Already migrated (new format)",
			username: "user3",
			oldAttributes: map[string]any{
				"roles": map[string]any{
					"role:super_admin": true, // prefix 유지
				},
				"info": map[string]any{},
			},
			expectedRoles: map[string]bool{
				"role:super_admin": true, // prefix 유지
			},
			expectedAttrs: map[string]bool{},
			expectedInfo:  map[string]any{},
			shouldMigrate: false, // Should skip
		},
		{
			name:          "Empty attributes",
			username:      "user4",
			oldAttributes: map[string]any{},
			expectedRoles: map[string]bool{},
			expectedAttrs: map[string]bool{},
			expectedInfo:  map[string]any{},
			shouldMigrate: true,
		},
		{
			name:     "Unknown keys preserved in info",
			username: "user5",
			oldAttributes: map[string]any{
				"role:super_admin": true,
				"custom_field":     "custom_value",
				"another_field":    123,
			},
			expectedRoles: map[string]bool{
				"role:super_admin": true, // prefix 유지
			},
			expectedAttrs: map[string]bool{},
			expectedInfo: map[string]any{
				"custom_field":  "custom_value",
				"another_field": float64(123), // JSON numbers become float64
			},
			shouldMigrate: true,
		},
	}

	// Insert test data
	for _, tt := range tests {
		extraJSON, err := json.Marshal(tt.oldAttributes)
		if err != nil {
			t.Fatalf("Failed to marshal test data for %s: %v", tt.name, err)
		}

		insertSQL := `INSERT INTO users (username, extra) VALUES (?, ?)`
		if _, err := db.Exec(insertSQL, tt.username, string(extraJSON)); err != nil {
			t.Fatalf("Failed to insert test data for %s: %v", tt.name, err)
		}
	}

	// Run migration
	ctx := context.Background()
	tx, err := db.Begin()
	if err != nil {
		t.Fatalf("Failed to begin transaction: %v", err)
	}
	_ = tx.Rollback() // Clean up

	// Run migration (no transaction needed - migration handles its own)
	if err := upRestructureUserAttributes(ctx, db); err != nil {
		t.Fatalf("Migration failed: %v", err)
	}

	// Verify results
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var extraJSON string
			query := `SELECT extra FROM users WHERE username = ?`
			if err := db.QueryRow(query, tt.username).Scan(&extraJSON); err != nil {
				t.Fatalf("Failed to query user %s: %v", tt.username, err)
			}

			var newAttrs UserAttributesNew
			if err := json.Unmarshal([]byte(extraJSON), &newAttrs); err != nil {
				t.Fatalf("Failed to parse new attributes for %s: %v", tt.username, err)
			}

			// Verify roles (roles 맵에는 role:* 와 attr:* 둘 다 포함됨)
			expectedTotalRoles := len(tt.expectedRoles) + len(tt.expectedAttrs)
			if len(newAttrs.Roles) != expectedTotalRoles {
				t.Errorf("Roles count mismatch for %s: got %d, want %d",
					tt.username, len(newAttrs.Roles), expectedTotalRoles)
			}
			for role, expected := range tt.expectedRoles {
				if got := newAttrs.Roles[role]; got != expected {
					t.Errorf("Role %s mismatch for %s: got %v, want %v",
						role, tt.username, got, expected)
				}
			}

			// Verify attrs (now in roles with attr: prefix)
			for attr, expected := range tt.expectedAttrs {
				attrKey := "attr:" + attr
				if got, exists := newAttrs.Roles[attrKey]; !exists {
					t.Errorf("Attr %s not found in roles for %s", attrKey, tt.username)
				} else if got != expected {
					t.Errorf("Attr %s mismatch for %s: got %v, want %v",
						attrKey, tt.username, got, expected)
				}
			}

			// Verify info (basic length check)
			if len(newAttrs.Info) != len(tt.expectedInfo) {
				t.Errorf("Info count mismatch for %s: got %d, want %d",
					tt.username, len(newAttrs.Info), len(tt.expectedInfo))
			}
		})
	}
}

func TestHasNewFormat(t *testing.T) {
	tests := []struct {
		name     string
		attrs    UserAttributesLegacy
		expected bool
	}{
		{
			name: "Has roles key",
			attrs: UserAttributesLegacy{
				"roles": map[string]bool{"super_admin": true},
			},
			expected: true,
		},
		{
			name: "Has attrs key (old format - should be false)",
			attrs: UserAttributesLegacy{
				"attrs": map[string]bool{"temporary_user": false},
			},
			expected: false, // attrs 키는 새 포맷에 없음, roles 키만 존재
		},
		{
			name: "Has info key",
			attrs: UserAttributesLegacy{
				"info": map[string]any{"email": "test@example.com"},
			},
			expected: true,
		},
		{
			name: "Old format with prefixes",
			attrs: UserAttributesLegacy{
				"role:super_admin":    true,
				"attr:temporary_user": false,
			},
			expected: false,
		},
		{
			name:     "Empty",
			attrs:    UserAttributesLegacy{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := hasNewFormat(tt.attrs)
			if got != tt.expected {
				t.Errorf("hasNewFormat() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestConvertLegacyToStructured(t *testing.T) {
	tests := []struct {
		name          string
		legacy        UserAttributesLegacy
		expectedRoles map[string]bool
		expectedAttrs map[string]bool
		expectedInfo  map[string]any
	}{
		{
			name: "Full conversion",
			legacy: UserAttributesLegacy{
				"role:super_admin":    true,
				"role:manage_user":    false,
				"attr:temporary_user": false,
				"userinfo": map[string]any{
					"email": "test@example.com",
				},
				"custom": "value",
			},
			expectedRoles: map[string]bool{
				"super_admin": true, // role: prefix는 검증 시 자동 추가됨
				"manage_user": false,
			},
			expectedAttrs: map[string]bool{
				"temporary_user": false, // attr: prefix는 검증 시 자동 추가됨
			},
			expectedInfo: map[string]any{
				"email":  "test@example.com",
				"custom": "value",
			},
		},
		{
			name:          "Empty",
			legacy:        UserAttributesLegacy{},
			expectedRoles: map[string]bool{},
			expectedAttrs: map[string]bool{},
			expectedInfo:  map[string]any{},
		},
		{
			name: "Composite roles without prefix",
			legacy: UserAttributesLegacy{
				"esn_admin_role":     true,
				"samsung_admin_role": true,
				"pv_admin_role":      false,
			},
			expectedRoles: map[string]bool{},
			expectedAttrs: map[string]bool{},
			expectedInfo:  map[string]any{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := convertLegacyToStructured(tt.legacy)

			// Verify roles (should include both role:* and attr:* keys, and *_role keys)
			expectedRolesCount := len(tt.expectedRoles) + len(tt.expectedAttrs)

			// Count expected composite roles (keys ending with _role in legacy)
			for key := range tt.legacy {
				if strings.HasSuffix(key, "_role") && !strings.HasPrefix(key, "role:") {
					expectedRolesCount++
				}
			}

			if len(result.Roles) != expectedRolesCount {
				t.Errorf("Roles count: got %d, want %d (result.Roles: %v)", len(result.Roles), expectedRolesCount, result.Roles)
			}

			// Verify role:* keys
			for role, expected := range tt.expectedRoles {
				if got := result.Roles["role:"+role]; got != expected {
					t.Errorf("Role %s: got %v, want %v", role, got, expected)
				}
			}

			// Verify attr:* keys (now in roles map)
			for attr, expected := range tt.expectedAttrs {
				if got := result.Roles["attr:"+attr]; got != expected {
					t.Errorf("Attr %s: got %v, want %v", attr, got, expected)
				}
			}

			// Verify composite roles (*_role keys without prefix)
			for key, value := range tt.legacy {
				if strings.HasSuffix(key, "_role") && !strings.HasPrefix(key, "role:") {
					expected := toBool(value)
					if got := result.Roles[key]; got != expected {
						t.Errorf("Composite role %s: got %v, want %v", key, got, expected)
					}
				}
			}

			// Verify info
			if len(result.Info) != len(tt.expectedInfo) {
				t.Errorf("Info count: got %d, want %d", len(result.Info), len(tt.expectedInfo))
			}
		})
	}
}

func TestToBool(t *testing.T) {
	tests := []struct {
		name     string
		value    any
		expected bool
	}{
		{"bool true", true, true},
		{"bool false", false, false},
		{"string true", "true", true},
		{"string 1", "1", true},
		{"string false", "false", false},
		{"string other", "other", false},
		{"int 0", 0, false},
		{"int 1", 1, true},
		{"int negative", -1, true},
		{"float64 0", float64(0), false},
		{"float64 non-zero", float64(1.5), true},
		{"nil", nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := toBool(tt.value)
			if got != tt.expected {
				t.Errorf("toBool(%v) = %v, want %v", tt.value, got, tt.expected)
			}
		})
	}
}

func TestMain(m *testing.M) {
	os.Exit(m.Run())
}
