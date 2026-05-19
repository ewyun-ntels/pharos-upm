package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/test"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/migration/schema"
	_ "ntels.com/pharos/core/pkg/migration/schema/master/base/sqlite"
)

var _ test.UserTestable = (*TestSqliteUser)(nil)

type TestSqliteUser struct {
	userRepository

	testDb   *sqlx.DB
	keepConn *sql.Conn
	isSetUp  bool // DB 설정 상태 추적
}

func (repo *TestSqliteUser) SetUpDb() error {
	// 이미 설정되어 있으면 테이블만 정리하고 반환
	if repo.isSetUp {
		return repo.ClearTable()
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

	repo.testDb = sqlx.NewDb(db, driverName)
	repo.userRepository = *newUserRepository(config.Database)
	repo.isSetUp = true
	return nil
}
func (repo *TestSqliteUser) TearDownDb() error {
	// 실제 종료는 최종 테스트가 끝날 때만 수행
	// 개별 테스트에서는 아무것도 하지 않음
	return nil
}

func (repo *TestSqliteUser) CleanupDb() error {
	// 실제 DB 정리 - 테스트 종료 시 호출
	if repo.testDb != nil {
		_ = repo.keepConn.Close()
		repo.keepConn = nil
		err := repo.testDb.Close()
		repo.testDb = nil
		repo.isSetUp = false
		orm.Unload()
		return err
	}
	return nil
}
func (repo *TestSqliteUser) SetUpUser(username string) error {
	if repo.testDb == nil {
		return fmt.Errorf("test db not initialized")
	}
	// 테스트용 사용자 삽입
	_, err := repo.testDb.Exec(`
	INSERT INTO users (username, password_hash, extra, created_at, retry, retry_at, blocked, prepare, password_expired_at)
	VALUES ($1, x'1234', '{}', CURRENT_TIMESTAMP, 0, NULL, 0, NULL, NULL)
	`, username)
	return err
}
func (repo *TestSqliteUser) ClearTable() error {
	if repo.testDb == nil {
		return fmt.Errorf("test db not initialized")
	}
	// 외래 키 제약으로 인해 user_password_history를 먼저 삭제
	_, err := repo.testDb.Exec("DELETE FROM user_password_history WHERE true")
	if err != nil {
		return err
	}
	_, err = repo.testDb.Exec("DELETE FROM users WHERE true")
	return err
}
func (repo *TestSqliteUser) GetUser(name string) (*repositories.UserEntity, error) {
	if repo.testDb == nil {
		return nil, fmt.Errorf("test db not initialized")
	}
	var u sqliteUser
	err := repo.testDb.Get(&u, "SELECT * FROM users WHERE username = ?", name)
	if err != nil {
		return nil, err
	}

	return repo.toUser(u)
}
func (repo *TestSqliteUser) GetUserCount() (int, error) {
	if repo.testDb == nil {
		return 0, fmt.Errorf("test db not initialized")
	}
	var count int
	err := repo.testDb.Get(&count, "SELECT COUNT(*) FROM users")
	if err != nil {
		return 0, err
	}
	return count, nil
}
func (repo *TestSqliteUser) GetPasswordHistories(username string) ([]string, error) {
	if repo.testDb == nil {
		return nil, fmt.Errorf("test db not initialized")
	}
	var histories []string
	err := repo.testDb.Select(&histories, "SELECT password_hash FROM user_password_history WHERE username=$1 ORDER BY created_at", username)
	if err != nil {
		return nil, err
	}
	return histories, nil
}
func (repo *TestSqliteUser) GetLatestPasswordHistory(username string) (string, time.Time, error) {
	if repo.testDb == nil {
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
	err := repo.testDb.Get(&history, query, username)
	if err != nil {
		return "", time.Time{}, err
	}
	return history.PasswordHash, history.CreatedAt, nil
}
func (repo *TestSqliteUser) CompareHashedPassword(hashedPassword, password []byte) error {
	return repo.compareHash(hashedPassword, password)
}

func Test_sqlite_generateHash(t *testing.T) {
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

func Test_sqlite(t *testing.T) {
	// DB 설정을 한 번만 수행하고 모든 테스트에서 재사용
	testRepo := &TestSqliteUser{}

	// 실제 정리는 모든 테스트가 끝난 후에만 수행
	t.Cleanup(func() {
		if err := testRepo.CleanupDb(); err != nil {
			t.Errorf("failed to cleanup test db: %v", err)
		}
	})

	// Common User repository test
	test.UserSuite(t, testRepo)
}
