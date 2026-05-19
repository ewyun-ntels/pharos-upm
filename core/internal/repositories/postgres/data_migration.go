package postgres

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/migration"
	"ntels.com/pharos/core/internal/repositories"
)

type dataMigrationRepository struct {
	repositories.BaseRepository
}

func newDataMigrationRepository(config orm.DatabaseConfig) repositories.DataMigrationRepository {
	return &dataMigrationRepository{
		BaseRepository: repositories.NewBaseRepository(config),
	}
}

func (r *dataMigrationRepository) CreateMigrationTable() error {
	return r.Handler(func(db *sqlx.DB) error {
		// Create table structure
		query := fmt.Sprintf(`
			CREATE TABLE IF NOT EXISTS %s (
				version_id BIGINT PRIMARY KEY,
				is_applied BOOLEAN NOT NULL DEFAULT true,
				tstamp TIMESTAMP DEFAULT CURRENT_TIMESTAMP
			)
		`, migration.DataMigrationTableName)

		_, err := db.Exec(query)
		return err
	})
}

func (r *dataMigrationRepository) GetAppliedMigrations() (map[int64]bool, error) {
	applied := make(map[int64]bool)

	err := r.Handler(func(db *sqlx.DB) error {
		query := fmt.Sprintf("SELECT version_id FROM %s WHERE is_applied = true", migration.DataMigrationTableName)
		rows, err := db.Query(query)
		if err != nil {
			return err
		}
		defer func() { _ = rows.Close() }()

		for rows.Next() {
			var version int64
			if err := rows.Scan(&version); err != nil {
				return err
			}
			applied[version] = true
		}
		return rows.Err()
	})

	return applied, err
}

func (r *dataMigrationRepository) RecordMigration(tx *sql.Tx, version int64) error {
	// PostgreSQL uses $1, $2, $3 placeholders
	query := fmt.Sprintf(
		"INSERT INTO %s (version_id, is_applied, tstamp) VALUES ($1, $2, $3)",
		migration.DataMigrationTableName,
	)
	_, err := tx.Exec(query, version, true, time.Now())
	return err
}

func (r *dataMigrationRepository) IsFreshInstall() (bool, error) {
	var isFresh bool
	err := r.Handler(func(db *sqlx.DB) error {
		// Fresh install 판단: core_db_version (스키마 마이그레이션) 테이블 존재 여부만 확인
		// core_db_version 없음 = 깡통 설치 = Fresh install → 데이터 마이그레이션 스킵
		// core_db_version 있음 = 기존 설치 → 데이터 마이그레이션 실행
		query := fmt.Sprintf(
			`SELECT EXISTS (
				SELECT FROM information_schema.tables 
				WHERE table_name = '%s'
			)`, migration.CoreDBVersionTableName)

		var exists bool
		err := db.QueryRow(query).Scan(&exists)
		if err != nil {
			return err
		}

		// core_db_version이 없으면 fresh install
		isFresh = !exists
		return nil
	})
	return isFresh, err
}

func (r *dataMigrationRepository) MarkAllAsApplied(versions []int64) error {
	return r.Handler(func(db *sqlx.DB) error {
		tx, err := db.Begin()
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

		// ON CONFLICT DO NOTHING: 이미 존재하는 version_id는 무시 (재배포/재시작 시 중복 INSERT 방지)
		query := fmt.Sprintf(
			"INSERT INTO %s (version_id, is_applied, tstamp) VALUES ($1, $2, $3) ON CONFLICT (version_id) DO NOTHING",
			migration.DataMigrationTableName,
		)

		for _, version := range versions {
			_, err := tx.Exec(query, version, true, time.Now())
			if err != nil {
				return fmt.Errorf("failed to mark version %d as applied: %w", version, err)
			}
		}

		return tx.Commit()
	})
}
