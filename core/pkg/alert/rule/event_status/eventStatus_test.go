package event_status

import (
	"testing"
	"time"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	alert_common "ntels.com/pharos/core/pkg/alert/common"
	"ntels.com/pharos/core/pkg/alert/model"
	"ntels.com/pharos/core/pkg/alert/resources"
	"ntels.com/pharos/core/pkg/common"
)

// helper to setup a sqlite database and schema for tests (mirrors model tests)
func newTestConfig(t *testing.T) (common.Config, func()) {
	t.Helper()

	cfg := common.Config{}
	cfg.Database.Driver = orm.DriverSqlite
	cfg.Database.SQLite.Path = t.TempDir() + "/test.db"

	// initialize schema required by event status logic
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
			// minimal tables used by model tests are not strictly needed here, but kept consistent
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
	r := &Rule{Name: "cpu_high"}
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
		Value:       80,
		Labels:      map[string]string{"host": "h1"},
	}
	if err := r.EventHandler(ev); err == nil {
		t.Fatalf("expected error for empty AlertId, got nil")
	}
}

func TestEventHandler_NewAlert_NonNormal(t *testing.T) {
	r, cleanup, m := newRuleForTest(t)
	defer cleanup()
	ev := &resources.Event{
		AlertId:     "A-1",
		Severity:    "Major",
		Description: "desc",
		Value:       90,
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
		t.Fatalf("expected alert to be created")
	}
	if got.Severity != "Major" {
		t.Fatalf("unexpected severity: %s", got.Severity)
	}
	if got.Status != alert_common.StatusAlerting {
		t.Fatalf("unexpected status: %s", got.Status)
	}
}

func TestEventHandler_NewEvent_Normal_NoExisting(t *testing.T) {
	r, cleanup, m := newRuleForTest(t)
	defer cleanup()
	ev := &resources.Event{
		AlertId:     "A-2",
		Severity:    "Normal",
		Description: "desc",
		Value:       10,
		Labels:      map[string]string{"host": "h2"},
	}
	if err := r.EventHandler(ev); err != nil {
		t.Fatalf("EventHandler error: %v", err)
	}
	got, err := m.GetAlert(r.Name, ev.AlertId)
	if err != nil {
		t.Fatalf("GetAlert error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected no alert stored for normal severity, got: %+v", got)
	}
}

func TestEventHandler_ExistingSameSeverity_Update(t *testing.T) {
	r, cleanup, m := newRuleForTest(t)
	defer cleanup()

	// Seed an existing alert with Major severity
	now := time.Now().UTC()
	ct := now
	seed := alert_common.Value{
		Id:          "seed-id",
		Timestamp:   now,
		UpdatedAt:   now,
		CheckTime:   &ct,
		Name:        r.Name,
		AlertType:   "event",
		Description: "seed",
		AlertId:     "A-3",
		Value:       80,
		Severity:    "Major",
		Status:      alert_common.StatusAlerting,
		Labels:      map[string]string{"host": "h3"},
	}
	_ = seed.ConvertLabelsToStringLabels()
	if err := m.UpdateAlert(&seed); err != nil {
		t.Fatalf("seed UpdateAlert error: %v", err)
	}

	// Send another event with the same severity but different value
	ev := &resources.Event{
		AlertId:     "A-3",
		Severity:    "Major",
		Description: "desc2",
		Value:       95,
		Labels:      map[string]string{"host": "h3"},
	}
	if err := r.EventHandler(ev); err != nil {
		t.Fatalf("EventHandler error: %v", err)
	}

	got, err := m.GetAlert(r.Name, ev.AlertId)
	if err != nil {
		t.Fatalf("GetAlert error: %v", err)
	}
	if got == nil {
		t.Fatalf("expected alert to remain")
	}
	if got.Id != seed.Id {
		t.Fatalf("expected same id to be preserved, got %s want %s", got.Id, seed.Id)
	}
	if got.Severity != "Major" {
		t.Fatalf("severity changed unexpectedly: %s", got.Severity)
	}
	if got.Value != 95 {
		t.Fatalf("value not updated: %v", got.Value)
	}
	// previous fields should be set
	if got.PreviousSeverity == "" || got.PreviousValue == nil {
		t.Fatalf("expected previous fields to be set, got PrevSeverity=%q PrevValue=%v", got.PreviousSeverity, got.PreviousValue)
	}
}

func TestEventHandler_SeverityChange_ToNormal_Deletes(t *testing.T) {
	r, cleanup, m := newRuleForTest(t)
	defer cleanup()

	// Seed existing non-normal alert
	now := time.Now().UTC()
	ct := now
	seed := alert_common.Value{
		Id:          "seed-id-2",
		Timestamp:   now,
		UpdatedAt:   now,
		CheckTime:   &ct,
		Name:        r.Name,
		AlertType:   "event",
		Description: "seed",
		AlertId:     "A-4",
		Value:       70,
		Severity:    "Minor",
		Status:      alert_common.StatusAlerting,
		Labels:      map[string]string{"host": "h4"},
	}
	_ = seed.ConvertLabelsToStringLabels()
	if err := m.UpdateAlert(&seed); err != nil {
		t.Fatalf("seed UpdateAlert error: %v", err)
	}

	// Now send a Normal event for the same alert
	ev := &resources.Event{
		AlertId:     "A-4",
		Severity:    "Normal",
		Description: "cleared",
		Value:       0,
		Labels:      map[string]string{"host": "h4"},
	}
	if err := r.EventHandler(ev); err != nil {
		t.Fatalf("EventHandler error: %v", err)
	}

	got, err := m.GetAlert(r.Name, ev.AlertId)
	if err != nil {
		t.Fatalf("GetAlert error: %v", err)
	}
	if got != nil {
		t.Fatalf("expected alert to be deleted on Normal severity")
	}
}

func TestEventHandler_RepeatedMajorEvents(t *testing.T) {
	r, cleanup, m := newRuleForTest(t)
	defer cleanup()

	id := "A-5"
	// First Major event
	ev1 := &resources.Event{AlertId: id, Severity: "Major", Description: "m1", Value: 80, Labels: map[string]string{"host": "h5"}}
	if err := r.EventHandler(ev1); err != nil {
		t.Fatalf("EventHandler err on first major: %v", err)
	}
	first, err := m.GetAlert(r.Name, id)
	if err != nil || first == nil {
		t.Fatalf("GetAlert after first major err:%v val:%v", err, first)
	}

	// Second Major event (same severity, higher value)
	ev2 := &resources.Event{AlertId: id, Severity: "Major", Description: "m2", Value: 90, Labels: map[string]string{"host": "h5"}}
	if err := r.EventHandler(ev2); err != nil {
		t.Fatalf("EventHandler err on second major: %v", err)
	}
	second, err := m.GetAlert(r.Name, id)
	if err != nil || second == nil {
		t.Fatalf("GetAlert after second major err:%v val:%v", err, second)
	}
	if second.Id != first.Id {
		t.Fatalf("expected same id across repeated majors: got %s want %s", second.Id, first.Id)
	}
	if second.Severity != "Major" {
		t.Fatalf("severity changed: %s", second.Severity)
	}
	if second.Value != 90 {
		t.Fatalf("value not updated on second major: %v", second.Value)
	}
	if second.PreviousValue == nil || *second.PreviousValue != first.Value {
		t.Fatalf("previous value should be first value: prev=%v first=%v", second.PreviousValue, first.Value)
	}

	// Third Major event (lower value)
	ev3 := &resources.Event{AlertId: id, Severity: "Major", Description: "m3", Value: 50, Labels: map[string]string{"host": "h5"}}
	if err := r.EventHandler(ev3); err != nil {
		t.Fatalf("EventHandler err on third major: %v", err)
	}
	third, err := m.GetAlert(r.Name, id)
	if err != nil || third == nil {
		t.Fatalf("GetAlert after third major err:%v val:%v", err, third)
	}
	if third.Id != first.Id {
		t.Fatalf("id changed on third major: got %s want %s", third.Id, first.Id)
	}
	if third.Severity != "Major" || third.Value != 50 {
		t.Fatalf("unexpected third state: sev=%s val=%v", third.Severity, third.Value)
	}
	if third.PreviousValue == nil || *third.PreviousValue != second.Value {
		t.Fatalf("previous value should equal second value: prev=%v second=%v", third.PreviousValue, second.Value)
	}
}

func TestEventHandler_RepeatedCriticalEvents(t *testing.T) {
	r, cleanup, m := newRuleForTest(t)
	defer cleanup()

	id := "A-6"
	// First Critical event
	ev1 := &resources.Event{AlertId: id, Severity: "Critical", Description: "c1", Value: 95, Labels: map[string]string{"host": "h6"}}
	if err := r.EventHandler(ev1); err != nil {
		t.Fatalf("EventHandler err on first critical: %v", err)
	}
	first, err := m.GetAlert(r.Name, id)
	if err != nil || first == nil {
		t.Fatalf("GetAlert after first critical err:%v val:%v", err, first)
	}

	// Second Critical event
	ev2 := &resources.Event{AlertId: id, Severity: "Critical", Description: "c2", Value: 99, Labels: map[string]string{"host": "h6"}}
	if err := r.EventHandler(ev2); err != nil {
		t.Fatalf("EventHandler err on second critical: %v", err)
	}
	second, err := m.GetAlert(r.Name, id)
	if err != nil || second == nil {
		t.Fatalf("GetAlert after second critical err:%v val:%v", err, second)
	}
	if second.Id != first.Id {
		t.Fatalf("expected same id across repeated criticals: got %s want %s", second.Id, first.Id)
	}
	if second.Severity != "Critical" || second.Value != 99 {
		t.Fatalf("unexpected state after second critical: sev=%s val=%v", second.Severity, second.Value)
	}
	if second.PreviousValue == nil || *second.PreviousValue != first.Value {
		t.Fatalf("previous value should be first value: prev=%v first=%v", second.PreviousValue, first.Value)
	}
}
