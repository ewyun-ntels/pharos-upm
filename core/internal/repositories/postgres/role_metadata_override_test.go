package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/jmoiron/sqlx"
	"github.com/moby/moby/api/types/network"
	"github.com/testcontainers/testcontainers-go"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories/test"
	"ntels.com/pharos/core/internal/testutil"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/migration/schema"
)

var _ test.RoleMetadataConfigTestable = (*TestPostgresRoleMetadataConfig)(nil)

type TestPostgresRoleMetadataConfig struct {
	roleMetadataConfigRepository

	testDatabase    string
	testDBHost      string
	testDBPort      int
	testDBUser      string
	testDBPassword  string
	defaultDatabase string

	config *common.Config
}

// TestRoleMetadataConfigWrapper wraps TestPostgresRoleMetadataConfig to avoid repeated DB setup/teardown
type TestRoleMetadataConfigWrapper struct {
	*TestPostgresRoleMetadataConfig
}

func (w *TestRoleMetadataConfigWrapper) SetUpDb() error {
	// DB는 이미 설정되어 있으므로 테이블만 초기화
	return w.ClearTable()
}

func (w *TestRoleMetadataConfigWrapper) TearDownDb() error {
	// DB는 유지하고 테이블만 초기화
	return w.ClearTable()
}

func (repo *TestPostgresRoleMetadataConfig) SetUpDb() error {
	if repo.config != nil {
		return errors.New("test db already set up")
	}

	repo.config = &common.Config{
		Serve: common.ServeConfig{},
		Database: orm.DatabaseConfig{
			Driver: orm.DriverPostgreSQL,
			PostgreSQL: orm.PostgreSQLConfig{
				Host:              repo.testDBHost,
				Port:              repo.testDBPort,
				Username:          repo.testDBUser,
				Password:          repo.testDBPassword,
				Database:          repo.defaultDatabase,
				MaxOpenConnection: 1,
			},
		},
	}
	driverName := repo.config.Database.GetDriverName()
	dsn := repo.config.Database.GetDataSourceName()

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("failed to open connect to default database: %w", err)
	}

	// 테스트용 DB 존재 확인
	row := db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1);", repo.testDatabase)
	if row.Err() != nil {
		_ = db.Close()
		return fmt.Errorf("failed to check existence of database %s: %w", repo.testDatabase, row.Err())
	}
	exists := false
	err = row.Scan(&exists)
	if err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to scan existence result for database %s: %w", repo.testDatabase, err)
	} else if exists {
		_ = db.Close()
		return fmt.Errorf("database %s already exists", repo.testDatabase)
	}

	_, err = db.Exec("CREATE DATABASE " + repo.testDatabase)
	if err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to create test database %s: %w", repo.testDatabase, err)
	}
	_ = db.Close()

	// 테스트용 DB로 접속
	repo.config.Database.PostgreSQL.Database = repo.testDatabase

	// 마이그레이션 수행
	schema.DisableMigrateLog()
	err = schema.UpMasterBase(*repo.config)
	if err != nil {
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	repo.roleMetadataConfigRepository = *newRoleMetadataConfigRepository(repo.config.Database)
	return nil
}

func (repo *TestPostgresRoleMetadataConfig) TearDownDb() error {
	if repo.config != nil {
		orm.Unload()

		repo.config.Database.PostgreSQL.Database = repo.defaultDatabase
		driver := repo.config.Database.GetDriverName()
		dsn := repo.config.Database.GetDataSourceName()
		db, err := sql.Open(driver, dsn)
		if err != nil {
			return err
		}
		_, err = db.Exec("DROP DATABASE IF EXISTS " + repo.testDatabase)
		if err != nil {
			_ = db.Close()
			return err
		}
		_ = db.Close()
		repo.config = nil

		return err
	}
	return nil
}

func (repo *TestPostgresRoleMetadataConfig) ClearTable() error {
	if repo.config == nil {
		return fmt.Errorf("test db not initialized")
	}
	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.Exec("DELETE FROM role_metadata_config_history WHERE true")
		return err
	})
}

func Test_postgres_RoleMetadataConfig(t *testing.T) {
	useLocal := os.Getenv("USE_LOCAL_DB") == "true"

	testable := &TestPostgresRoleMetadataConfig{
		defaultDatabase: "postgres",
		testDatabase:    "pharos_test_role_metadata_config",
	}

	if useLocal {
		testable.testDBHost = os.Getenv("TEST_DB_HOST")
		if testable.testDBHost == "" {
			testable.testDBHost = "localhost"
		}
		portString := os.Getenv("TEST_DB_PORT")
		if portString != "" {
			port, err := strconv.Atoi(portString)
			if err != nil {
				t.Fatalf("Invalid TEST_DB_PORT: %v", err)
			}
			testable.testDBPort = port
		} else {
			testable.testDBPort = 5432
		}
		testable.testDBUser = os.Getenv("TEST_DB_USER")
		if testable.testDBUser == "" {
			testable.testDBUser = "postgres"
		}
		testable.testDBPassword = os.Getenv("TEST_DB_PASSWORD")
		if testable.testDBPassword == "" {
			testable.testDBPassword = "postgres"
		}
	} else {
		ctx := context.Background()

		testable.testDBUser = "postgres"
		testable.testDBPassword = "postgres"
		container, err := testutil.StartPGContainer(ctx, testable.defaultDatabase, testable.testDBUser, testable.testDBPassword)
		if err != nil {
			t.Fatalf("Failed to start PostgreSQL container: %v", err)
		}
		t.Cleanup(func() {
			if err := testcontainers.TerminateContainer(container); err != nil {
				t.Logf("failed to terminate container: %s", err)
			}
		})

		testable.testDBHost, err = container.Host(ctx)
		if err != nil {
			t.Fatalf("Failed to get container host: %v", err)
		}
		var port network.Port
		port, err = container.MappedPort(ctx, "5432/tcp")
		if err != nil {
			t.Fatalf("Failed to get mapped port: %v", err)
		}
		testable.testDBPort = int(port.Num())
	}

	// 데이터베이스를 한 번만 설정하고 모든 테스트에서 재사용
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("Failed to set up test database: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("Failed to tear down test database: %v", err)
		}
	})

	// 래퍼를 사용하여 각 서브테스트에서 DB를 재설정하지 않고 테이블만 초기화
	wrapper := &TestRoleMetadataConfigWrapper{TestPostgresRoleMetadataConfig: testable}

	// Run common test suite
	test.RoleMetadataConfigSuite(t, wrapper)
}
