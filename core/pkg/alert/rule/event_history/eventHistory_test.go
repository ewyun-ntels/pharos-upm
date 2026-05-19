package event_history

import (
	"context"
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	alert_common "ntels.com/pharos/core/pkg/alert/common"
	"ntels.com/pharos/core/pkg/alert/model"
	"ntels.com/pharos/core/pkg/alert/resources"
	"ntels.com/pharos/core/pkg/common"
)

// setup a sqlite database schema sufficient for model.UpdateAlert/GetAlert used by event_history
func newTestConfig(t *testing.T) (common.Config, func()) {
	t.Helper()

	cfg := common.Config{}
	cfg.Database.Driver = orm.DriverSqlite
	cfg.Database.SQLite.Path = t.TempDir() + "/test.db"

	err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
		stmts := []string{
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

	cleanup := func() {
		_ = orm.DatabasePool.Remove(cfg.Database)
	}
	return cfg, cleanup
}

func newRuleForTest(t *testing.T) (*Rule, func(), *model.Model) {
	cfg, cleanup := newTestConfig(t)
	r := &Rule{Name: "fault"}
	r.config = cfg
	m := &model.Model{Config: &cfg}
	return r, cleanup, m
}

func TestEventHandler_EmptyAlertId(t *testing.T) {
	r, cleanup, _ := newRuleForTest(t)
	defer cleanup()
	ev := &resources.Event{ // missing AlertId
		Severity:    "Major",
		Description: "desc",
		Value:       1,
		Labels:      map[string]string{"k": "v"},
	}
	if err := r.EventHandler(ev); err == nil {
		t.Fatalf("expected error for empty AlertId, got nil")
	}
}

func TestEventHandler_CreateHistoryEventAndPersist(t *testing.T) {
	r, cleanup, m := newRuleForTest(t)
	defer cleanup()
	ev := &resources.Event{
		AlertId:     "H-1",
		Severity:    "Minor",
		Description: "hist",
		Value:       12.3,
		Labels:      map[string]string{"host": "h1"},
	}
	if err := r.EventHandler(ev); err != nil {
		t.Fatalf("EventHandler error: %v", err)
	}
	got, err := m.GetAlert(r.Name, ev.AlertId)
	if err != nil {
		t.Fatalf("GetAlert error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected stored alert row")
	}
	if got.Severity != "Minor" || got.Status != alert_common.StatusEvent {
		t.Fatalf("unexpected stored fields: %+v", got)
	}
	if got.AlertType != string(alert_common.TypeEventHistory) {
		t.Fatalf("alert type should be event-history, got %s", got.AlertType)
	}
	if got.Name != r.Name {
		t.Fatalf("name mismatch: %s vs %s", got.Name, r.Name)
	}
}

func TestExecute_DeletesOldByRetention(t *testing.T) {
	r, cleanup, m := newRuleForTest(t)
	defer cleanup()
	// set retention to 5 seconds to avoid timing issues
	r.RetentionPeriod = 5

	// insert two alerts: one old, one recent
	// old alert: 10 seconds ago (should be deleted)
	oldTime := time.Now().Add(-10 * time.Second).UTC()
	old := alert_common.Value{
		Id:          "id-old",
		Timestamp:   oldTime,
		UpdatedAt:   oldTime,
		CheckTime:   &oldTime,
		Name:        r.Name,
		AlertType:   string(alert_common.TypeEventHistory),
		Description: "old",
		AlertId:     "H-2",
		Value:       1,
		Severity:    "Major",
		Status:      alert_common.StatusEvent,
		Labels:      map[string]string{"host": "h2"},
	}
	_ = old.ConvertLabelsToStringLabels()
	if err := m.UpdateAlert(&old); err != nil {
		t.Fatalf("seed old UpdateAlert error: %v", err)
	}

	// recent alert: 2 seconds ago (should remain)
	recentTime := time.Now().Add(-2 * time.Second).UTC()
	recent := alert_common.Value{
		Id:          "id-new",
		Timestamp:   recentTime,
		UpdatedAt:   recentTime,
		CheckTime:   &recentTime,
		Name:        r.Name,
		AlertType:   string(alert_common.TypeEventHistory),
		Description: "new",
		AlertId:     "H-3",
		Value:       2,
		Severity:    "Minor",
		Status:      alert_common.StatusEvent,
		Labels:      map[string]string{"host": "h3"},
	}
	_ = recent.ConvertLabelsToStringLabels()
	if err := m.UpdateAlert(&recent); err != nil {
		t.Fatalf("seed recent UpdateAlert error: %v", err)
	}

	// run Execute to apply retention
	if err := r.Execute(context.TODO()); err != nil {
		t.Fatalf("Execute error: %v", err)
	}

	// old should be deleted, recent should remain
	gotOld, err := m.GetAlert(r.Name, old.AlertId)
	if err != nil {
		t.Fatalf("GetAlert old error: %v", err)
	}
	if gotOld != nil {
		t.Fatalf("expected old alert deleted by retention")
	}
	gotNew, err := m.GetAlert(r.Name, recent.AlertId)
	if err != nil {
		t.Fatalf("GetAlert recent error: %v", err)
	}
	if gotNew == nil {
		t.Fatalf("expected recent alert to remain")
	}
}
