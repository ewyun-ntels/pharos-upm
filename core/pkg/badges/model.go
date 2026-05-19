package badges

import (
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

// Model provides DB access for badges.
type Model struct {
	Config *common.Config
}

// BadgeRow represents a stored badge definition.
type BadgeRow struct {
	Name       string       `db:"name"`
	Datasource string       `db:"datasource"`
	QueryJSON  string       `db:"query_json"`
	ColumnName string       `db:"column_name"`
	UpdatedAt  orm.Datetime `db:"updated_at"`
}

// Upsert stores or updates a badge definition by name.
func (m *Model) Upsert(name, datasource string, query any, column string) error {
	b, err := json.Marshal(query)
	if err != nil {
		return err
	}
	now := orm.Datetime{Time: time.Now().UTC()}
	row := BadgeRow{Name: name, Datasource: datasource, QueryJSON: string(b), ColumnName: column, UpdatedAt: now}
	return orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		_, err := db.NamedExec(`
  INSERT INTO badges (name, datasource, query_json, column_name, updated_at)
		VALUES (:name, :datasource, :query_json, :column_name, :updated_at)
		ON CONFLICT(name) DO UPDATE SET
			datasource = EXCLUDED.datasource,
			query_json = EXCLUDED.query_json,
			column_name = EXCLUDED.column_name,
			updated_at = EXCLUDED.updated_at;
		`, row)
		return err
	})
}

// Get retrieves a badge by name and decodes the stored query JSON into dst.
func (m *Model) Get(name string, dst any) (*BadgeRow, error) {
	var row BadgeRow
	err := orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		return db.Get(&row, `SELECT name, datasource, query_json, column_name, updated_at FROM badges WHERE name = $1`, name)
	})
	if err != nil {
		// on not found, return nil,nil
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	if row.Name == "" {
		return nil, nil
	}
	if dst != nil {
		if err := json.Unmarshal([]byte(row.QueryJSON), dst); err != nil {
			return nil, err
		}
	}
	return &row, nil
}

// List returns all badge names.
func (m *Model) List() ([]string, error) {
	var rows []string
	err := orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		return db.Select(&rows, `SELECT name FROM badges ORDER BY name`)
	})
	return rows, err
}

// Delete removes a badge by name.
func (m *Model) Delete(name string) error {
	if name == "" {
		return errors.New("name is empty")
	}
	return orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`DELETE FROM badges WHERE name = $1`, name)
		return err
	})
}
