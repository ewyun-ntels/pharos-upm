package api

import (
	"context"
	"crypto/rand"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/loykin/apirun"
	"github.com/loykin/apirun/pkg/env"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/authhandler"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/shared/types/role"
)

func buildEnvFromConfig(cfg common.Config) *env.Env {
	base := env.New()
	for _, e := range cfg.Migration.Env {
		if e.Name == "" {
			continue
		}
		var val string
		if e.ValueFromEnv != "" {
			if v, ok := os.LookupEnv(e.ValueFromEnv); ok {
				val = v
			} else {
				val = ""
			}
		} else {
			val = e.Value
		}
		base.Global[e.Name] = env.Str(val)
	}
	return base
}

// makeAPIBase computes the server base URL from config (schema + localhost + port).
// It uses serve.server_schema and servers.<schema>.port (defaulting to 8080 when zero).
func makeAPIBase(cfg common.Config) string {
	schema := cfg.Serve.ServerSchema
	if schema == "" {
		schema = common.SchemaHttp
	}
	port := 0
	if srv, ok := cfg.Servers[schema]; ok {
		port = srv.Port
	}
	if port == 0 {
		port = 8080
	}
	return fmt.Sprintf("%s://localhost:%d", schema, port)
}

// Run reads migration settings from the shared common.Config and runs MigrateUp in background.
// It never panics; errors are logged and ignored so server startup is not blocked.
func Run(cfg common.Config) {
	if cfg.Migration.Disable {
		slog.Info("migration: disabled")
		return
	}

	// Run in a goroutine to avoid blocking callers even if they call Run synchronously.
	go func() {
		// Optional configurable wait before starting migrations
		if d := cfg.Migration.Wait; d != "" {
			if dur, err := time.ParseDuration(d); err == nil {
				slog.Info("migration: waiting before start", "wait", dur.String())
				time.Sleep(dur)
			} else {
				slog.Error("migration: invalid wait duration", "wait", d, "error", err)
			}
		}

		dir := cfg.Migration.MigrationDir
		if dir == "" {
			slog.Info("migration: skipped (no [migration].dir configured)")
			return
		}

		ctx := context.Background()

		var dbConfig apirun.StoreConfig
		dbConfig.TableNames = apirun.TableNames{
			SchemaMigrations: "core_app_data_schema_migrations",
			MigrationRuns:    "core_app_data_migration_runs",
			StoredEnv:        "core_app_data_stored_env",
		}
		switch cfg.Database.Driver {
		case orm.DriverSqlite:
			dbConfig.DriverConfig = &apirun.SqliteConfig{
				Path: cfg.Database.SQLite.Path,
			}
			dbConfig.Driver = apirun.DriverSqlite
		case orm.DriverPostgreSQL:
			dbConfig.DriverConfig = &apirun.PostgresConfig{
				Host:     cfg.Database.PostgreSQL.Host,
				Port:     cfg.Database.PostgreSQL.Port,
				User:     cfg.Database.PostgreSQL.Username,
				Password: cfg.Database.PostgreSQL.Password,
				DBName:   cfg.Database.PostgreSQL.Database,
				SSLMode:  "disable",
			}
			dbConfig.Driver = apirun.DriverPostgresql
		default:
			slog.Info("migration: skipped (unsupported database driver)", "driver", cfg.Database.Driver)
			return
		}

		// Build environment for apimigrate templates using github.com/loykin/apimigrate/pkg/env
		base := buildEnvFromConfig(cfg)
		// Inject/override api_base so migrations can resolve server URL without requiring it in config
		apiBase := makeAPIBase(cfg)
		if base.Global == nil {
			base.Global = env.Map{}
		}
		base.Global["api_base"] = env.Str(apiBase)
		base.Global["artifact_dir"] = env.Str(cfg.Migration.ArtifactDir)

		// Initialize migrator (apimigrate) with env pointer per apimigrate v16 example style
		m := apirun.Migrator{Dir: dir, StoreConfig: &dbConfig, SaveResponseBody: true, TLSConfig: common.BuildDefaultClientTLS()}
		m.Env = base

		m.TLSConfig.InsecureSkipVerify = true
		m.TLSConfig.MinVersion = tls.VersionTLS11
		m.TLSConfig.MaxVersion = tls.VersionTLS13
		// Generate temporary credentials
		username := "migration"
		tmpPass, genErr := func() (string, error) {
			b := make([]byte, 16)
			if _, err := rand.Read(b); err != nil {
				return "", err
			}
			return hex.EncodeToString(b), nil
		}()
		if genErr != nil {
			slog.Error("migration: failed to generate temporary password", "error", genErr)
			return
		}

		// Create user attributes in new structure: {roles: {role:super_admin: true}, info: {}}
		attrs := map[string]any{
			common.ExtraKeyRoles: map[string]bool{
				string(role.RoleSuperAdmin): true, // role.RoleSuperAdmin = "role:super_admin"
			},
			common.ExtraKeyInfo: map[string]any{},
		}
		var prepare []string
		if err := authhandler.CreateUser(username, tmpPass, attrs, &prepare); err != nil {
			slog.Error("migration: temporary user create failed", "error", err)
			return
		}
		defer func() {
			if err := authhandler.DeleteUser(username); err != nil {
				slog.Info("migration: temporary user delete skipped or failed", "user", username, "error", err)
			}
		}()

		// proceed to set up OAuth2 password flow
		tokenURL := apiBase + "/auth/token"
		if len(cfg.Auth.AllowedClients) == 0 {
			slog.Error("migration: auth acquire failed", "token_url", tokenURL, "error", "no allowed clients configured")
			return
		}
		clientId := cfg.Auth.AllowedClients[0]
		pwdCfg := apirun.OAuth2PasswordConfig{
			TokenURL: tokenURL,
			ClientID: clientId,
			Username: username,
			Password: tmpPass,
		}
		auth := apirun.Auth{
			Type:    apirun.AuthTypeOAuth2,
			Name:    "migration",
			Methods: pwdCfg,
		}
		m.Auth = []apirun.Auth{auth}

		slog.Info("migration: starting", "dir", dir, "env", base, "api_base", apiBase)

		vres, err := m.MigrateUp(ctx, 0)
		if err != nil {
			slog.Error("migration: up failed", "error", err)
			return
		}
		// Log brief per-version result info if available
		for _, vr := range vres {
			if vr != nil && vr.Result != nil {
				slog.Info("migration: applied",
					"version", vr.Version,
					"status", vr.Result.StatusCode,
					"env", vr.Result.ExtractedEnv,
				)
			}
		}
		slog.Info("migration: completed successfully")
	}()
}
