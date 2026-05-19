package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories"
	sharedRole "ntels.com/pharos/shared/types/role"
)

var _ repositories.RoleMetadataConfigRepository = (*roleMetadataConfigRepository)(nil)

type roleMetadataConfigRepository struct {
	repositories.BaseRepository
}

func newRoleMetadataConfigRepository(config orm.DatabaseConfig) *roleMetadataConfigRepository {
	return &roleMetadataConfigRepository{
		BaseRepository: repositories.NewBaseRepository(config),
	}
}

type postgresRoleMetadataConfig struct {
	ID          string         `db:"id"`
	ConfigJSON  string         `db:"config_json"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedBy   sql.NullString `db:"updated_by"`
	Description sql.NullString `db:"description"`
}

func (repo *roleMetadataConfigRepository) toEntity(row postgresRoleMetadataConfig) (repositories.RoleMetadataConfigEntity, error) {
	entity := repositories.RoleMetadataConfigEntity{
		ID:         row.ID,
		ConfigJSON: row.ConfigJSON,
		CreatedAt:  row.CreatedAt,
	}

	if row.UpdatedBy.Valid {
		entity.UpdatedBy = &row.UpdatedBy.String
	}

	if row.Description.Valid {
		entity.Description = &row.Description.String
	}

	// Parse JSON to Config array
	var config []sharedRole.RoleMetadata
	if err := json.Unmarshal([]byte(row.ConfigJSON), &config); err != nil {
		return entity, err
	}
	entity.Config = config

	return entity, nil
}

func (repo *roleMetadataConfigRepository) GetActiveConfig(ctx context.Context) (*repositories.RoleMetadataConfigEntity, error) {
	// Get the most recent configuration
	query := `SELECT id, config_json, created_at, updated_by, description
	          FROM role_metadata_config_history
	          ORDER BY created_at DESC
	          LIMIT 1`

	var row postgresRoleMetadataConfig
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.GetContext(ctx, &row, query)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil // No configuration found
		}
		return nil, err
	}

	entity, err := repo.toEntity(row)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}

func (repo *roleMetadataConfigRepository) GetHistory(ctx context.Context, limit int) ([]repositories.RoleMetadataConfigEntity, error) {
	query := `SELECT id, config_json, created_at, updated_by, description
	          FROM role_metadata_config_history
	          ORDER BY created_at DESC
	          LIMIT $1`

	var rows []postgresRoleMetadataConfig
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.SelectContext(ctx, &rows, query, limit)
	})
	if err != nil {
		return nil, err
	}

	entities := make([]repositories.RoleMetadataConfigEntity, 0, len(rows))
	for _, row := range rows {
		entity, err := repo.toEntity(row)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}

	return entities, nil
}

func (repo *roleMetadataConfigRepository) SaveConfig(ctx context.Context, config repositories.RoleMetadataConfigEntity, maxHistory int) error {
	return repo.Handler(func(db *sqlx.DB) error {
		tx, err := db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

		// 1. Insert new configuration with UUID
		updatedBy := sql.NullString{}
		if config.UpdatedBy != nil {
			updatedBy = sql.NullString{String: *config.UpdatedBy, Valid: true}
		}

		description := sql.NullString{}
		if config.Description != nil {
			description = sql.NullString{String: *config.Description, Valid: true}
		}

		newID := uuid.New().String()
		_, err = tx.ExecContext(ctx,
			`INSERT INTO role_metadata_config_history (id, config_json, created_at, updated_by, description)
			 VALUES ($1, $2, $3, $4, $5)`,
			newID, config.ConfigJSON, time.Now(), updatedBy, description)
		if err != nil {
			return err
		}

		// 2. Delete old versions beyond maxHistory
		_, err = tx.ExecContext(ctx,
			`DELETE FROM role_metadata_config_history
			 WHERE id NOT IN (
			     SELECT id FROM role_metadata_config_history
			     ORDER BY created_at DESC
			     LIMIT $1
			 )`,
			maxHistory)
		if err != nil {
			return err
		}

		return tx.Commit()
	})
}

func (repo *roleMetadataConfigRepository) Rollback(ctx context.Context, id string, updatedBy *string, maxHistory int) error {
	return repo.Handler(func(db *sqlx.DB) error {
		tx, err := db.BeginTxx(ctx, nil)
		if err != nil {
			return err
		}
		defer func() { _ = tx.Rollback() }()

		// 1. Get the version to rollback to
		var configJSON string
		err = tx.GetContext(ctx, &configJSON,
			`SELECT config_json FROM role_metadata_config_history WHERE id = $1`, id)
		if err != nil {
			return err
		}

		// 2. Insert rolled-back version as new record with new UUID
		updatedBySQL := sql.NullString{}
		if updatedBy != nil {
			updatedBySQL = sql.NullString{String: *updatedBy, Valid: true}
		}

		description := sql.NullString{String: "Rollback to version " + id, Valid: true}

		newID := uuid.New().String()
		_, err = tx.ExecContext(ctx,
			`INSERT INTO role_metadata_config_history (id, config_json, created_at, updated_by, description)
			 VALUES ($1, $2, $3, $4, $5)`,
			newID, configJSON, time.Now(), updatedBySQL, description)
		if err != nil {
			return err
		}

		// 3. Cleanup old versions
		_, err = tx.ExecContext(ctx,
			`DELETE FROM role_metadata_config_history
			 WHERE id NOT IN (
			     SELECT id FROM role_metadata_config_history
			     ORDER BY created_at DESC
			     LIMIT $1
			 )`,
			maxHistory)
		if err != nil {
			return err
		}

		return tx.Commit()
	})
}

func (repo *roleMetadataConfigRepository) DeleteAll(ctx context.Context) error {
	query := `DELETE FROM role_metadata_config_history`
	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.ExecContext(ctx, query)
		return err
	})
}
