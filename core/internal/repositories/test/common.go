package test

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"github.com/moby/moby/api/types/network"
	"github.com/testcontainers/testcontainers-go"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/testutil"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/migration/schema"
)

// testTimeZone UTC+03:00
var testTimeZone = time.FixedZone("test-zone", 3*60*60)
var maximumAllowedTimeDeltaForNow = 10 * time.Second

type SqliteTestable struct {
	TestDb   *sqlx.DB
	DbConfig orm.DatabaseConfig

	keepConn *sql.Conn
}

func (repo *SqliteTestable) SetUpDb() error {
	if repo.TestDb != nil {
		return errors.New("test db already set up")
	}
	driverName := "sqlite"
	dsn := "file:sharedmem?mode=memory&cache=shared"
	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return err
	}
	repo.keepConn, err = db.Conn(context.Background())
	if err != nil {
		_ = db.Close()
		return err
	}

	config := common.Config{
		Serve: common.ServeConfig{},
		Database: orm.DatabaseConfig{
			Driver: driverName,
			SQLite: orm.SQLiteConfig{
				Path: dsn,
			},
		},
	}

	schema.DisableMigrateLog()
	err = schema.UpMasterBase(config)
	if err != nil {
		_ = db.Close()
		return fmt.Errorf("failed to run migrations: %w", err)
	}

	repo.TestDb = sqlx.NewDb(db, driverName)
	repo.DbConfig = config.Database
	return nil
}
func (repo *SqliteTestable) TearDownDb() error {
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

type PostgresTestable struct {
	DefaultDatabase string
	TestDatabase    string
	Config          *common.Config

	testDBHost     string
	testDBPort     int
	testDBUser     string
	testDBPassword string
}

func (repo *PostgresTestable) Connect(t *testing.T) error {
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

	return nil
}
func (repo *PostgresTestable) SetUpDb() error {
	if repo.Config != nil {
		return errors.New("test db already set up")
	}

	repo.Config = &common.Config{
		Serve: common.ServeConfig{},
		Database: orm.DatabaseConfig{
			Driver: orm.DriverPostgreSQL,
			PostgreSQL: orm.PostgreSQLConfig{
				Host:     repo.testDBHost,
				Port:     repo.testDBPort,
				Username: repo.testDBUser,
				Password: repo.testDBPassword,
				Database: repo.DefaultDatabase,
				// 타임존 테스트를 위해서 1개의 세션만 사용하도록 설정
				MaxOpenConnection: 1,
			},
		},
	}
	driverName := repo.Config.Database.GetDriverName()
	dsn := repo.Config.Database.GetDataSourceName()

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("01 failed to open connect to default database: %w", err)
	}

	// 테스트용 DB를 쓰기 위해 존재하는지 확인
	// 만약 존재하면 에러, 존재하지 않으면 생성
	row := db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1);", repo.TestDatabase)
	if row.Err() != nil {
		_ = db.Close()
		return fmt.Errorf("02 failed to check existence of database %s: %w", repo.TestDatabase, row.Err())
	}
	exists := false
	err = row.Scan(&exists)
	if err != nil {
		_ = db.Close()
		return fmt.Errorf("03 failed to scan existence result for database %s: %w", repo.TestDatabase, err)
	} else if exists {
		_ = db.Close()
		return fmt.Errorf("04 database %s already exists", repo.TestDatabase)
	}

	_, err = db.Exec("CREATE DATABASE " + repo.TestDatabase)
	if err != nil {
		_ = db.Close()
		return fmt.Errorf("05 failed to create test database %s: %w", repo.TestDatabase, err)
	}
	_ = db.Close()

	// 테스트용 DB로 접속
	repo.Config.Database.PostgreSQL.Database = repo.TestDatabase
	err = orm.Handler("", &repo.Config.Database, func(db *sqlx.DB) error {
		// 세션 타임존 설정
		// Golang의 time 타입은 타임존 정보를 포함하지만
		// PostgreSQL에서는 타임존 정보가 없는 TIMESTAMP WITHOUT TIME ZONE
		// 타입을 사용하고 서버 또는 세션의 타임존을 기준으로 사용한다.
		// 그러나 GOlang의 time.Time 타입을 삽입할 때 타임존 정보를 명시하지 않으면
		// UTC로 변환되어 삽입된다.
		// 예를 들어, Asia/Seoul 타임존에서 2024-01-01 10:00:00 (KST) 시간을 삽입하면
		// DB에는 2024-01-01 10:00:00 으로 저장된다. 이것을 Golang에서 조회하면
		// 2024-01-01 10:00:00 (UTC)로 해석되어 9시간 차이가 발생한다.
		// 따라서, 세션 타임존을 Asia/Seoul로 설정하여 TIMESTAMP 필드에 삽입되는 시간이
		// 코드상에서 명시하지 않으면 Asia/Seoul(예상치 못한) 타임존 기준이 되도록 한다.
		_, dbErr := db.Exec(`SET TIME ZONE 'Asia/Seoul'`)
		return dbErr
	})
	if err != nil {
		return fmt.Errorf("06 failed to set session time zone: %w", err)
	}

	// 마이그레이션 수행
	schema.DisableMigrateLog()
	err = schema.UpMasterBase(*repo.Config)
	if err != nil {
		_ = db.Close()
		return fmt.Errorf("07 failed to run migrations: %w", err)
	}

	return nil
}
func (repo *PostgresTestable) TearDownDb() error {
	if repo.Config != nil {
		orm.Unload()

		repo.Config.Database.PostgreSQL.Database = repo.DefaultDatabase
		driver := repo.Config.Database.GetDriverName()
		dsn := repo.Config.Database.GetDataSourceName()
		db, err := sql.Open(driver, dsn)
		if err != nil {
			return err
		}
		_, err = db.Exec("DROP DATABASE IF EXISTS " + repo.TestDatabase)
		if err != nil {
			_ = db.Close()
			return err
		}
		_ = db.Close()
		repo.Config = nil

		return err
	}
	return nil
}
