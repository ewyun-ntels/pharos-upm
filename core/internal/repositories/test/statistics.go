package test

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
	"ntels.com/pharos/core/internal/testutil"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/migration/schema"
)

type StatisticsTestable struct {
	TestDb   *sqlx.DB
	DbConfig orm.DatabaseConfig

	keepConn *sql.Conn
}

func (repo *StatisticsTestable) SetUpDb() error {
	if repo.TestDb != nil {
		return errors.New("statistics test db already set up")
	}

	driverName := "sqlite"
	dsn := "file:sharedmem_stats?mode=memory&cache=shared"

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("failed to open statistics test db: %w", err)
	}
	repo.keepConn, err = db.Conn(context.Background())
	if err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to keep statistics test db connection: %w", err)
	}

	config := common.Config{
		Statistics: common.StatisticsConfig{
			Database: orm.DatabaseConfig{
				Driver: driverName,
				SQLite: orm.SQLiteConfig{
					Path: dsn,
				},
			},
		},
	}

	schema.DisableMigrateLog()
	if err = schema.UpMasterStatistics(config); err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to run statistics migrations: %w", err)
	}

	repo.TestDb = sqlx.NewDb(db, driverName)
	repo.DbConfig = config.Statistics.Database
	return nil
}

func (repo *StatisticsTestable) TearDownDb() error {
	if repo.TestDb != nil {
		_ = repo.keepConn.Close()
		repo.keepConn = nil
		err := repo.TestDb.Close()
		repo.TestDb = nil
		orm.Unload()
		return err
	}
	return nil
}

// PostgresStatisticsTestable provides a PostgreSQL-backed statistics DB for tests.
// Call Connect(t) first, then embed this struct and delegate SetUpDb/TearDownDb.
type PostgresStatisticsTestable struct {
	TestDb   *sqlx.DB
	DbConfig orm.DatabaseConfig

	DefaultDatabase string
	TestDatabase    string
	testDBHost      string
	testDBPort      int
	testDBUser      string
	testDBPassword  string
	config          *common.Config
}

func (repo *PostgresStatisticsTestable) Connect(t *testing.T) {
	t.Helper()
	useLocal := os.Getenv("USE_LOCAL_DB") == "true"

	if useLocal {
		repo.testDBHost = os.Getenv("TEST_DB_HOST")
		if repo.testDBHost == "" {
			repo.testDBHost = "localhost"
		}
		portString := os.Getenv("TEST_DB_PORT")
		if portString != "" {
			port, err := strconv.Atoi(portString)
			if err != nil {
				t.Fatalf("Invalid TEST_DB_PORT: %v", err)
			}
			repo.testDBPort = port
		} else {
			repo.testDBPort = 5432
		}
		repo.testDBUser = os.Getenv("TEST_DB_USER")
		if repo.testDBUser == "" {
			repo.testDBUser = "postgres"
		}
		repo.testDBPassword = os.Getenv("TEST_DB_PASSWORD")
		if repo.testDBPassword == "" {
			repo.testDBPassword = "postgres"
		}
	} else {
		ctx := context.Background()
		repo.testDBUser = "postgres"
		repo.testDBPassword = "postgres"
		container, err := testutil.StartPGContainer(ctx, repo.DefaultDatabase, repo.testDBUser, repo.testDBPassword)
		if err != nil {
			t.Fatalf("Failed to start PostgreSQL container: %v", err)
		}
		t.Cleanup(func() {
			if err := testcontainers.TerminateContainer(container); err != nil {
				t.Logf("failed to terminate container: %s", err)
			}
		})
		repo.testDBHost, err = container.Host(ctx)
		if err != nil {
			t.Fatalf("Failed to get container host: %v", err)
		}
		var port network.Port
		port, err = container.MappedPort(ctx, "5432/tcp")
		if err != nil {
			t.Fatalf("Failed to get mapped port: %v", err)
		}
		repo.testDBPort = int(port.Num())
	}
}

func (repo *PostgresStatisticsTestable) SetUpDb() error {
	if repo.config != nil {
		return errors.New("statistics postgres test db already set up")
	}

	repo.config = &common.Config{
		Statistics: common.StatisticsConfig{
			Database: orm.DatabaseConfig{
				Driver: orm.DriverPostgreSQL,
				PostgreSQL: orm.PostgreSQLConfig{
					Host:              repo.testDBHost,
					Port:              repo.testDBPort,
					Username:          repo.testDBUser,
					Password:          repo.testDBPassword,
					Database:          repo.DefaultDatabase,
					MaxOpenConnection: 1,
				},
			},
		},
	}

	driverName := repo.config.Statistics.Database.GetDriverName()
	dsn := repo.config.Statistics.Database.GetDataSourceName()

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("01 failed to connect to default database: %w", err)
	}

	row := db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", repo.TestDatabase)
	if row.Err() != nil {
		_ = db.Close()
		return fmt.Errorf("02 failed to check database existence: %w", row.Err())
	}
	exists := false
	if err = row.Scan(&exists); err != nil {
		_ = db.Close()
		return fmt.Errorf("03 failed to scan database existence: %w", err)
	}
	if exists {
		_ = db.Close()
		return fmt.Errorf("04 database %s already exists", repo.TestDatabase)
	}

	if _, err = db.Exec("CREATE DATABASE " + repo.TestDatabase); err != nil {
		_ = db.Close()
		return fmt.Errorf("05 failed to create test database: %w", err)
	}
	_ = db.Close()

	repo.config.Statistics.Database.PostgreSQL.Database = repo.TestDatabase
	err = orm.StatisticsHandler("", &repo.config.Statistics.Database, func(db *sqlx.DB) error {
		_, dbErr := db.Exec(`SET TIME ZONE 'UTC'`)
		return dbErr
	})
	if err != nil {
		return fmt.Errorf("06 failed to set timezone: %w", err)
	}

	schema.DisableMigrateLog()
	if err = schema.UpMasterStatistics(*repo.config); err != nil {
		return fmt.Errorf("07 failed to run statistics migrations: %w", err)
	}

	testDSN := repo.config.Statistics.Database.GetDataSourceName()
	rawDB, err := sql.Open(driverName, testDSN)
	if err != nil {
		return fmt.Errorf("08 failed to open test database direct connection: %w", err)
	}
	repo.TestDb = sqlx.NewDb(rawDB, driverName)
	repo.DbConfig = repo.config.Statistics.Database
	return nil
}

func (repo *PostgresStatisticsTestable) TearDownDb() error {
	if repo.config == nil {
		return nil
	}
	if repo.TestDb != nil {
		_ = repo.TestDb.Close()
		repo.TestDb = nil
	}
	orm.Unload()

	repo.config.Statistics.Database.PostgreSQL.Database = repo.DefaultDatabase
	driver := repo.config.Statistics.Database.GetDriverName()
	dsn := repo.config.Statistics.Database.GetDataSourceName()
	db, err := sql.Open(driver, dsn)
	if err != nil {
		repo.config = nil
		return err
	}
	_, err = db.Exec("DROP DATABASE IF EXISTS " + repo.TestDatabase)
	_ = db.Close()
	repo.config = nil
	return err
}

// ClickHouseStatisticsTestable provides a ClickHouse-backed statistics DB for tests.
// Call Connect(t) first, then embed this struct and delegate SetUpDb/TearDownDb.
type ClickHouseStatisticsTestable struct {
	DbConfig orm.DatabaseConfig

	testDBHost     string
	testDBPort     int
	testDBUser     string
	testDBPassword string
	testDBName     string
	config         *common.Config
}

func (repo *ClickHouseStatisticsTestable) Connect(t *testing.T) {
	t.Helper()
	useLocal := os.Getenv("USE_LOCAL_DB") == "true"

	if useLocal {
		repo.testDBHost = os.Getenv("PHAROS_TEST_CH_HOST")
		if repo.testDBHost == "" {
			repo.testDBHost = "localhost"
		}
		portStr := os.Getenv("PHAROS_TEST_CH_PORT")
		if portStr != "" {
			port, err := strconv.Atoi(portStr)
			if err != nil {
				t.Fatalf("Invalid PHAROS_TEST_CH_PORT: %v", err)
			}
			repo.testDBPort = port
		} else {
			repo.testDBPort = 9000
		}
		repo.testDBUser = os.Getenv("PHAROS_TEST_CH_USER")
		if repo.testDBUser == "" {
			repo.testDBUser = "default"
		}
		repo.testDBPassword = os.Getenv("PHAROS_TEST_CH_PASS")
		repo.testDBName = os.Getenv("PHAROS_TEST_CH_DB")
		if repo.testDBName == "" {
			repo.testDBName = "default"
		}
	} else {
		ctx := context.Background()
		repo.testDBUser = "default"
		repo.testDBPassword = ""
		repo.testDBName = "default"
		container, err := testutil.StartClickHouseContainer(ctx, repo.testDBName, repo.testDBUser, repo.testDBPassword)
		if err != nil {
			t.Fatalf("Failed to start ClickHouse container: %v", err)
		}
		t.Cleanup(func() {
			if err := testcontainers.TerminateContainer(container); err != nil {
				t.Logf("failed to terminate clickhouse container: %s", err)
			}
		})
		repo.testDBHost, err = container.Host(ctx)
		if err != nil {
			t.Fatalf("Failed to get ClickHouse container host: %v", err)
		}
		var port network.Port
		port, err = container.MappedPort(ctx, "9000/tcp")
		if err != nil {
			t.Fatalf("Failed to get ClickHouse mapped port: %v", err)
		}
		repo.testDBPort = int(port.Num())
	}
}

func (repo *ClickHouseStatisticsTestable) SetUpDb() error {
	if repo.config != nil {
		return errors.New("clickhouse statistics test db already set up")
	}

	repo.config = &common.Config{
		Statistics: common.StatisticsConfig{
			Database: orm.DatabaseConfig{
				Driver: orm.DriverClickHouse,
				ClickHouse: orm.ClickHouseConfig{
					Host:              repo.testDBHost,
					Port:              repo.testDBPort,
					Username:          repo.testDBUser,
					Password:          repo.testDBPassword,
					Database:          repo.testDBName,
					MaxOpenConnection: 1,
				},
			},
		},
	}

	schema.DisableMigrateLog()
	if err := schema.UpMasterStatistics(*repo.config); err != nil {
		return fmt.Errorf("failed to run clickhouse statistics migrations: %w", err)
	}

	repo.DbConfig = repo.config.Statistics.Database
	return nil
}

func (repo *ClickHouseStatisticsTestable) TearDownDb() error {
	if repo.config == nil {
		return nil
	}
	_ = orm.StatisticsHandler("", &repo.DbConfig, func(db *sqlx.DB) error {
		_, err := db.Exec("TRUNCATE TABLE history_login")
		return err
	})
	orm.Unload()
	repo.config = nil
	repo.DbConfig = orm.DatabaseConfig{}
	return nil
}
