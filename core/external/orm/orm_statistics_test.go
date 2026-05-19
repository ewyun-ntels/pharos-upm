package orm

import (
	"os"
	"path/filepath"
	"strconv"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/stretchr/testify/require"
)

// helper to create a temporary SQLite DB file
func createTempStatsSQLiteDB(tb testing.TB) string {
	tb.Helper()
	tempDir := tb.TempDir()
	return filepath.Join(tempDir, "stats.db")
}

// Test that the statistics pool can connect to SQLite and run a trivial query
func TestStatisticsHandler_SQLite_Integration(t *testing.T) {
	dbPath := createTempStatsSQLiteDB(t)

	cfg := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{Path: dbPath},
	}

	err := StatisticsHandler(DriverDefault, cfg, func(db *sqlx.DB) error {
		_, err := db.Exec("SELECT 1")
		return err
	})
	require.NoError(t, err)
}

// Integration test for PostgreSQL statistics pool. Skips if env not provided.
// Set the following env vars to enable:
//
//	PHAROS_TEST_PG_HOST, PHAROS_TEST_PG_PORT, PHAROS_TEST_PG_USER, PHAROS_TEST_PG_PASS, PHAROS_TEST_PG_DB
func TestStatisticsHandler_PostgreSQL_Connect(t *testing.T) {
	host := os.Getenv("PHAROS_TEST_PG_HOST")
	port := os.Getenv("PHAROS_TEST_PG_PORT")
	user := os.Getenv("PHAROS_TEST_PG_USER")
	pass := os.Getenv("PHAROS_TEST_PG_PASS")
	dbname := os.Getenv("PHAROS_TEST_PG_DB")

	if host == "" || port == "" || user == "" || dbname == "" {
		t.Skip("PostgreSQL integration env not set; skipping")
	}

	// default sensible values
	pgPort := 5432
	if p, ok := os.LookupEnv("PHAROS_TEST_PG_PORT"); ok && p != "" {
		// ignore parse error and keep default if any
		if v, err := strconv.Atoi(p); err == nil {
			pgPort = v
		}
	}

	cfg := &DatabaseConfig{
		Driver: DriverPostgreSQL,
		PostgreSQL: PostgreSQLConfig{
			Host:     host,
			Port:     pgPort,
			Username: user,
			Password: pass,
			Database: dbname,
		},
	}

	err := StatisticsHandler(DriverDefault, cfg, func(db *sqlx.DB) error {
		return db.Ping()
	})
	require.NoError(t, err)
}

// Integration test for ClickHouse statistics pool. Skips if env not provided.
// Set the following env vars to enable:
//
//	PHAROS_TEST_CH_HOST, PHAROS_TEST_CH_PORT, PHAROS_TEST_CH_USER, PHAROS_TEST_CH_PASS, PHAROS_TEST_CH_DB
func TestStatisticsHandler_ClickHouse_Connect(t *testing.T) {
	host := os.Getenv("PHAROS_TEST_CH_HOST")
	port := os.Getenv("PHAROS_TEST_CH_PORT")
	user := os.Getenv("PHAROS_TEST_CH_USER")
	pass := os.Getenv("PHAROS_TEST_CH_PASS")
	dbname := os.Getenv("PHAROS_TEST_CH_DB")

	if host == "" || port == "" || user == "" || dbname == "" {
		t.Skip("ClickHouse integration env not set; skipping")
	}

	chPort := 9000
	if p, ok := os.LookupEnv("PHAROS_TEST_CH_PORT"); ok && p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			chPort = v
		}
	}

	cfg := &DatabaseConfig{
		Driver: DriverClickHouse,
		ClickHouse: ClickHouseConfig{
			Host:     host,
			Port:     chPort,
			Username: user,
			Password: pass,
			Database: dbname,
		},
	}

	err := StatisticsHandler(DriverDefault, cfg, func(db *sqlx.DB) error {
		return db.Ping()
	})
	require.NoError(t, err)
}

// Integration test for Vertica statistics pool. Skips if env not provided.
// Set the following env vars to enable:
//
//	PHAROS_TEST_VERTICA_HOST, PHAROS_TEST_VERTICA_PORT, PHAROS_TEST_VERTICA_USER, PHAROS_TEST_VERTICA_PASS, PHAROS_TEST_VERTICA_DB
func TestStatisticsHandler_Vertica_Connect(t *testing.T) {
	host := os.Getenv("PHAROS_TEST_VERTICA_HOST")
	port := os.Getenv("PHAROS_TEST_VERTICA_PORT")
	user := os.Getenv("PHAROS_TEST_VERTICA_USER")
	pass := os.Getenv("PHAROS_TEST_VERTICA_PASS")
	dbname := os.Getenv("PHAROS_TEST_VERTICA_DB")

	if host == "" || port == "" || user == "" || dbname == "" {
		t.Skip("Vertica integration env not set; skipping")
	}

	verticaPort := 5433
	if p, ok := os.LookupEnv("PHAROS_TEST_VERTICA_PORT"); ok && p != "" {
		if v, err := strconv.Atoi(p); err == nil {
			verticaPort = v
		}
	}

	cfg := &DatabaseConfig{
		Driver: DriverVertica,
		Vertica: VerticaConfig{
			Host:     host,
			Port:     verticaPort,
			Username: user,
			Password: pass,
			Database: dbname,
		},
	}

	err := StatisticsHandler(DriverDefault, cfg, func(db *sqlx.DB) error {
		return db.Ping()
	})
	require.NoError(t, err)
}

// Test StatisticsHandler vs HandlerWithRole equivalence
func TestStatisticsHandler_EquivalentToHandlerWithRole(t *testing.T) {
	dbPath := createTempStatsSQLiteDB(t)

	cfg := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{Path: dbPath},
	}

	// Create a table using StatisticsHandler
	err := StatisticsHandler(DriverDefault, cfg, func(db *sqlx.DB) error {
		_, err := db.Exec("CREATE TABLE test_equivalence (id INTEGER PRIMARY KEY, name TEXT)")
		return err
	})
	require.NoError(t, err)

	// Verify the table exists using HandlerWithRole with PoolStatistics
	err = HandlerWithRole(PoolStatistics, DriverDefault, cfg, func(db *sqlx.DB) error {
		var count int
		err := db.Get(&count, "SELECT COUNT(*) FROM sqlite_master WHERE type='table' AND name='test_equivalence'")
		require.NoError(t, err)
		require.Equal(t, 1, count, "Table should exist, confirming StatisticsHandler uses statistics pool")
		return nil
	})
	require.NoError(t, err)
}

// Test multiple database drivers with StatisticsHandler
func TestStatisticsHandler_MultipleDrivers(t *testing.T) {
	dbPath := createTempStatsSQLiteDB(t)

	tests := []struct {
		name     string
		driver   string
		config   *DatabaseConfig
		skipTest bool
		skipMsg  string
	}{
		{
			name:   "SQLite",
			driver: DriverSqlite,
			config: &DatabaseConfig{
				Driver: DriverSqlite,
				SQLite: SQLiteConfig{Path: dbPath},
			},
			skipTest: false,
		},
		{
			name:   "PostgreSQL",
			driver: DriverPostgreSQL,
			config: &DatabaseConfig{
				Driver: DriverPostgreSQL,
				PostgreSQL: PostgreSQLConfig{
					Host:     "localhost",
					Port:     5432,
					Username: "test",
					Password: "test",
					Database: "test",
				},
			},
			skipTest: true,
			skipMsg:  "PostgreSQL not available in test environment",
		},
		{
			name:   "ClickHouse",
			driver: DriverClickHouse,
			config: &DatabaseConfig{
				Driver: DriverClickHouse,
				ClickHouse: ClickHouseConfig{
					Host:     "localhost",
					Port:     9000,
					Username: "default",
					Password: "",
					Database: "default",
				},
			},
			skipTest: true,
			skipMsg:  "ClickHouse not available in test environment",
		},
		{
			name:   "Vertica",
			driver: DriverVertica,
			config: &DatabaseConfig{
				Driver: DriverVertica,
				Vertica: VerticaConfig{
					Host:     "localhost",
					Port:     5433,
					Username: "dbadmin",
					Password: "password",
					Database: "test",
				},
			},
			skipTest: true,
			skipMsg:  "Vertica not available in test environment",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.skipTest {
				t.Skip(tt.skipMsg)
			}

			err := StatisticsHandler(DriverDefault, tt.config, func(db *sqlx.DB) error {
				// Simple query to verify connection works
				_, err := db.Exec("SELECT 1")
				return err
			})

			if tt.driver == DriverSqlite {
				// SQLite should work
				require.NoError(t, err)
			} else {
				// Other drivers may fail due to unavailable servers, but that's expected
				t.Logf("Driver %s result: %v", tt.driver, err)
			}
		})
	}
}

// Benchmark StatisticsHandler
func BenchmarkStatisticsHandler_SQLite(b *testing.B) {
	dbPath := createTempStatsSQLiteDB(b)

	cfg := &DatabaseConfig{
		Driver: DriverSqlite,
		SQLite: SQLiteConfig{Path: dbPath},
	}

	handlerFunc := func(db *sqlx.DB) error {
		_, err := db.Exec("SELECT 1")
		return err
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		err := StatisticsHandler(DriverDefault, cfg, handlerFunc)
		if err != nil {
			b.Fatal(err)
		}
	}
}
