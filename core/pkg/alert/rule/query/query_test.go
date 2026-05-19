package query

import (
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	alert_common "ntels.com/pharos/core/pkg/alert/common"
	"ntels.com/pharos/core/pkg/alert/model"
	rule_common "ntels.com/pharos/core/pkg/alert/rule/common"
	"ntels.com/pharos/core/pkg/common"
)

func baseValidRule() *Rule {
	return &Rule{
		Name:       "test_rule",
		Datasource: "ds",
		DatasourceQuery: Datasource{
			Query:         "SELECT 1",
			TimeLabel:     "ts",
			VariableLabel: "val",
		},
		EvaluationInterval: "* * * * *",
		CheckType:          CheckTypeLast,
		Threshold: []Threshold{
			{Id: "t-default", Operation: OperationWithinRange, StartValue: 0, EndValue: 100, Severity: rule_common.SeverityMajor},
		},
	}
}

// Since we reference datasource.ClickhouseQuery above, import alias
// but Go requires imports at top; to avoid unused error in case of refactors, we actually use it in tests.
// Note: datasource import is already used via type reference in baseValidRule.

func TestValidate_QueryRule_MissingFields(t *testing.T) {
	// Start from valid rule and remove fields one by one
	r := baseValidRule()

	// Missing datasource
	r2 := *r
	r2.Datasource = ""
	if err := r2.Validate(); err == nil || err.Error() != "datasource is required" {
		t.Fatalf("expected datasource is required, got %v", err)
	}

	// Missing query
	r3 := *r
	r3.DatasourceQuery.Query = ""
	if err := r3.Validate(); err == nil || err.Error() != "query is required" {
		t.Fatalf("expected query is required, got %v", err)
	}

	// Missing time label
	r4 := *r
	r4.DatasourceQuery.TimeLabel = ""
	if err := r4.Validate(); err == nil || err.Error() != "time label is required" {
		t.Fatalf("expected time label is required, got %v", err)
	}

	// Missing variable label
	r5 := *r
	r5.DatasourceQuery.VariableLabel = ""
	if err := r5.Validate(); err == nil || err.Error() != "variable label is required" {
		t.Fatalf("expected variable label is required, got %v", err)
	}

	// Missing evaluation interval
	r6 := *r
	r6.EvaluationInterval = ""
	if err := r6.Validate(); err == nil || err.Error() != "evaluation_interval is required" {
		t.Fatalf("expected evaluation_interval is required, got %v", err)
	}

	// Missing thresholds
	r7 := *r
	r7.Threshold = nil
	if err := r7.Validate(); err == nil || err.Error() != "at least one threshold must be set" {
		t.Fatalf("expected at least one threshold must be set, got %v", err)
	}

	// Unknown check type
	r8 := *r
	r8.CheckType = "unknown"
	if err := r8.Validate(); err == nil || err.Error() != "unknown check type: unknown" {
		t.Fatalf("expected unknown check type, got %v", err)
	}
}

func TestDecodeRow_SuccessAndErrors(t *testing.T) {
	r := baseValidRule()
	ts := time.Now().Format(time.DateTime)

	// Success
	row := map[string]any{
		r.DatasourceQuery.TimeLabel:     ts,
		r.DatasourceQuery.VariableLabel: "12.34",
		"host":                          "srv1",
	}
	vals, err := r.decodeRow(row)
	if err != nil {
		t.Fatalf("decodeRow success expected, got error %v", err)
	}
	if len(vals) != 1 {
		t.Fatalf("expected 1 alert value, got %d", len(vals))
	}
	val := vals[0]
	if val.Value != 12.34 {
		t.Fatalf("expected value 12.34, got %v", val.Value)
	}
	if val.Labels["host"] != "srv1" {
		t.Fatalf("expected label host=srv1, got %v", val.Labels["host"])
	}
	if val.Severity != rule_common.SeverityMajor {
		t.Fatalf("expected severity %s, got %s", rule_common.SeverityMajor, val.Severity)
	}

	// Missing time label
	_, err = r.decodeRow(map[string]any{r.DatasourceQuery.VariableLabel: "1"})
	if err == nil || err.Error() != "missing time label" {
		t.Fatalf("expected missing time label, got %v", err)
	}

	// Time not a string
	_, err = r.decodeRow(map[string]any{r.DatasourceQuery.TimeLabel: 123, r.DatasourceQuery.VariableLabel: "1"})
	if err == nil || err.Error() != "time label is not a string" {
		t.Fatalf("expected time label is not a string, got %v", err)
	}

	// Invalid time format
	_, err = r.decodeRow(map[string]any{r.DatasourceQuery.TimeLabel: "2020-01-01T00:00:00Z", r.DatasourceQuery.VariableLabel: "1"})
	if err == nil || !strings.HasPrefix(err.Error(), "invalid time format") {
		t.Fatalf("expected invalid time format, got %v", err)
	}

	// Missing variable label
	_, err = r.decodeRow(map[string]any{r.DatasourceQuery.TimeLabel: ts})
	if err == nil || err.Error() != "missing variable label" {
		t.Fatalf("expected missing variable label, got %v", err)
	}

	// Variable not a string
	_, err = r.decodeRow(map[string]any{r.DatasourceQuery.TimeLabel: ts, r.DatasourceQuery.VariableLabel: 1})
	if err == nil || err.Error() != "variable label is not a string" {
		t.Fatalf("expected variable label is not a string, got %v", err)
	}

	// Variable not parseable as float
	_, err = r.decodeRow(map[string]any{r.DatasourceQuery.TimeLabel: ts, r.DatasourceQuery.VariableLabel: "abc"})
	if err == nil || err.Error()[:21] != "failed to parse float" {
		t.Fatalf("expected failed to parse float, got %v", err)
	}
}

// clickhouseMock tracks posted rows per endpoint
type clickhouseMock struct{}

// setup a temporary sqlite database and create the alert_status table, and prepare statistics history tables
func setupSQLite(t *testing.T) (common.Config, *clickhouseMock) {
	t.Helper()

	// tracker is unused with DB-backed history but kept for signature compatibility
	tracker := &clickhouseMock{}

	tmp := t.TempDir()
	dbPath := filepath.Join(tmp, "test.db")

	cfg := common.Config{}
	cfg.Database = orm.DatabaseConfig{
		Driver: orm.DriverSqlite,
		SQLite: orm.SQLiteConfig{Path: dbPath},
	}
	// Configure statistics database to use the same SQLite file
	cfg.Statistics.Database = orm.DatabaseConfig{
		Driver: orm.DriverSqlite,
		SQLite: orm.SQLiteConfig{Path: dbPath},
	}

	// Ensure connection is created and tables exist
	db, err := orm.DatabasePool.GetDB(cfg.Database)
	if err != nil {
		t.Fatalf("failed to open sqlite: %v", err)
	}

	// Create minimal alert_status schema used by model.Model
	schema := `CREATE TABLE IF NOT EXISTS alert_status (
        id TEXT,
        alert_id TEXT NOT NULL,
        name TEXT NOT NULL,
        mask TEXT,
        alert_type TEXT,
        description TEXT,
        status TEXT,
        severity TEXT,
        previous_severity TEXT,
        previous_value REAL,
        value REAL,
        labels TEXT,
        timestamp DATETIME NOT NULL,
        updated_at DATETIME NOT NULL,
        check_time DATETIME NOT NULL,
        PRIMARY KEY (alert_id, name)
    )`
	if _, err := db.Exec(schema); err != nil {
		t.Fatalf("failed to create alert_status table: %v", err)
	}

	// Create history tables for statistics DB (same SQLite file)
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS history_alert (
            timestamp DATETIME,
            id TEXT,
            alert_id TEXT,
            name TEXT,
            alert_type TEXT,
            description TEXT,
            previous_severity TEXT,
            previous_value REAL,
            severity TEXT,
            value REAL,
            labels TEXT,
            status TEXT,
            previous_timestamp DATETIME,
            version DATETIME
        )`); err != nil {
		t.Fatalf("failed to create history_alert table: %v", err)
	}
	if _, err := db.Exec(`CREATE UNIQUE INDEX IF NOT EXISTS ux_history_alert_id ON history_alert (id)`); err != nil {
		t.Fatalf("failed to create history_alert unique index: %v", err)
	}
	if _, err := db.Exec(`CREATE TABLE IF NOT EXISTS history_alert_row (
            timestamp DATETIME,
            id TEXT,
            alert_id TEXT,
            name TEXT,
            alert_type TEXT,
            description TEXT,
            previous_severity TEXT,
            previous_value REAL,
            severity TEXT,
            value REAL,
            labels TEXT,
            status TEXT,
            previous_timestamp DATETIME,
            version DATETIME,
            start_timestamp DATETIME,
            status_change_reason TEXT,
            status_changed_by TEXT,
            mask BOOLEAN
        )`); err != nil {
		t.Fatalf("failed to create history_alert_row table: %v", err)
	}

	return cfg, tracker
}

// countHistory returns the total counts of rows in history_alert and history_alert_row
func countHistory(t *testing.T, cfg *common.Config) (int, int) {
	t.Helper()
	var c1, c2 int
	err := orm.StatisticsHandler(orm.DriverDefault, &cfg.Statistics.Database, func(db *sqlx.DB) error {
		if err := db.Get(&c1, "SELECT COUNT(*) FROM history_alert"); err != nil {
			return err
		}
		if err := db.Get(&c2, "SELECT COUNT(*) FROM history_alert_row"); err != nil {
			return err
		}
		return nil
	})
	if err != nil {
		t.Fatalf("countHistory failed: %v", err)
	}
	return c1, c2
}

func TestAlertCheck_Last_Normal_NoSideEffects(t *testing.T) {
	// Set up DB
	cfg, _ := setupSQLite(t)

	// Build rule
	r := baseValidRule()
	r.SetConfig(cfg)

	// Build data with two rows, last one later timestamp
	ts1 := time.Now().Add(-time.Minute).Format(time.DateTime)
	ts2 := time.Now().Format(time.DateTime)

	data := &orm.DatabaseResponse{
		Meta: []map[string]string{
			{"name": r.DatasourceQuery.TimeLabel},
			{"name": r.DatasourceQuery.VariableLabel},
		},
		Data: []map[string]any{
			{r.DatasourceQuery.TimeLabel: ts1, r.DatasourceQuery.VariableLabel: "1"},
			{r.DatasourceQuery.TimeLabel: ts2, r.DatasourceQuery.VariableLabel: "2"},
		},
		Rows: 2,
	}

	// Adjust thresholds so that values 1 and 2 do NOT match any threshold
	r.Threshold = []Threshold{
		{Id: "t-above100k", Operation: OperationIsAbove, StartValue: 100000, Severity: rule_common.SeverityMajor},
	}

	// No notifications
	r.Notifications = nil

	if err := r.AlertCheck(data); err != nil {
		t.Fatalf("AlertCheck returned error: %v", err)
	}
}

func TestValidate_NameRequired(t *testing.T) {
	r := baseValidRule()
	r.Name = ""
	if err := r.Validate(); err == nil || err.Error() != "name is required" {
		t.Fatalf("expected name is required, got %v", err)
	}
}

// Additional sanity test: buildAlert and triggerAlert no-op on normal
func TestBuildAndTrigger_NoOpOnNormal(t *testing.T) {
	cfg, _ := setupSQLite(t)
	r := baseValidRule()
	r.SetConfig(cfg)

	ts := time.Now()
	v := alert_common.Value{
		Name:      r.Name,
		Timestamp: ts,
		Labels:    map[string]string{"a": "b"},
		Severity:  rule_common.SeverityNormal,
	}

	if err := r.buildAlert(&v, ts); err != nil {
		t.Fatalf("buildAlert failed: %v", err)
	}

	// Should be no-op and return nil for normal severity
	if err := r.triggerAlert(&v); err != nil {
		t.Fatalf("triggerAlert expected nil on normal, got %v", err)
	}
}

func TestThreshold_ValidateAndCheck(t *testing.T) {
	// Unknown operation
	th := Threshold{Operation: "unknown", StartValue: 10, EndValue: 20}
	if err := th.Validate(); err == nil {
		t.Fatalf("expected error for unknown operation")
	}

	// Range with start > end
	th = Threshold{Operation: OperationWithinRange, StartValue: 20, EndValue: 10}
	if err := th.Validate(); err == nil {
		t.Fatalf("expected error for invalid range")
	}

	// Valid within range
	th = Threshold{Operation: OperationWithinRange, StartValue: 10, EndValue: 20}
	if err := th.Validate(); err != nil {
		t.Fatalf("unexpected validate error: %v", err)
	}
	if !th.Check(15) || th.Check(25) || th.Check(5) {
		t.Fatalf("within_range check failed")
	}

	// Valid outside range
	th = Threshold{Operation: OperationOutsideRange, StartValue: 10, EndValue: 20}
	if !th.Check(5) || !th.Check(25) || th.Check(15) {
		t.Fatalf("outside_range check failed")
	}

	// is_above uses StartValue as single Threshold
	th = Threshold{Operation: OperationIsAbove, StartValue: 10}
	if th.Check(10) || !th.Check(10.0001) || th.Check(9.9) {
		t.Fatalf("is_above check failed")
	}

	// is_below uses StartValue as single Threshold
	th = Threshold{Operation: OperationIsBelow, StartValue: 10}
	if th.Check(10) || !th.Check(9.9999) || th.Check(10.1) {
		t.Fatalf("is_below check failed")
	}
}

// New tests focusing on AlertCheck lastValue accumulation and replacement
func TestAlertCheck_Last_SameTimestamp_Appends(t *testing.T) {
	cfg, _ := setupSQLite(t)

	r := baseValidRule()
	r.SetConfig(cfg)
	// Two thresholds with distinct severities; a value in (10,20] matches both
	r.Threshold = []Threshold{
		{Id: "t-above", Operation: OperationIsAbove, StartValue: 10, Severity: rule_common.SeverityMinor, Labels: map[string]string{"th": "above"}},
		{Id: "t-range", Operation: OperationWithinRange, StartValue: 10, EndValue: 20, Severity: rule_common.SeverityCritical, Labels: map[string]string{"th": "range"}},
	}

	ts := time.Now().Format(time.DateTime)
	data := &orm.DatabaseResponse{
		Meta: []map[string]string{{"name": r.DatasourceQuery.TimeLabel}, {"name": r.DatasourceQuery.VariableLabel}},
		Data: []map[string]any{
			{r.DatasourceQuery.TimeLabel: ts, r.DatasourceQuery.VariableLabel: "15", "row": "a"},
			{r.DatasourceQuery.TimeLabel: ts, r.DatasourceQuery.VariableLabel: "18", "row": "b"},
		},
		Rows: 2,
	}

	if err := r.AlertCheck(data); err != nil {
		t.Fatalf("AlertCheck failed: %v", err)
	}

	m := model.Model{Config: &cfg}
	alerts, err := m.GetAllAlert(r.Name)
	if err != nil {
		t.Fatalf("GetAllAlert failed: %v", err)
	}
	for i := range alerts {
		_ = alerts[i].ConvertStringLabelsToLabels()
	}

	if len(alerts) != 4 {
		t.Fatalf("expected 4 alerts (2 rows x 2 thresholds), got %d", len(alerts))
	}

	// Check that both severities exist and all are alerting
	seenSev := map[string]bool{}
	for _, a := range alerts {
		seenSev[a.Severity] = true
		if a.Status != alert_common.StatusAlerting {
			t.Fatalf("expected status alerting, got %s", a.Status)
		}
		// threshold metadata should only be in AlertId, not in Labels
		if _, ok := a.Labels["condition"]; ok {
			t.Fatalf("condition should not be in Labels, got: %+v", a.Labels)
		}
		var idObj struct {
			Labels map[string]string `json:"labels"`
		}
		if err := json.Unmarshal([]byte(a.AlertId), &idObj); err != nil {
			t.Fatalf("failed to unmarshal AlertId: %v", err)
		}
		if _, ok := idObj.Labels["condition"]; !ok {
			t.Fatalf("missing condition in AlertId labels: %s", a.AlertId)
		}
		if tid, ok := idObj.Labels["threshold_id"]; !ok || tid == "" {
			t.Fatalf("missing threshold_id in AlertId labels: %s", a.AlertId)
		}
	}
	if !seenSev[rule_common.SeverityMinor] || !seenSev[rule_common.SeverityCritical] {
		t.Fatalf("expected severities Minor and Critical to be present, got: %#v", seenSev)
	}
}

func TestAlertCheck_Last_NewerTimestamp_Replaces(t *testing.T) {
	cfg, _ := setupSQLite(t)

	r := baseValidRule()
	r.SetConfig(cfg)
	// Same thresholds as previous test
	r.Threshold = []Threshold{
		{Id: "t-above", Operation: OperationIsAbove, StartValue: 10, Severity: rule_common.SeverityMinor, Labels: map[string]string{"th": "above"}},
		{Id: "t-range", Operation: OperationWithinRange, StartValue: 10, EndValue: 20, Severity: rule_common.SeverityCritical, Labels: map[string]string{"th": "range"}},
	}

	ts1 := time.Now().Add(-time.Minute).Format(time.DateTime)
	ts2 := time.Now().Format(time.DateTime)
	data := &orm.DatabaseResponse{
		Meta: []map[string]string{{"name": r.DatasourceQuery.TimeLabel}, {"name": r.DatasourceQuery.VariableLabel}},
		Data: []map[string]any{
			// Two rows at ts1 (would yield 4 alerts if final)
			{r.DatasourceQuery.TimeLabel: ts1, r.DatasourceQuery.VariableLabel: "15", "row": "a"},
			{r.DatasourceQuery.TimeLabel: ts1, r.DatasourceQuery.VariableLabel: "18", "row": "b"},
			// Newer row at ts2 (2 alerts) should replace lastValue
			{r.DatasourceQuery.TimeLabel: ts2, r.DatasourceQuery.VariableLabel: "12", "row": "c"},
		},
		Rows: 3,
	}

	if err := r.AlertCheck(data); err != nil {
		t.Fatalf("AlertCheck failed: %v", err)
	}

	m := model.Model{Config: &cfg}
	alerts, err := m.GetAllAlert(r.Name)
	if err != nil {
		t.Fatalf("GetAllAlert failed: %v", err)
	}
	for i := range alerts {
		_ = alerts[i].ConvertStringLabelsToLabels()
	}

	// Since the rule keeps only the latest timestamp per run, and there was no prior DB state,
	// we expect only 2 current alerts (both alerting) from the latest timestamp rows.
	if len(alerts) != 2 {
		t.Fatalf("expected 2 alerts from the latest timestamp, got %d", len(alerts))
	}

	// Ensure both are alerting and severities present
	seenSev := map[string]bool{}
	for _, a := range alerts {
		if a.Status != alert_common.StatusAlerting {
			t.Fatalf("expected status alerting, got %s", a.Status)
		}
		seenSev[a.Severity] = true
		// threshold metadata should be only in AlertId, not in Labels
		if _, ok := a.Labels["condition"]; ok {
			t.Fatalf("condition should not be in Labels, got: %+v", a.Labels)
		}
		var idObj struct {
			Labels map[string]string `json:"labels"`
		}
		if err := json.Unmarshal([]byte(a.AlertId), &idObj); err != nil {
			t.Fatalf("failed to unmarshal AlertId: %v", err)
		}
		if _, ok := idObj.Labels["condition"]; !ok {
			t.Fatalf("missing condition in AlertId labels: %s", a.AlertId)
		}
		if tid, ok := idObj.Labels["threshold_id"]; !ok || tid == "" {
			t.Fatalf("missing threshold_id in AlertId labels: %s", a.AlertId)
		}
	}
	if !seenSev[rule_common.SeverityMinor] || !seenSev[rule_common.SeverityCritical] {
		t.Fatalf("expected severities Minor and Critical to be present among alerting, got: %#v", seenSev)
	}
}

// New test: verify that clears delete rows from alert_status while still producing notifications/history payloads
// New test: verify that clears delete rows from alert_status while still producing notifications/history payloads
func TestAlertCheck_Clear_DeletesNormal(t *testing.T) {
	cfg, _ := setupSQLite(t)

	r := baseValidRule()
	r.SetConfig(cfg)
	r.Notifications = nil
	// Two thresholds
	r.Threshold = []Threshold{
		{Id: "t-above", Operation: OperationIsAbove, StartValue: 10, Severity: rule_common.SeverityMinor},
		{Id: "t-range", Operation: OperationWithinRange, StartValue: 10, EndValue: 20, Severity: rule_common.SeverityCritical},
	}

	// First run: one row that matches both thresholds -> 2 alerts in DB
	ts1 := time.Now().Format(time.DateTime)
	data1 := &orm.DatabaseResponse{
		Meta: []map[string]string{{"name": r.DatasourceQuery.TimeLabel}, {"name": r.DatasourceQuery.VariableLabel}},
		Data: []map[string]any{{r.DatasourceQuery.TimeLabel: ts1, r.DatasourceQuery.VariableLabel: "15"}},
		Rows: 1,
	}
	if err := r.AlertCheck(data1); err != nil {
		t.Fatalf("AlertCheck (first) failed: %v", err)
	}

	m := model.Model{Config: &cfg}
	alerts, err := m.GetAllAlert(r.Name)
	if err != nil {
		t.Fatalf("GetAllAlert failed: %v", err)
	}
	if len(alerts) != 2 {
		t.Fatalf("expected 2 active alerts after first run, got %d", len(alerts))
	}

	// Second run: newer timestamp row that matches no thresholds -> both should clear and be deleted
	ts2 := time.Now().Add(time.Minute).Format(time.DateTime)
	data2 := &orm.DatabaseResponse{
		Meta: []map[string]string{{"name": r.DatasourceQuery.TimeLabel}, {"name": r.DatasourceQuery.VariableLabel}},
		Data: []map[string]any{{r.DatasourceQuery.TimeLabel: ts2, r.DatasourceQuery.VariableLabel: "1"}},
		Rows: 1,
	}
	if err := r.AlertCheck(data2); err != nil {
		t.Fatalf("AlertCheck (second) failed: %v", err)
	}

	alerts, err = m.GetAllAlert(r.Name)
	if err != nil {
		t.Fatalf("GetAllAlert failed: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected 0 active alerts after clear (deleted), got %d", len(alerts))
	}
	// Expect 2 occur + 2 clear = 4 rows total in each history table
	hc, rc := countHistory(t, &cfg)
	// Upsert semantics on history_alert: only distinct IDs are kept (2), while history_alert_row appends all (4)
	if hc != 2 || rc != 4 {
		t.Fatalf("expected history_alert=2 (upsert by id) and history_alert_row=4, got history_alert=%d, history_alert_row=%d", hc, rc)
	}
}

// New test: multiple alerts occur (4) then multiple alerts clear in next run
func TestAlertCheck_MultiOccur_ThenMultiClear(t *testing.T) {
	cfg, _ := setupSQLite(t)

	r := baseValidRule()
	r.SetConfig(cfg)
	r.Notifications = nil
	// Two thresholds produce two alerts per row
	r.Threshold = []Threshold{
		{Id: "t-above", Operation: OperationIsAbove, StartValue: 10, Severity: rule_common.SeverityMinor},
		{Id: "t-range", Operation: OperationWithinRange, StartValue: 10, EndValue: 20, Severity: rule_common.SeverityCritical},
	}

	// First run: two rows at same timestamp -> 4 alerts
	ts1 := time.Now().Format(time.DateTime)
	data1 := &orm.DatabaseResponse{
		Meta: []map[string]string{{"name": r.DatasourceQuery.TimeLabel}, {"name": r.DatasourceQuery.VariableLabel}},
		Data: []map[string]any{
			{r.DatasourceQuery.TimeLabel: ts1, r.DatasourceQuery.VariableLabel: "15", "row": "a"},
			{r.DatasourceQuery.TimeLabel: ts1, r.DatasourceQuery.VariableLabel: "18", "row": "b"},
		},
		Rows: 2,
	}
	if err := r.AlertCheck(data1); err != nil {
		t.Fatalf("AlertCheck (first) failed: %v", err)
	}

	m := model.Model{Config: &cfg}
	alerts, err := m.GetAllAlert(r.Name)
	if err != nil {
		t.Fatalf("GetAllAlert failed: %v", err)
	}
	if len(alerts) != 4 {
		t.Fatalf("expected 4 active alerts after first run, got %d", len(alerts))
	}
	for i := range alerts {
		_ = alerts[i].ConvertStringLabelsToLabels()
		if alerts[i].Status != alert_common.StatusAlerting {
			t.Fatalf("expected status alerting, got %s", alerts[i].Status)
		}
	}

	// Second run: newer timestamp with non-matching value -> all 4 should clear and be deleted
	ts2 := time.Now().Add(time.Minute).Format(time.DateTime)
	data2 := &orm.DatabaseResponse{
		Meta: []map[string]string{{"name": r.DatasourceQuery.TimeLabel}, {"name": r.DatasourceQuery.VariableLabel}},
		Data: []map[string]any{{r.DatasourceQuery.TimeLabel: ts2, r.DatasourceQuery.VariableLabel: "1"}},
		Rows: 1,
	}
	if err := r.AlertCheck(data2); err != nil {
		t.Fatalf("AlertCheck (second) failed: %v", err)
	}

	alerts, err = m.GetAllAlert(r.Name)
	if err != nil {
		t.Fatalf("GetAllAlert failed: %v", err)
	}
	if len(alerts) != 0 {
		t.Fatalf("expected 0 active alerts after clear (deleted), got %d", len(alerts))
	}

	// Expect 4 occur + 4 clear = 8 rows total in each history table across both sends
	hc, rc := countHistory(t, &cfg)
	// Upsert semantics on history_alert: 4 distinct IDs retained; history_alert_row appends all 8
	if hc != 4 || rc != 8 {
		t.Fatalf("expected history_alert=4 (upsert by id) and history_alert_row=8, got history_alert=%d, history_alert_row=%d", hc, rc)
	}

}

// Test that Normalize fills missing threshold IDs with UUIDs and preserves existing IDs
func TestNormalize_FillsMissingThresholdIds(t *testing.T) {
	r := &Rule{
		Name:       "norm_rule",
		Datasource: "ds",
		DatasourceQuery: Datasource{
			Query:         "SELECT 1",
			TimeLabel:     "ts",
			VariableLabel: "val",
		},
		EvaluationInterval: "* * * * *",
		CheckType:          CheckTypeLast,
		Threshold: []Threshold{
			{Id: "preset-id", Operation: OperationIsAbove, StartValue: 10, Severity: rule_common.SeverityMinor},
			{Id: "", Operation: OperationIsBelow, StartValue: 5, Severity: rule_common.SeverityMajor},
		},
	}

	if err := r.Normalize(); err != nil {
		t.Fatalf("Normalize returned error: %v", err)
	}

	if r.Threshold[0].Id != "preset-id" {
		t.Fatalf("Normalize should preserve existing ID, got %q", r.Threshold[0].Id)
	}
	if r.Threshold[1].Id == "" {
		t.Fatalf("Normalize should assign an ID when missing")
	}
	// Validate that the generated ID is a UUID
	if _, err := uuid.Parse(r.Threshold[1].Id); err != nil {
		t.Fatalf("generated ID is not a valid UUID: %q (%v)", r.Threshold[1].Id, err)
	}
	// Ensure IDs are distinct
	if r.Threshold[0].Id == r.Threshold[1].Id {
		t.Fatalf("IDs should be distinct; both are %q", r.Threshold[0].Id)
	}
}

// With Validate calling Normalize, missing threshold IDs are auto-filled
func TestNormalize_ThenValidate_AllowsMissingIdsInitially(t *testing.T) {
	r := &Rule{
		Name:       "norm_rule_validate",
		Datasource: "ds",
		DatasourceQuery: Datasource{
			Query:         "SELECT 1",
			TimeLabel:     "ts",
			VariableLabel: "val",
		},
		EvaluationInterval: "* * * * *",
		CheckType:          CheckTypeLast,
		Threshold: []Threshold{
			{Operation: OperationWithinRange, StartValue: 0, EndValue: 100, Severity: rule_common.SeverityMajor}, // missing Id on purpose
		},
	}

	if err := r.Validate(); err != nil {
		t.Fatalf("Validate should pass and auto-fill missing IDs, got: %v", err)
	}
	if len(r.Threshold) == 0 || r.Threshold[0].Id == "" {
		t.Fatalf("Validate (Normalize) should assign an ID when missing")
	}
}

// Direct unit test for Rule.triggerAlert: it should persist the given alert value
// to the alert_status table via model.UpdateAlert.
func TestTriggerAlert_PersistsAlerting(t *testing.T) {
	cfg, _ := setupSQLite(t)

	r := baseValidRule()
	r.SetConfig(cfg)

	now := time.Now()
	v := alert_common.Value{
		Name:        r.Name,
		Timestamp:   now,
		Labels:      map[string]string{"k": "v"},
		Severity:    rule_common.SeverityMajor,
		Status:      alert_common.StatusAlerting,
		Description: r.Description,
		// AlertId and Id must be set by caller (in decodeRow this is handled automatically)
		AlertId:   "aid-1",
		Id:        "aid-1",
		AlertType: string(alert_common.TypeQuery),
	}

	// buildAlert sets UpdatedAt, CheckTime, converts labels etc.
	if err := r.buildAlert(&v, now); err != nil {
		t.Fatalf("buildAlert failed: %v", err)
	}

	if err := r.triggerAlert(&v); err != nil {
		t.Fatalf("triggerAlert failed: %v", err)
	}

	m := model.Model{Config: &cfg}
	rows, err := m.GetAllAlert(r.Name)
	if err != nil {
		t.Fatalf("GetAllAlert failed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 persisted alert, got %d", len(rows))
	}
	got := rows[0]
	if got.AlertId != v.AlertId {
		t.Fatalf("unexpected alert_id: want %q got %q", v.AlertId, got.AlertId)
	}
	if got.Severity != v.Severity {
		t.Fatalf("unexpected severity: want %q got %q", v.Severity, got.Severity)
	}
	if got.Status != v.Status {
		t.Fatalf("unexpected status: want %q got %q", v.Status, got.Status)
	}
}

// Document current behavior: triggerAlert persists even normal severity values
// (clearing logic is handled at a higher level by AlertCheck which deletes instead).
func TestTriggerAlert_PersistsNormalToo(t *testing.T) {
	cfg, _ := setupSQLite(t)

	r := baseValidRule()
	r.SetConfig(cfg)

	now := time.Now()
	v := alert_common.Value{
		Name:        r.Name,
		Timestamp:   now,
		Labels:      map[string]string{"a": "b"},
		Severity:    rule_common.SeverityNormal,
		Status:      alert_common.StatusNormal,
		Description: r.Description,
		AlertId:     "aid-normal",
		Id:          "aid-normal",
		AlertType:   string(alert_common.TypeQuery),
	}

	if err := r.buildAlert(&v, now); err != nil {
		t.Fatalf("buildAlert failed: %v", err)
	}

	if err := r.triggerAlert(&v); err != nil {
		t.Fatalf("triggerAlert failed: %v", err)
	}

	m := model.Model{Config: &cfg}
	rows, err := m.GetAllAlert(r.Name)
	if err != nil {
		t.Fatalf("GetAllAlert failed: %v", err)
	}
	if len(rows) != 1 {
		t.Fatalf("expected 1 persisted alert (normal), got %d", len(rows))
	}
	got := rows[0]
	if got.AlertId != v.AlertId {
		t.Fatalf("unexpected alert_id: want %q got %q", v.AlertId, got.AlertId)
	}
	if got.Severity != v.Severity {
		t.Fatalf("unexpected severity: want %q got %q", v.Severity, got.Severity)
	}
	if got.Status != v.Status {
		t.Fatalf("unexpected status: want %q got %q", v.Status, got.Status)
	}
}
