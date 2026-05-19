package authhandler

import (
	"database/sql"
	"os"
	"testing"

	_ "modernc.org/sqlite"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/shared/types/role"
)

// Load는 테스트용 wrapper 함수입니다
func Load(configPath string, config common.Config) error {
	api := &Api{}
	api.Init(configPath, config)
	return api.Load()
}

// helper to initialize auth subsystem with a file-based sqlite DB and minimal schema
func setupAuthForTest(t *testing.T) func() {
	t.Helper()
	// Create temp sqlite file
	tmp := t.TempDir() + "/auth_test.db"
	// Pre-create minimal schema so user store can operate
	db, err := sql.Open("sqlite", tmp)
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}
	ddl := `
	CREATE TABLE IF NOT EXISTS users (
	    username TEXT PRIMARY KEY NOT NULL,
	    password_hash BLOB,
	    extra TEXT,
	    created_at DATETIME,
	    retry INTEGER DEFAULT 0,
	    retry_at DATETIME,
	    blocked BOOLEAN DEFAULT FALSE,
	    password_updated_at DATETIME,
	    password_expired_at DATETIME,
	    prepare TEXT
	);
	CREATE TABLE IF NOT EXISTS user_password_history (
	    id INTEGER PRIMARY KEY AUTOINCREMENT,
	    username TEXT NOT NULL,
	    password_hash BLOB NOT NULL,
	    created_at DATETIME
	);
	`
	if _, err := db.Exec(ddl); err != nil {
		_ = db.Close()
		t.Fatalf("failed to create schema: %v", err)
	}
	_ = db.Close()

	cfg := common.Config{}
	cfg.Database.Driver = "sqlite"
	cfg.Database.SQLite.Path = tmp
	cfg.Auth.GlobalSecret = "0123456789abcdef0123456789abcdef"
	cfg.Auth.JWTIssuer = "test"
	cfg.User.PasswordHistoryRetention = 3
	cfg.User.UserRule.MinLength = 1
	cfg.User.UserRule.MaxLength = 32
	cfg.Auth.JwtCert.AutoGenerate = true
	cfg.Auth.JwtCert.JwtDir = os.TempDir()
	cfg.Auth.JwtCert.AutoGenerateInfo.NotAfter = "2100-01-01T23:59:59Z"
	// minimal server config for Load signature
	if err := Load("", cfg); err != nil {
		t.Fatalf("auth.Load failed: %v", err)
	}
	return func() {}
}

func TestAuthAPI_CreateUser_DeleteUser(t *testing.T) {
	teardown := setupAuthForTest(t)
	defer teardown()

	username := "migration"
	password := "secret"
	attrs := map[string]any{
		common.ExtraKeyRoles: map[string]bool{
			string(role.RoleSuperAdmin): true, // role.RoleSuperAdmin = "role:super_admin"
		},
		common.ExtraKeyInfo: map[string]any{},
	}

	// ensure not exists
	_ = DeleteUser(username)

	if err := CreateUser(username, password, attrs, nil); err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	// calling CreateUser again should trigger delete+recreate path and still succeed
	if err := CreateUser(username, password, attrs, nil); err != nil {
		t.Fatalf("CreateUser (recreate) failed: %v", err)
	}

	if err := DeleteUser(username); err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}
}
