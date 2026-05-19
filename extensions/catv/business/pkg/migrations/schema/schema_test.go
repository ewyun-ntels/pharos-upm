package schema

import (
	"testing"

	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	master_statistics "ntels.com/pharos/extensions/catv/business/pkg/migrations/schema/master/statistics"
)

func TestLoad(t *testing.T) {
	tests := []struct {
		name   string
		config common.Config
	}{
		{
			name: "clickhouse",
			config: common.Config{
				Serve: common.ServeConfig{},
				Catv: common.CatvConfig{
					Migration: struct {
						Database orm.DatabaseConfig "mapstructure:\"database\""
					}{
						Database: orm.DatabaseConfig{
							Driver: orm.DriverClickHouse,
							ClickHouse: orm.ClickHouseConfig{
								Host:     "localhost",
								Port:     9000,
								Database: "catv",
							},
						},
					},
				},
			},
		},
		{
			name: "postgres",
			config: common.Config{
				Serve: common.ServeConfig{},
				Catv: common.CatvConfig{
					Migration: struct {
						Database orm.DatabaseConfig "mapstructure:\"database\""
					}{
						Database: orm.DatabaseConfig{
							Driver: orm.DriverPostgreSQL,
							PostgreSQL: orm.PostgreSQLConfig{
								Host:     "localhost",
								Port:     5432,
								Database: "catv",
							},
						},
					},
				},
			},
		},
		{
			name:   "empty config",
			config: common.Config{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Should not panic
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Load() panicked: %v", r)
				}
			}()

			Load(tt.config)
		})
	}
}

func TestLoadMasterFS(t *testing.T) {
	// Test that masterFS is properly embedded
	entries, err := masterFS.ReadDir("master/statistics/clickhouse")
	if err != nil {
		t.Fatalf("Failed to read embedded masterFS: %v", err)
	}

	if len(entries) == 0 {
		t.Error("Expected embedded files in masterFS, got none")
	}

	t.Logf("Found %d embedded files", len(entries))
	for _, entry := range entries {
		t.Logf("  - %s", entry.Name())
	}
}

func TestMasterStatisticsConfiguration(t *testing.T) {
	// Test that master statistics configuration is correct
	if master_statistics.TableName == "" {
		t.Error("master_statistics.TableName should not be empty")
	}

	expectedTableName := "extensions_catv_statistics_db_version"
	if master_statistics.TableName != expectedTableName {
		t.Errorf("Expected TableName to be %q, got %q", expectedTableName, master_statistics.TableName)
	}

	// Check that ClickHouse migrations are registered
	migrations, ok := master_statistics.Migrations[orm.DriverClickHouse]
	if !ok {
		t.Error("Expected ClickHouse migrations to be registered")
	}

	if len(migrations) == 0 {
		t.Error("Expected at least one ClickHouse migration")
	}

	t.Logf("Found %d ClickHouse migrations", len(migrations))

	// Verify first migration
	if len(migrations) > 0 {
		firstMigration := migrations[0]
		if firstMigration == nil {
			t.Error("First migration should not be nil")
			return
		}
		if firstMigration.Version == 0 {
			t.Error("First migration version should not be 0")
		}
		t.Logf("First migration version: %d", firstMigration.Version)
	}
}

func TestLoadWithDifferentDrivers(t *testing.T) {
	drivers := []struct {
		name   string
		driver string
	}{
		{"clickhouse", orm.DriverClickHouse},
		{"postgresql", orm.DriverPostgreSQL},
		{"sqlite", orm.DriverSqlite},
		{"vertica", orm.DriverVertica},
	}

	for _, d := range drivers {
		t.Run(d.name, func(t *testing.T) {
			config := common.Config{
				Serve: common.ServeConfig{},
				Catv: common.CatvConfig{
					Migration: struct {
						Database orm.DatabaseConfig "mapstructure:\"database\""
					}{
						Database: orm.DatabaseConfig{
							Driver: d.driver,
						},
					},
				},
			}

			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Load() panicked for %s driver: %v", d.name, r)
				}
			}()

			Load(config)
		})
	}
}

func TestEmbeddedMigrationsStructure(t *testing.T) {
	// Test the structure of embedded migrations
	expectedDir := "master/statistics/clickhouse"

	entries, err := masterFS.ReadDir(expectedDir)
	if err != nil {
		t.Fatalf("Failed to read embedded directory %q: %v", expectedDir, err)
	}

	// Check for at least one migration file
	hasMigrationFile := false
	for _, entry := range entries {
		if !entry.IsDir() && len(entry.Name()) > 0 {
			hasMigrationFile = true
			t.Logf("Found migration file: %s", entry.Name())
		}
	}

	if !hasMigrationFile {
		t.Error("Expected at least one migration file in embedded FS")
	}
}

func BenchmarkLoad(b *testing.B) {
	config := common.Config{
		Serve: common.ServeConfig{},
		Catv: common.CatvConfig{
			Migration: struct {
				Database orm.DatabaseConfig "mapstructure:\"database\""
			}{
				Database: orm.DatabaseConfig{
					Driver: orm.DriverClickHouse,
					ClickHouse: orm.ClickHouseConfig{
						Host:     "localhost",
						Port:     9000,
						Database: "catv",
					},
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Load(config)
	}
}

func BenchmarkLoadWithAgent(b *testing.B) {
	config := common.Config{
		Serve: common.ServeConfig{},
		Catv: common.CatvConfig{
			Migration: struct {
				Database orm.DatabaseConfig "mapstructure:\"database\""
			}{
				Database: orm.DatabaseConfig{
					Driver: orm.DriverClickHouse,
				},
			},
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		Load(config)
	}
}
