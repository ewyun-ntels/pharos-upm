package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories/test"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/migration/schema"
	_ "ntels.com/pharos/core/pkg/migration/schema/master/base/sqlite"
)

var _ test.RoleMetadataConfigTestable = (*testRoleMetadataConfig)(nil)

type testRoleMetadataConfig struct {
	roleMetadataConfigRepository
	testDb   *sqlx.DB
	keepConn *sql.Conn
	isSetUp  bool // DB 설정 상태 추적
}

func (repo *testRoleMetadataConfig) SetUpDb() error {
	// 이미 설정되어 있으면 테이블만 정리하고 반환
	if repo.isSetUp {
		return repo.ClearTable()
	}

	driverName := "sqlite"
	dsn := "file:test_role_override_" + uuid.New().String() + "?mode=memory&cache=shared"
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
	repo.roleMetadataConfigRepository = *newRoleMetadataConfigRepository(config.Database)
	repo.isSetUp = true
	return nil
}

func (repo *testRoleMetadataConfig) ClearTable() error {
	_, err := repo.testDb.Exec("DELETE FROM role_metadata_config_history")
	return err
}

func (repo *testRoleMetadataConfig) TearDownDb() error {
	// 실제 종료는 최종 테스트가 끝날 때만 수행
	// 개별 테스트에서는 아무것도 하지 않음
	return nil
}

func (repo *testRoleMetadataConfig) CleanupDb() error {
	// 실제 DB 정리 - 테스트 종료 시 호출
	if repo.testDb != nil {
		_ = repo.testDb.Close()
	}
	if repo.keepConn != nil {
		_ = repo.keepConn.Close()
	}
	return nil
}

func Test_sqlite_RoleMetadataConfig(t *testing.T) {
	// DB 설정을 한 번만 수행하고 모든 테스트에서 재사용
	testRepo := &testRoleMetadataConfig{}

	// 실제 정리는 모든 테스트가 끝난 후에만 수행
	t.Cleanup(func() {
		if err := testRepo.CleanupDb(); err != nil {
			t.Errorf("failed to cleanup test db: %v", err)
		}
	})

	// Common test suite
	test.RoleMetadataConfigSuite(t, testRepo)
}
