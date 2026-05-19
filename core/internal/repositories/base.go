package repositories

import (
	"encoding/json"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
)

// BaseRepository provides common database operations
type BaseRepository struct {
	config orm.DatabaseConfig
}

// NewBaseRepository creates a new base repository
func NewBaseRepository(config orm.DatabaseConfig) BaseRepository {
	return BaseRepository{
		config: config,
	}
}

// Handler executes a function with a database connection
func (repo *BaseRepository) Handler(f func(db *sqlx.DB) error) error {
	return orm.Handler("", &repo.config, f)
}

// MarshalAttributes converts attributes map to JSON bytes
func (repo *BaseRepository) MarshalAttributes(attributes map[string]any) ([]byte, error) {
	if attributes == nil {
		return []byte("{}"), nil
	}
	return json.Marshal(attributes)
}

// UnmarshalAttributes converts JSON bytes to attributes map
func (repo *BaseRepository) UnmarshalAttributes(data []byte) (map[string]any, error) {
	if data == nil {
		return map[string]any{}, nil
	}
	var attributes map[string]any
	err := json.Unmarshal(data, &attributes)
	if err != nil {
		return nil, err
	}
	return attributes, nil
}

// MarshalPrepare converts prepare slice to JSON bytes
func (repo *BaseRepository) MarshalPrepare(prepare []string) ([]byte, error) {
	if prepare == nil {
		return []byte("[]"), nil
	}
	return json.Marshal(prepare)
}

// UnmarshalPrepare converts JSON bytes to prepare slice
func (repo *BaseRepository) UnmarshalPrepare(data []byte) ([]string, error) {
	if data != nil {
		var prepare []string
		err := json.Unmarshal(data, &prepare)
		if err != nil {
			return nil, err
		}
		return prepare, nil
	}
	return nil, nil
}
