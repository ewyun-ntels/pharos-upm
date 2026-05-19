package query

import (
	"testing"
	"time"

	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/alert/model"
	rule_common "ntels.com/pharos/core/pkg/alert/rule/common"
)

// Test verifying with multiple thresholds across time that only active alerts are in status,
// and only changes are sent to history (measured via ClickHouse mock server).
func TestQuery_MultiThresholds_StatusOnlyActive_And_SendOnlyChanges(t *testing.T) {
	cfg, _ := setupSQLite(t)

	r := &Rule{
		Name:       "multi_thresholds",
		Datasource: "ds",
		DatasourceQuery: Datasource{
			Query:         "SELECT 1",
			TimeLabel:     "ts",
			VariableLabel: "val",
		},
		EvaluationInterval: "* * * * *",
		CheckType:          CheckTypeLast,
		// Three thresholds: below 20 (Minor), 50-80 inclusive (Major), above 80 (Critical)
		Threshold: []Threshold{
			{Id: "t-low", Operation: OperationIsBelow, StartValue: 20, Severity: rule_common.SeverityMinor},
			{Id: "t-mid", Operation: OperationWithinRange, StartValue: 50, EndValue: 80, Severity: rule_common.SeverityMajor},
			{Id: "t-high", Operation: OperationIsAbove, StartValue: 80, Severity: rule_common.SeverityCritical},
		},
		Notifications: nil,
	}
	r.SetConfig(cfg)

	baseTs := time.Now()
	mkTs := func(d time.Duration) string { return baseTs.Add(d).Format(time.DateTime) }

	type step struct {
		val        string
		wantHist   int // number of alerts changed in this step
		wantActive int // number of active alerts in status after this step
	}

	steps := []step{
		{val: "10", wantHist: 1, wantActive: 1}, // low occurs
		{val: "60", wantHist: 2, wantActive: 1}, // low clears, mid occurs
		{val: "90", wantHist: 2, wantActive: 1}, // mid clears, high occurs
		{val: "85", wantHist: 0, wantActive: 1}, // still high -> no change
		{val: "15", wantHist: 2, wantActive: 1}, // high clears, low occurs
		{val: "55", wantHist: 2, wantActive: 1}, // low clears, mid occurs
	}

	prevHist := 0
	prevHistRow := 0
	for i, s := range steps {
		data := &orm.DatabaseResponse{
			Meta: []map[string]string{{"name": r.DatasourceQuery.TimeLabel}, {"name": r.DatasourceQuery.VariableLabel}},
			Data: []map[string]any{{r.DatasourceQuery.TimeLabel: mkTs(time.Duration(i) * time.Minute), r.DatasourceQuery.VariableLabel: s.val}},
			Rows: 1,
		}
		if err := r.AlertCheck(data); err != nil {
			t.Fatalf("AlertCheck step %d failed: %v", i+1, err)
		}

		m := model.Model{Config: &cfg}
		alerts, err := m.GetAllAlert(r.Name)
		if err != nil {
			t.Fatalf("GetAllAlert step %d failed: %v", i+1, err)
		}
		if got := len(alerts); got != s.wantActive {
			t.Fatalf("step %d: expected %d active alerts, got %d", i+1, s.wantActive, got)
		}

		// History rows persisted to two tables; compute delta counts from DB
		hc, rc := countHistory(t, &cfg)
		deltaH := hc - prevHist
		deltaR := rc - prevHistRow
		// history_alert_row appends every change
		if deltaR != s.wantHist {
			t.Fatalf("step %d: expected %d history rows to history_alert_row, got %d", i+1, s.wantHist, deltaR)
		}
		// history_alert uses upsert semantics (by id, with version gating)
		// So per-step delta can be less than total changes if multiple changes affect the same id
		if s.wantHist == 0 {
			if deltaH != 0 {
				t.Fatalf("step %d: expected 0 history rows to history_alert, got %d", i+1, deltaH)
			}
		} else {
			// For SQLite/PostgreSQL upsert-by-id semantics, updates to existing IDs do not increase row count.
			// Therefore, allow deltaH to be in [0, wantHist].
			if deltaH < 0 || deltaH > s.wantHist {
				t.Fatalf("step %d: expected between 0 and %d history rows to history_alert (upsert), got %d", i+1, s.wantHist, deltaH)
			}
		}
		prevHist = hc
		prevHistRow = rc
	}
}
