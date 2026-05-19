package model

import (
	"errors"
	"log/slog"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

type Rule struct {
	ID               string       `db:"id"`
	NotificationType string       `db:"notification_type"`
	Name             string       `db:"name"`
	Rule             string       `db:"rule"`
	Timestamp        orm.Datetime `db:"timestamp"` // formating 문제로 인하여 string형식으로 변경
	UpdatedAt        orm.Datetime `db:"updated_at"`
}

type Model struct {
	Config *common.Config
}

func (m *Model) InsertRule(rule *Rule) (err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		_, err = db.NamedExec(`
	INSERT INTO notification_rule (id, name, notification_type, rule, timestamp, updated_at) VALUES
	(:id, :name, :notification_type, :rule, :timestamp, :updated_at);
	`, rule)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return
}

func (m *Model) UpdateRule(rule *Rule) (err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		_, err = db.NamedExec(`
	UPDATE notification_rule SET
	id =:id,
	name =:name,
	notification_type =:notification_type,
	rule =:rule,
	updated_at =:updated_at
	WHERE id = :id
	`, rule)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return
}

func (m *Model) DeleteRule(id string) (err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		_, err = db.Exec(`
	DELETE FROM notification_rule WHERE id = $1
	`, id)
		if err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		return err
	}

	return nil
}

func (m *Model) GetRuleByName(name string) (rule *Rule, err error) {
	var res []Rule
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		err = db.Select(&res, `SELECT * FROM notification_rule WHERE name = $1`, name)
		if err != nil {
			return err
		}

		return nil
	})

	if len(res) == 0 {
		return nil, err
	}

	if len(res) > 1 {
		return nil, errors.New("too many results")
	}

	return &res[0], err
}

func (m *Model) GetRule(id string) (rule *Rule, err error) {
	var res []Rule
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		err = db.Select(&res, `SELECT * FROM notification_rule WHERE id = $1`, id)
		if err != nil {
			slog.Error("rule get failed", "error", err)
			return err
		}

		return nil
	})

	if len(res) == 0 {
		return nil, err
	}

	if len(res) > 1 {
		return nil, errors.New("too many results")
	}

	return &res[0], err
}

func (m *Model) GetAllRuleDb() (ruleDbs []Rule, err error) {
	err = orm.Handler(orm.DriverDefault, &m.Config.Database, func(db *sqlx.DB) error {
		err = db.Select(&ruleDbs, `SELECT * FROM notification_rule`)
		if err != nil {
			slog.Error("rule list load failed", "error", err)
			return err
		}

		return nil
	})

	return
}
