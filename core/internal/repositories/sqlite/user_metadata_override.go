package sqlite

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
	sharedUser "ntels.com/pharos/shared/types/user"
)

var _ repositories.UserMetadataConfigRepository = (*userMetadataConfigRepository)(nil)

type userMetadataConfigRepository struct {
	repositories.BaseRepository
}

func newUserMetadataConfigRepository(config orm.DatabaseConfig) *userMetadataConfigRepository {
	return &userMetadataConfigRepository{
		BaseRepository: repositories.NewBaseRepository(config),
	}
}

type sqliteUserMetadataConfig struct {
	ID          string         `db:"id"`
	ConfigJSON  string         `db:"config_json"`
	CreatedAt   time.Time      `db:"created_at"`
	UpdatedBy   sql.NullString `db:"updated_by"`
	Description sql.NullString `db:"description"`
}

func (repo *userMetadataConfigRepository) toEntity(row sqliteUserMetadataConfig) (repositories.UserMetadataConfigEntity, error) {
	entity := repositories.UserMetadataConfigEntity{
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

	// Parse JSON to Config struct
	var config sharedUser.UserMetadataConfig
	if err := json.Unmarshal([]byte(row.ConfigJSON), &config); err != nil {
		return entity, err
	}
	entity.Config = &config

	return entity, nil
}

func (repo *userMetadataConfigRepository) GetActiveConfig(ctx context.Context) (*repositories.UserMetadataConfigEntity, error) {
	// Get the most recent configuration
	query := `SELECT id, config_json, created_at, updated_by, description
	          FROM user_metadata_config_history
	          ORDER BY created_at DESC
	          LIMIT 1`

	var row sqliteUserMetadataConfig
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

func (repo *userMetadataConfigRepository) GetHistory(ctx context.Context, limit int) ([]repositories.UserMetadataConfigEntity, error) {
	query := `SELECT id, config_json, created_at, updated_by, description
	          FROM user_metadata_config_history
	          ORDER BY created_at DESC
	          LIMIT ?`

	var rows []sqliteUserMetadataConfig
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.SelectContext(ctx, &rows, query, limit)
	})
	if err != nil {
		return nil, err
	}

	entities := make([]repositories.UserMetadataConfigEntity, 0, len(rows))
	for _, row := range rows {
		entity, err := repo.toEntity(row)
		if err != nil {
			return nil, err
		}
		entities = append(entities, entity)
	}

	return entities, nil
}

func (repo *userMetadataConfigRepository) SaveConfig(ctx context.Context, config repositories.UserMetadataConfigEntity, maxHistory int) error {
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
		createdAt := toDatabaseTime(time.Now())
		_, err = tx.ExecContext(ctx,
			`INSERT INTO user_metadata_config_history (id, config_json, created_at, updated_by, description)
			 VALUES (?, ?, ?, ?, ?)`,
			newID, config.ConfigJSON, createdAt, updatedBy, description)
		if err != nil {
			return err
		}

		// 2. Delete old versions beyond maxHistory
		_, err = tx.ExecContext(ctx,
			`DELETE FROM user_metadata_config_history
			 WHERE id NOT IN (
			     SELECT id FROM user_metadata_config_history
			     ORDER BY created_at DESC
			     LIMIT ?
			 )`,
			maxHistory)
		if err != nil {
			return err
		}

		return tx.Commit()
	})
}

func (repo *userMetadataConfigRepository) GetByID(ctx context.Context, id string) (*repositories.UserMetadataConfigEntity, error) {
	query := `SELECT id, config_json, created_at, updated_by, description
	          FROM user_metadata_config_history
	          WHERE id = ?`

	var row sqliteUserMetadataConfig
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.GetContext(ctx, &row, query, id)
	})
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}

	entity, err := repo.toEntity(row)
	if err != nil {
		return nil, err
	}
	return &entity, nil
}
