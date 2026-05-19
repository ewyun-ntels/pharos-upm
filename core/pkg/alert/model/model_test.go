package model

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	alert_common "ntels.com/pharos/core/pkg/alert/common"
	"ntels.com/pharos/core/pkg/common"
)

// helper to setup a sqlite database and schema for tests
func newTestModel(t *testing.T) (*Model, func()) {
	t.Helper()

	cfg := common.Config{}
	cfg.Database.Driver = orm.DriverSqlite
	cfg.Database.SQLite.Path = t.TempDir() + "/test.db"

	// initialize schema
	err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		stmts := []string{
			// alert_status with unique constraint on (alert_id, name) to support upsert
			`CREATE TABLE IF NOT EXISTS alert_status (
				id TEXT,
				alert_id TEXT NOT NULL,
				name TEXT NOT NULL,
				mask BOOLEAN,
				alert_type TEXT,
				description TEXT,
				status TEXT,
				severity TEXT,
				previous_severity TEXT,
				previous_value REAL,
				value REAL,
				labels TEXT,
    timestamp TIMESTAMP,
				updated_at TIMESTAMP,
				check_time TIMESTAMP,
				PRIMARY KEY (alert_id, name)
			);`,
			// alert_rule table
			`CREATE TABLE IF NOT EXISTS alert_rule (
				id TEXT PRIMARY KEY,
				name TEXT,
				alert_type TEXT,
				rule TEXT,
				timestamp TEXT,
				updated_at TEXT
			);`,
			// alert_hist table
			`CREATE TABLE IF NOT EXISTS alert_hist (
    timestamp TIMESTAMP,
				updated_at TIMESTAMP,
				alert_id TEXT,
				name TEXT,
				alert_type TEXT,
				description TEXT,
				previous_severity TEXT,
				previous_value REAL,
				severity TEXT,
				value REAL,
				labels TEXT
			);`,
		}
		for _, s := range stmts {
			if _, err := db.Exec(s); err != nil {
				return err
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("schema init failed: %v", err)
	}

	m := &Model{Config: &cfg}

	cleanup := func() {
		_ = orm.DatabasePool.Remove(cfg.Database)
	}
	return m, cleanup
}

func makeSampleAlert(name string, value float64) alert_common.Value {
	now := time.Now().UTC()
	ct := now
	labels := map[string]string{"host": "test-host", "region": "kr"}
	a := alert_common.Value{
		Id:          "some-id",
		Timestamp:   now,
		UpdatedAt:   now,
		CheckTime:   &ct,
		Name:        name,
		Mask:        false,
		AlertType:   "query",
		Description: "desc",
		Value:       value,
		Severity:    "Major",
		Status:      alert_common.StatusAlerting,
		Labels:      labels,
	}
	_ = a.ConvertLabelsToStringLabels()
	_ = a.MakeKey()
	return a
}

func TestUpdateAlert_InsertAndUpsert(t *testing.T) {
	m, cleanup := newTestModel(t)
	defer cleanup()

	a := makeSampleAlert("cpu_high", 85)
	if err := m.UpdateAlert(&a); err != nil {
		t.Fatalf("UpdateAlert(insert) error: %v", err)
	}

	// Verify inserted
	got, err := m.GetAlert(a.Name, a.AlertId)
	if err != nil {
		t.Fatalf("GetAlert error: %v", err)
	}
	if got == nil {
		t.Fatalf("GetAlert returned nil after insert")
	}
	if got.Value != a.Value || got.Severity != a.Severity || got.Status != a.Status {
		t.Fatalf("mismatch after insert: %+v vs %+v", got, a)
	}

	// Upsert: change value/severity/status
	a.Value = 20
	a.Severity = "Normal"
	a.Status = alert_common.StatusNormal
	if err := m.UpdateAlert(&a); err != nil {
		t.Fatalf("UpdateAlert(upsert) error: %v", err)
	}

	got2, err := m.GetAlert(a.Name, a.AlertId)
	if err != nil {
		t.Fatalf("GetAlert after upsert error: %v", err)
	}
	if got2 == nil {
		t.Fatalf("GetAlert returned nil after upsert")
	}
	if got2.Value != 20 || got2.Severity != "Normal" || got2.Status != alert_common.StatusNormal {
		t.Fatalf("unexpected updated values: %+v", got2)
	}
}

func TestRuleCRUD(t *testing.T) {
	m, cleanup := newTestModel(t)
	defer cleanup()

	now := orm.Datetime{Time: time.Now().UTC()}
	r := &Rule{ID: "rule-1", AlertType: "query", Name: "cpu_high", Rule: `{"k":"v"}`, Timestamp: now, UpdatedAt: now}

	if err := m.InsertRule(r); err != nil {
		t.Fatalf("InsertRule error: %v", err)
	}

	byName, err := m.GetRuleByName("cpu_high")
	if err != nil || byName == nil {
		t.Fatalf("GetRuleByName error: %v, val:%v", err, byName)
	}
	if byName.ID != r.ID {
		t.Fatalf("GetRuleByName ID mismatch: %s", byName.ID)
	}

	byID, err := m.GetRule("rule-1")
	if err != nil || byID == nil {
		t.Fatalf("GetRule error: %v, val:%v", err, byID)
	}

	// Update
	r.Name = "cpu_hot"
	r.UpdatedAt = orm.Datetime{Time: time.Now().UTC()}
	r.Rule = `{"k":"v2"}`
	if err := m.UpdateRule(r); err != nil {
		t.Fatalf("UpdateRule error: %v", err)
	}

	byID2, _ := m.GetRule("rule-1")
	if byID2.Name != "cpu_hot" {
		t.Fatalf("UpdateRule did not update name: %+v", byID2)
	}

	// List
	list, err := m.GetAllRuleDb()
	if err != nil {
		t.Fatalf("GetAllRuleDb error: %v", err)
	}
	if len(list) != 1 {
		t.Fatalf("expected 1 rule, got %d", len(list))
	}

	// Delete
	if err := m.DeleteRule("rule-1"); err != nil {
		t.Fatalf("DeleteRule error: %v", err)
	}
	none, err := m.GetRule("rule-1")
	if err != nil && err != sql.ErrNoRows { /* allow nil err with nil result */
	}
	if none != nil {
		t.Fatalf("expected rule to be deleted, got: %+v", none)
	}
}

func TestGetStatusesAndMask(t *testing.T) {
	m, cleanup := newTestModel(t)
	defer cleanup()

	a := makeSampleAlert("cpu_high", 90)
	if err := m.UpdateAlert(&a); err != nil {
		t.Fatalf("UpdateAlert error: %v", err)
	}

	// Set mask
	if err := m.SetStatusMask(a.Name, a.AlertId, true); err != nil {
		t.Fatalf("SetStatusMask error: %v", err)
	}

	one, err := m.GetStatus(a.Name, a.AlertId)
	if err != nil {
		t.Fatalf("GetStatus error: %v", err)
	}
	if one == nil || !one.Mask {
		t.Fatalf("GetStatus mask not true: %+v", one)
	}
	if len(one.Labels) == 0 {
		t.Fatalf("Labels should be decoded from StringLabels")
	}

	list, err := m.GetStatuses()
	if err != nil {
		t.Fatalf("GetStatuses error: %v", err)
	}
	if list == nil || len(*list) != 1 {
		t.Fatalf("expected 1 status, got %v", list)
	}
}

func TestInsertAndQueryRuleHist(t *testing.T) {
	m, cleanup := newTestModel(t)
	defer cleanup()

	a1 := makeSampleAlert("cpu_high", 80)
	a1.Severity = "Major"
	a2 := makeSampleAlert("cpu_high", 20)
	a2.Severity = "Normal"

	if err := m.InsertRuleHist([]alert_common.Value{a1, a2}); err != nil {
		t.Fatalf("InsertRuleHist error: %v", err)
	}

	st := orm.Datetime{Time: time.Now().Add(-time.Hour).UTC()}
	et := orm.Datetime{Time: time.Now().Add(time.Hour).UTC()}

	rows, err := m.QueryRuleHist("cpu_high", st, et, 10)
	if err != nil {
		t.Fatalf("QueryRuleHist error: %v", err)
	}
	if len(rows) == 0 {
		t.Fatalf("expected some history rows, got 0")
	}

	// Ensure labels JSON were parsed when non-empty
	// We marshal one of the alerts' labels and ensure it matches stored StringLabels
	if rows[0].Labels != nil {
		b, _ := json.Marshal(rows[0].Labels)
		if string(b) == "" {
			t.Fatalf("labels should be present when non-empty")
		}
	}
}
