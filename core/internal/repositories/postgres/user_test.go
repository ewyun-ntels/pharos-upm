package postgres

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
	"golang.org/x/crypto/bcrypt"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/test"
	"ntels.com/pharos/core/internal/testutil"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/migration/schema"
)

var _ test.UserTestable = (*TestPostgresUser)(nil)

type TestPostgresUser struct {
	userRepository

	testDatabase    string
	testDBHost      string
	testDBPort      int
	testDBUser      string
	testDBPassword  string
	defaultDatabase string

	config *common.Config
}

// TestUserWrapper wraps TestPostgresUser to avoid repeated DB setup/teardown
type TestUserWrapper struct {
	*TestPostgresUser
}

func (w *TestUserWrapper) SetUpDb() error {
	// DB는 이미 설정되어 있으므로 테이블만 초기화
	return w.ClearTable()
}

func (w *TestUserWrapper) TearDownDb() error {
	// DB는 유지하고 테이블만 초기화
	return w.ClearTable()
}

func (repo *TestPostgresUser) SetUpDb() error {
	if repo.config != nil {
		return errors.New("test db already set up")
	}

	repo.config = &common.Config{
		Serve: common.ServeConfig{},
		Database: orm.DatabaseConfig{
			Driver: orm.DriverPostgreSQL,
			PostgreSQL: orm.PostgreSQLConfig{
				Host:     repo.testDBHost,
				Port:     repo.testDBPort,
				Username: repo.testDBUser,
				Password: repo.testDBPassword,
				Database: repo.defaultDatabase,
				// 타임존 테스트를 위해서 1개의 세션만 사용하도록 설정
				MaxOpenConnection: 1,
			},
		},
	}
	driverName := repo.config.Database.GetDriverName()
	dsn := repo.config.Database.GetDataSourceName()

	db, err := sql.Open(driverName, dsn)
	if err != nil {
		return fmt.Errorf("01 failed to open connect to default database: %w", err)
	}

	// 테스트용 DB를 쓰기 위해 존재하는지 확인
	// 만약 존재하면 에러, 존재하지 않으면 생성
	row := db.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1);", repo.testDatabase)
	if row.Err() != nil {
		_ = db.Close()
		return fmt.Errorf("02 failed to check existence of database %s: %w", repo.testDatabase, row.Err())
	}
	exists := false
	err = row.Scan(&exists)
	if err != nil {
		_ = db.Close()
		return fmt.Errorf("03 failed to scan existence result for database %s: %w", repo.testDatabase, err)
	} else if exists {
		_ = db.Close()
		return fmt.Errorf("04 database %s already exists", repo.testDatabase)
	}

	_, err = db.Exec("CREATE DATABASE " + repo.testDatabase)
	if err != nil {
		_ = db.Close()
		return fmt.Errorf("05 failed to create test database %s: %w", repo.testDatabase, err)
	}
	_ = db.Close()

	// 테스트용 DB로 접속
	repo.config.Database.PostgreSQL.Database = repo.testDatabase
	err = orm.Handler("", &repo.config.Database, func(db *sqlx.DB) error {
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
	err = schema.UpMasterBase(*repo.config)
	if err != nil {
		_ = db.Close()
		return fmt.Errorf("07 failed to run migrations: %w", err)
	}

	repo.userRepository = *newUserRepository(repo.config.Database)
	return nil
}
func (repo *TestPostgresUser) TearDownDb() error {
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
func (repo *TestPostgresUser) SetUpUser(username string) error {
	if repo.config == nil {
		return fmt.Errorf("test db not initialized")
	}
	// 테스트용 사용자 삽입
	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.Exec(`
	INSERT INTO users (username, password_hash, extra, created_at, retry, retry_at, blocked, prepare, password_expired_at)
	VALUES ($1, x'1234', '{}', CURRENT_TIMESTAMP, 0, NULL, false, NULL, NULL)
	`, username)
		return err
	})
}
func (repo *TestPostgresUser) ClearTable() error {
	if repo.config == nil {
		return fmt.Errorf("test db not initialized")
	}
	return repo.Handler(func(db *sqlx.DB) error {
		// password history 테이블도 함께 초기화
		_, err := db.Exec("DELETE FROM user_password_history WHERE true")
		if err != nil {
			return err
		}
		_, err = db.Exec("DELETE FROM users WHERE true")
		return err
	})
}
func (repo *TestPostgresUser) GetUser(name string) (*repositories.UserEntity, error) {
	if repo.config == nil {
		return nil, fmt.Errorf("test db not initialized")
	}
	var u postgresUser
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.Get(&u, "SELECT * FROM users WHERE username = $1", name)
	})
	if err != nil {
		return nil, err
	}

	return repo.toUser(u)
}
func (repo *TestPostgresUser) GetUserCount() (int, error) {
	if repo.config == nil {
		return 0, fmt.Errorf("test db not initialized")
	}
	var count int
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.Get(&count, "SELECT COUNT(*) FROM users")
	})
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (repo *TestPostgresUser) GetPasswordHistories(username string) ([]string, error) {
	if repo.config == nil {
		return nil, fmt.Errorf("test db not initialized")
	}
	var histories []string
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.Select(&histories, "SELECT password_hash FROM user_password_history WHERE username=$1 ORDER BY created_at", username)
	})
	if err != nil {
		return nil, err
	}
	return histories, nil
}
func (repo *TestPostgresUser) GetLatestPasswordHistory(username string) (string, time.Time, error) {
	if repo.config == nil {
		return "", time.Time{}, fmt.Errorf("test db not initialized")
	}
	history := struct {
		PasswordHash string    `db:"password_hash"`
		CreatedAt    time.Time `db:"created_at"`
	}{}
	query := `
SELECT password_hash, created_at
FROM user_password_history
WHERE username=$1
ORDER BY created_at DESC LIMIT 1`
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.Get(&history, query, username)
	})
	if err != nil {
		return "", time.Time{}, err
	}
	return history.PasswordHash, history.CreatedAt, nil
}
func (repo *TestPostgresUser) CompareHashedPassword(hashedPassword, password []byte) error {
	return repo.compareHash(hashedPassword, password)
}

func Test_postgres_generateHash(t *testing.T) {
	t.Parallel()

	r := newUserRepository(orm.DatabaseConfig{})

	password := []byte("mysecretpassword")
	hash, err := r.generateHash(password)
	if err != nil {
		t.Errorf("generateHash() error = %v", err)
		return
	}
	if len(hash) == 0 {
		t.Errorf("generateHash() returned empty hash")
		return
	}

	// Verify the hash matches the original password
	err = bcrypt.CompareHashAndPassword(hash, password)
	if err != nil {
		t.Errorf("Generated hash does not match the original password: %v", err)
	}

	hash, err = r.generateHash(nil)
	if err != nil {
		t.Errorf("generateHash() error = %v", err)
		return
	}

	if hash != nil {
		t.Errorf("generateHash() with nil password should return nil hash, got %v", hash)
	}
}

func Test_postgres(t *testing.T) {
	useLocal := os.Getenv("USE_LOCAL_DB") == "true"

	testable := &TestPostgresUser{
		defaultDatabase: "postgres",
		testDatabase:    "pharos_test_user_repo",
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
	wrapper := &TestUserWrapper{TestPostgresUser: testable}

	// Common User repository test
	test.UserSuite(t, wrapper)
}
