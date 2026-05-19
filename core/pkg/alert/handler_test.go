package alert

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/jmoiron/sqlx"
	"github.com/ory/fosite/token/jwt"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/external/orm"
	alert_common "ntels.com/pharos/core/pkg/alert/common"
	"ntels.com/pharos/core/pkg/alert/model"
	"ntels.com/pharos/core/pkg/common"
	alert_types "ntels.com/pharos/shared/types/alert"
)

// 테스트용 전역 config 변수
var config common.Config

// PutStatusClean은 테스트용 wrapper 함수입니다
func PutStatusClean(c *gin.Context) {
	handler := GetPutStatusCleanHandler(config)
	handler(c)
}

// PutStatusMask는 테스트용 wrapper 함수입니다
func PutStatusMask(c *gin.Context) {
	handler := GetPutStatusMaskHandler(config)
	handler(c)
}

// setup a sqlite database schema for alert_status used by model
func newTestDBConfig(t *testing.T) common.Config {
	t.Helper()
	cfg := common.Config{}
	cfg.Database.Driver = orm.DriverSqlite
	cfg.Database.SQLite.Path = t.TempDir() + "/test.db"
	// Use the same SQLite file for statistics DB
	cfg.Statistics.Database.Driver = orm.DriverSqlite
	cfg.Statistics.Database.SQLite.Path = cfg.Database.SQLite.Path

	if err := orm.Handler(orm.DriverSqlite, &cfg.Database, func(db *sqlx.DB) error {
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
			// Minimal statistics history tables used by history writer
			`CREATE TABLE IF NOT EXISTS history_alert (
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
			);`,
			`CREATE UNIQUE INDEX IF NOT EXISTS ux_history_alert_id ON history_alert (id);`,
			`CREATE TABLE IF NOT EXISTS history_alert_row (
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
			);`,
		}
		for _, s := range stmts {
			if _, err := db.Exec(s); err != nil {
				return err
			}
		}
		return nil
	}); err != nil {
		t.Fatalf("schema init failed: %v", err)
	}
	return cfg
}

// setupClickhouseMock configures a mock ClickHouse HTTP server and returns a
// common.Config pointing to it (schema=http) and a tracker for assertions.
type clickhouseTracker struct{ hist, histRow int }

func setupClickhouseMock(t *testing.T, cfg *common.Config) *clickhouseTracker {
	t.Helper()
	tracker := &clickhouseTracker{}
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Count newline-separated JSON objects in body
		body, _ := io.ReadAll(r.Body)
		_ = r.Body.Close()
		lines := 0
		for _, b := range body {
			if b == '\n' {
				lines++
			}
		}
		q := r.URL.Query().Get("query")
		if strings.Contains(q, "history_alert_row") {
			tracker.histRow += lines
		} else if strings.Contains(q, "history_alert") {
			tracker.hist += lines
		}
		w.WriteHeader(http.StatusOK)
	}))
	t.Cleanup(ts.Close)

	// Parse host and port
	url := ts.URL // http://127.0.0.1:PORT
	hostPort := strings.TrimPrefix(url, "http://")
	parts := strings.Split(hostPort, ":")
	host := parts[0]
	port, _ := strconv.Atoi(parts[1])

	cfg.Master.Statistics.Database.Driver = orm.DriverClickHouse
	cfg.Master.Statistics.Database.ClickHouse.Host = host
	cfg.Master.Statistics.Database.ClickHouse.Port = port
	cfg.Master.Statistics.Database.ClickHouse.Database = "test"
	cfg.Master.Statistics.Database.ClickHouse.Username = "u"
	cfg.Master.Statistics.Database.ClickHouse.Password = "p"
	return tracker
}

func seedAlert(t *testing.T, cfg *common.Config, a alert_common.Value) {
	m := &model.Model{Config: cfg}
	_ = a.ConvertLabelsToStringLabels()
	require.NoError(t, m.UpdateAlert(&a))
}

func TestPutStatusClean_SuccessDeletesAndWritesHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newTestDBConfig(t)
	config = cfg // set package global

	// seed an alert
	now := time.Now().UTC()
	ct := now
	alert := alert_common.Value{
		Id:          "id-1",
		Timestamp:   now,
		UpdatedAt:   now,
		CheckTime:   &ct,
		Name:        "ruleA",
		Mask:        false,
		AlertType:   string(alert_common.TypeQuery),
		Description: "desc",
		AlertId:     "AL-1",
		Value:       9.9,
		Status:      alert_common.StatusAlerting,
		Severity:    "Major",
		Labels:      map[string]string{"k": "v"},
	}
	seedAlert(t, &cfg, alert)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{gin.Param{Key: "name", Value: alert.Name}, gin.Param{Key: "id", Value: alert.AlertId}}
	claims := &jwt.JWTClaims{Subject: "tester", Extra: map[string]any{"attrs": map[string]any{}}}
	c.Set(common.ContextKeyJWTClaims, claims)

	PutStatusClean(c)
	require.Equal(t, http.StatusOK, rec.Code)

	// verify DB: alert removed
	m := &model.Model{Config: &cfg}
	got, err := m.GetAlert(alert.Name, alert.AlertId)
	require.NoError(t, err)
	require.Nil(t, got)

	// verify history was written once to both tables (row appends; head upserts by id)
	hc, rc := countHistory(t, &cfg)
	require.Equal(t, 1, rc)
	require.Equal(t, 1, hc)
}

func TestPutStatusClean_NotFound(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newTestDBConfig(t)
	_ = setupClickhouseMock(t, &cfg)
	config = cfg

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Params = gin.Params{gin.Param{Key: "name", Value: "ruleX"}, gin.Param{Key: "id", Value: "NOPE"}}
	claims := &jwt.JWTClaims{Subject: "tester", Extra: map[string]any{"attrs": map[string]any{}}}
	c.Set(common.ContextKeyJWTClaims, claims)

	PutStatusClean(c)
	require.Equal(t, http.StatusNotFound, rec.Code)
}

func TestPutStatusMask_TogglesAndWritesHistory(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newTestDBConfig(t)
	config = cfg

	// seed alert
	now := time.Now().UTC()
	ct := now
	a := alert_common.Value{
		Id:        "id-2",
		Timestamp: now, UpdatedAt: now, CheckTime: &ct,
		Name: "ruleB", AlertType: string(alert_common.TypeQuery), AlertId: "AL-2",
		Value: 1, Severity: "Minor", Status: alert_common.StatusAlerting,
	}
	seedAlert(t, &cfg, a)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	// JSON body로 mask 값 전송
	requestBody := `{"mask": true}`
	c.Request = httptest.NewRequest(http.MethodPut, "/status/ruleB/AL-2/mask", strings.NewReader(requestBody))
	c.Request.Header.Set("Content-Type", "application/json")
	c.Params = gin.Params{gin.Param{Key: "name", Value: a.Name}, gin.Param{Key: "id", Value: a.AlertId}}
	claims := &jwt.JWTClaims{Subject: "tester"}
	c.Set(common.ContextKeyJWTClaims, claims)

	PutStatusMask(c)
	require.Equal(t, http.StatusOK, rec.Code)

	// verify updated mask persisted
	m := &model.Model{Config: &cfg}
	got, err := m.GetAlert(a.Name, a.AlertId)
	require.NoError(t, err)
	require.NotNil(t, got)
	require.True(t, got.Mask)

	// history wrote one to both tables (row appends; head upserts by id)
	hc2, rc2 := countHistory(t, &cfg)
	require.Equal(t, 1, rc2)
	require.Equal(t, 1, hc2)
}

// countHistory returns counts in history_alert and history_alert_row in the statistics DB
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

// seedHistoryData inserts test history data into the statistics database
func seedHistoryData(t *testing.T, cfg *common.Config) {
	t.Helper()

	err := orm.StatisticsHandler(orm.DriverDefault, &cfg.Statistics.Database, func(db *sqlx.DB) error {
		now := time.Now().UTC()

		// Insert test history records
		records := []struct {
			timestamp    time.Time
			id           string
			alertId      string
			name         string
			alertType    string
			description  string
			value        float64
			severity     string
			prevSeverity string
			prevValue    *float64
			labels       string
		}{
			{
				timestamp:    now.Add(-time.Hour),
				id:           "hist-1",
				alertId:      "AL-1",
				name:         "test-rule",
				alertType:    "query",
				description:  "Test alert 1",
				value:        10.5,
				severity:     "Major",
				prevSeverity: "Normal",
				prevValue:    nil,
				labels:       `{"service":"api","env":"prod"}`,
			},
			{
				timestamp:    now.Add(-30 * time.Minute),
				id:           "hist-2",
				alertId:      "AL-2",
				name:         "test-rule",
				alertType:    "query",
				description:  "Test alert 2",
				value:        5.0,
				severity:     "Minor",
				prevSeverity: "Normal",
				prevValue:    func() *float64 { v := 2.0; return &v }(),
				labels:       `{"service":"db","env":"prod"}`,
			},
			{
				timestamp:    now.Add(-15 * time.Minute),
				id:           "hist-3",
				alertId:      "AL-3",
				name:         "other-rule",
				alertType:    "metric",
				description:  "Different rule alert",
				value:        100.0,
				severity:     "Critical",
				prevSeverity: "Major",
				prevValue:    func() *float64 { v := 90.0; return &v }(),
				labels:       `{"service":"web","env":"staging"}`,
			},
		}

		for _, r := range records {
			_, err := db.Exec(`
				INSERT INTO history_alert 
				(timestamp, id, alert_id, name, alert_type, description, value, severity, previous_severity, previous_value, labels, status, previous_timestamp, version)
				VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			`, r.timestamp, r.id, r.alertId, r.name, r.alertType, r.description, r.value, r.severity, r.prevSeverity, r.prevValue, r.labels, "alerting", nil, r.timestamp)
			if err != nil {
				return err
			}
		}
		return nil
	})

	require.NoError(t, err)
}

// GetHistQuery는 테스트용 wrapper 함수입니다
func GetHistQuery(c *gin.Context) {
	handler := GetHistQueryHandler(config)
	handler(c)
}

func TestGetHistQueryHandler_Success_AllRecords(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newTestDBConfig(t)
	config = cfg

	// Seed test data
	seedHistoryData(t, &cfg)

	// Test: Get all records (no filters)
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/hist", nil)

	GetHistQuery(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var result []alert_types.AlertValue
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	require.Len(t, result, 3) // All 3 records

	// Check order (should be descending by timestamp)
	require.Equal(t, "hist-3", result[0].ID) // Most recent
	require.Equal(t, "hist-2", result[1].ID)
	require.Equal(t, "hist-1", result[2].ID) // Oldest
}

func TestGetHistQueryHandler_Success_FilterByName(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newTestDBConfig(t)
	config = cfg

	seedHistoryData(t, &cfg)

	// Test: Filter by name
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/hist?name=test-rule", nil)

	GetHistQuery(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var result []alert_types.AlertValue
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	require.Len(t, result, 2) // Only test-rule records

	for _, r := range result {
		require.Equal(t, "test-rule", r.Name)
	}
}

func TestGetHistQueryHandler_Success_FilterByTimeRange(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newTestDBConfig(t)
	config = cfg

	seedHistoryData(t, &cfg)

	// Test: Filter by time range (last 45 minutes)
	now := time.Now().UTC()
	startTime := now.Add(-45 * time.Minute).Format(time.DateTime)
	endTime := now.Format(time.DateTime)

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)

	// Create request with proper URL encoding
	req := httptest.NewRequest(http.MethodGet, "/hist", nil)
	q := req.URL.Query()
	q.Add("start-time", startTime)
	q.Add("end-time", endTime)
	req.URL.RawQuery = q.Encode()
	c.Request = req

	GetHistQuery(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var result []alert_types.AlertValue
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	require.Len(t, result, 2) // Records within 45 minutes
}

func TestGetHistQueryHandler_Success_WithCount(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newTestDBConfig(t)
	config = cfg

	seedHistoryData(t, &cfg)

	// Test: Limit count to 1
	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/hist?count=1", nil)

	GetHistQuery(c)

	require.Equal(t, http.StatusOK, rec.Code)

	var result []alert_types.AlertValue
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &result))
	require.Len(t, result, 1)                // Only 1 record
	require.Equal(t, "hist-3", result[0].ID) // Most recent
}

func TestGetHistQueryHandler_ValidationErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	cfg := newTestDBConfig(t)
	config = cfg

	tests := []struct {
		name           string
		setupRequest   func() *http.Request
		expectedStatus int
	}{
		{
			name: "Invalid name - too long",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/hist", nil)
				q := req.URL.Query()
				q.Add("name", strings.Repeat("a", 256))
				req.URL.RawQuery = q.Encode()
				return req
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid time format",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/hist", nil)
				q := req.URL.Query()
				q.Add("start-time", "invalid-time")
				req.URL.RawQuery = q.Encode()
				return req
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid count format",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/hist", nil)
				q := req.URL.Query()
				q.Add("count", "invalid")
				req.URL.RawQuery = q.Encode()
				return req
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Count too large",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/hist", nil)
				q := req.URL.Query()
				q.Add("count", "20000")
				req.URL.RawQuery = q.Encode()
				return req
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Start time after end time",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/hist", nil)
				q := req.URL.Query()
				q.Add("start-time", "2024-01-02 00:00:00")
				q.Add("end-time", "2024-01-01 00:00:00")
				req.URL.RawQuery = q.Encode()
				return req
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Time range too long (over 1 year)",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/hist", nil)
				q := req.URL.Query()
				q.Add("start-time", "2023-01-01 00:00:00")
				q.Add("end-time", "2025-01-01 00:00:00")
				req.URL.RawQuery = q.Encode()
				return req
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rec := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(rec)
			c.Request = tt.setupRequest()

			GetHistQuery(c)

			require.Equal(t, tt.expectedStatus, rec.Code)
		})
	}
}

func TestGetHistQueryHandler_DatabaseNotConfigured(t *testing.T) {
	gin.SetMode(gin.TestMode)

	// Use config without statistics database
	emptyConfig := common.Config{}
	originalConfig := config
	config = emptyConfig
	t.Cleanup(func() { config = originalConfig })

	rec := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(rec)
	c.Request = httptest.NewRequest(http.MethodGet, "/hist", nil)

	GetHistQuery(c)

	require.Equal(t, http.StatusInternalServerError, rec.Code)

	var errorResp map[string]string
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &errorResp))
	require.Contains(t, errorResp["message"], "statistics database not configured")
}

func TestValidateHistoryQueryParams_Success(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name          string
		setupRequest  func() *http.Request
		expectedName  string
		expectedCount int
	}{
		{
			name: "Empty params",
			setupRequest: func() *http.Request {
				return httptest.NewRequest(http.MethodGet, "/hist", nil)
			},
			expectedName:  "",
			expectedCount: 0,
		},
		{
			name: "With name only",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/hist", nil)
				q := req.URL.Query()
				q.Add("name", "test-rule")
				req.URL.RawQuery = q.Encode()
				return req
			},
			expectedName:  "test-rule",
			expectedCount: 0,
		},
		{
			name: "With count",
			setupRequest: func() *http.Request {
				req := httptest.NewRequest(http.MethodGet, "/hist", nil)
				q := req.URL.Query()
				q.Add("count", "50")
				req.URL.RawQuery = q.Encode()
				return req
			},
			expectedName:  "",
			expectedCount: 50,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c, _ := gin.CreateTestContext(httptest.NewRecorder())
			c.Request = tt.setupRequest()

			params, err := validateHistoryQueryParams(c)

			require.NoError(t, err)
			require.Equal(t, tt.expectedName, params.Name)
			require.Equal(t, tt.expectedCount, params.Count)
		})
	}
}

func TestQueryHistoryFromDB_EmptyDatabase(t *testing.T) {
	cfg := newTestDBConfig(t)

	params := &HistoryQueryParams{
		Name:  "nonexistent",
		Count: 10,
	}

	values, err := queryHistoryFromDB(&cfg.Statistics.Database, params)

	require.NoError(t, err)
	require.Empty(t, values)
}
