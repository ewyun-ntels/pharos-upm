package model

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

func setupInMemoryDB(t *testing.T) (cfg common.Config, cleanup func()) {
	t.Helper()
	cfg = common.Config{}
	cfg.Database = orm.DatabaseConfig{Driver: orm.DriverSqlite, SQLite: orm.SQLiteConfig{Path: ":memory:"}}
	// initialize schema
	err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		_, err := db.Exec(`CREATE TABLE IF NOT EXISTS notification_rule (
			id TEXT PRIMARY KEY,
			name TEXT,
			notification_type TEXT,
			rule TEXT,
			timestamp TEXT,
			updated_at TEXT
		);`)
		return err
	})
	if err != nil {
		t.Fatalf("schema init err: %v", err)
	}
	return cfg, func() { _ = orm.DatabasePool.Remove(cfg.Database) }
}

func TestModel_CRUD(t *testing.T) {
	cfg, cleanup := setupInMemoryDB(t)
	defer cleanup()

	m := &Model{Config: &cfg}
	now := time.Now().Truncate(time.Second)
	r := &Rule{
		ID:               "r1",
		NotificationType: "snmp",
		Name:             "rule-name",
		Rule:             `{"k":"v"}`,
		Timestamp:        orm.Datetime{Time: now},
		UpdatedAt:        orm.Datetime{Time: now},
	}

	// Insert
	if err := m.InsertRule(r); err != nil {
		t.Fatalf("insert: %v", err)
	}

	// Get by id
	got, err := m.GetRule("r1")
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if got.Name != r.Name || got.NotificationType != r.NotificationType || got.Rule != r.Rule {
		t.Fatalf("mismatch after get: %+v", got)
	}

	// Get by name
	gotByName, err := m.GetRuleByName("rule-name")
	if err != nil {
		t.Fatalf("get by name: %v", err)
	}
	if gotByName.ID != "r1" {
		t.Fatalf("unexpected id: %s", gotByName.ID)
	}

	// Update
	r.Rule = `{"k":"v2"}`
	r.UpdatedAt = orm.Datetime{Time: now.Add(time.Minute)}
	if err := m.UpdateRule(r); err != nil {
		t.Fatalf("update: %v", err)
	}
	got, err = m.GetRule("r1")
	if err != nil {
		t.Fatalf("get2: %v", err)
	}
	if got.Rule != `{"k":"v2"}` {
		t.Fatalf("update not applied: %s", got.Rule)
	}

	// List all
	all, err := m.GetAllRuleDb()
	if err != nil {
		t.Fatalf("list: %v", err)
	}
	if len(all) != 1 {
		t.Fatalf("expected 1, got %d", len(all))
	}

	// Delete
	if err := m.DeleteRule("r1"); err != nil {
		t.Fatalf("delete: %v", err)
	}
	all, err = m.GetAllRuleDb()
	if err != nil {
		t.Fatalf("list2: %v", err)
	}
	if len(all) != 0 {
		t.Fatalf("expected 0 after delete, got %d", len(all))
	}
}

func TestModel_GetRuleByName_NotFound(t *testing.T) {
	cfg, cleanup := setupInMemoryDB(t)
	defer cleanup()
	m := &Model{Config: &cfg}
	if r, err := m.GetRuleByName("nope"); r != nil || err != nil {
		t.Fatalf("expected nil,nil got r=%v err=%v", r, err)
	}
}
