package version

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

type dbVersion struct {
	Module    string `db:"module"`
	Version   string `db:"version"`
	Commit    string `db:"commit"`
	BuildTime string `db:"build_time"`
}

// Model provides DB access for the version module, mirroring the pattern used in alert/model.
type Model struct {
	Config *common.Config
}

type CoreVersionRow struct {
	Module    string         `db:"module"`
	Version   string         `db:"version"`
	Commit    sql.NullString `db:"commit"`
	BuildTime sql.NullString `db:"build_time"`
	UpdatedAt sql.NullTime   `db:"updated_at"`
}

func (m *Model) GetCoreVersionRow() (CoreVersionRow, bool, error) {
	var row CoreVersionRow
	var found bool
	err := orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		if err := db.Get(&row, `SELECT module, version, "commit", build_time, updated_at FROM core_app_version WHERE module = 'core';`); err != nil {
			// On any error (including no rows), treat as not found and no error returned to caller (it can fallback)
			return nil
		}
		found = true
		return nil
	})
	return row, found, err
}

func (m *Model) UpsertVersion() error {
	v := dbVersion{
		Module:    "core",
		Version:   Version,
		Commit:    Commit,
		BuildTime: BuildTime,
	}
	return orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		_, err := db.NamedExec(`
INSERT INTO core_app_version(module, version, "commit", build_time, updated_at)
VALUES (:module, :version, :commit, :build_time, CURRENT_TIMESTAMP)
ON CONFLICT(module) DO UPDATE SET
		version=excluded.version,
  "commit"=excluded."commit",
		build_time=excluded.build_time,
		updated_at=CURRENT_TIMESTAMP;
`, v)
		return err
	})
}
