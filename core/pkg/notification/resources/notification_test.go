package resources

import (
	"encoding/json"
	"testing"
	"time"

	"ntels.com/pharos/core/external/orm"
	notifcommon "ntels.com/pharos/core/pkg/notification/common"
	"ntels.com/pharos/core/pkg/notification/model"
)

func TestRuleBuildFromDb_Success(t *testing.T) {
	// arrange
	ruleMap := map[string]any{"foo": "bar", "num": float64(123)}
	ruleBytes, _ := json.Marshal(ruleMap)
	ts := time.Now().Add(-time.Hour)
	upd := time.Now()
	m := &model.Rule{
		ID:               "id-1",
		NotificationType: "snmp",
		Name:             "name",
		Rule:             string(ruleBytes),
		Timestamp:        orm.Datetime{Time: ts},
		UpdatedAt:        orm.Datetime{Time: upd},
	}

	var r Rule
	// act
	if err := r.BuildFromDb(m); err != nil {
		t.Fatalf("expected no error, got %v", err)
	}

	// assert
	if r.NotificationType != notifcommon.Type(m.NotificationType) {
		t.Errorf("notification type mismatch: %v != %v", r.NotificationType, m.NotificationType)
	}
	if r.Timestamp.Unix() != ts.Unix() {
		t.Errorf("timestamp mismatch: %v != %v", r.Timestamp, ts)
	}
	if r.UpdatedAt.Unix() != upd.Unix() {
		t.Errorf("updatedAt mismatch: %v != %v", r.UpdatedAt, upd)
	}
	if got := r.Rule["foo"]; got != "bar" {
		t.Errorf("rule field foo mismatch: %v", got)
	}
	if got := r.Rule["num"]; got != float64(123) {
		t.Errorf("rule field num mismatch: %v", got)
	}
}

func TestRuleBuildFromDb_InvalidJSON(t *testing.T) {
	m := &model.Rule{NotificationType: "snmp", Rule: "{invalid json}", Timestamp: orm.Datetime{Time: time.Now()}, UpdatedAt: orm.Datetime{Time: time.Now()}}
	var r Rule
	if err := r.BuildFromDb(m); err == nil {
		t.Fatalf("expected error for invalid json")
	}
}
