package store

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"
	"strings"
	"time"

	"github.com/jmoiron/sqlx"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/third_party/GoVisual/internal/model"
)

type HistoryRequest struct {
	Id              string `json:"id"`
	Timestamp       string `json:"timestamp"`
	User            string `json:"user"`
	Method          string `json:"method"`
	Path            string `json:"path"`
	Query           string `json:"query"`
	RequestHeaders  string `json:"request_headers"`
	ResponseHeaders string `json:"response_headers"`
	StatusCode      string `json:"status_code"`
	Duration        string `json:"duration"`
	RequestBody     string `json:"request_body"`
	ResponseBody    string `json:"response_body"`
	Error           string `json:"error"`
	MiddlewareTrace string `json:"middleware_trace"`
	RouteTrace      string `json:"route_trace"`
}

// ClickHouse implements the Store interface with ClickHouse as backend
type ClickHouseStore struct {
	config    *common.Config
	tableName string
	capacity  int
}

// isValidCHTableName checks whether the given identifier is a safe ClickHouse table name
// allowing alphanumerics and underscores, optionally schema-qualified as schema.table.
func isValidCHTableName(name string) bool {
	re := regexp.MustCompile(`^[a-zA-Z0-9_]+(\.[a-zA-Z0-9_]+)?$`)
	return re.MatchString(name)
}

// quoteCHQualifiedIdentifier safely quotes an optional schema-qualified identifier using backticks.
// It assumes the identifier has already been validated by isValidCHTableName.
func quoteCHQualifiedIdentifier(name string) string {
	parts := strings.Split(name, ".")
	for i, p := range parts {
		// escape backticks by doubling them
		p = strings.ReplaceAll(p, "`", "``")
		parts[i] = "`" + p + "`"
	}
	return strings.Join(parts, ".")
}

// NewClickHouseStore creates a new ClickHouse-backed store
func NewClickHouseStore(config *common.Config, tableName string, capacity int) (*ClickHouseStore, error) {
	if capacity <= 0 {
		capacity = 100
	}

	if tableName == "" {
		tableName = "history_requests"
	}
	// Validate table name to avoid SQL injection via identifier interpolation
	if !isValidCHTableName(tableName) {
		return nil, fmt.Errorf("invalid table name: must match [a-zA-Z0-9_]+ or schema.table with same rules")
	}

	store := &ClickHouseStore{
		config:    config,
		tableName: tableName,
		capacity:  capacity,
	}

	return store, nil
}

// Add adds a new request log to the store
func (s *ClickHouseStore) Add(reqLog *model.RequestLog) {
	// Extract user from JWT subject or client_id from Basic Auth
	subject, err := getJwtSubject(reqLog.RequestHeaders)
	if err != nil {
		slog.Warn("failed to extract subject from authorization header",
			"error", err,
			"method", reqLog.Method,
			"path", reqLog.Path)
		subject = ""
	} else if subject != "" {
		slog.Debug("extracted subject from authorization header",
			"subject", subject,
			"method", reqLog.Method,
			"path", reqLog.Path)
	}

	statsDB := &s.config.Statistics.Database
	if statsDB == nil || statsDB.Driver == "" {
		return
	}

	quotedTbl := quoteCHQualifiedIdentifier(s.tableName)
	// #nosec G201 -- injecting a validated, safely quoted identifier; values are parameterized.
	insertSQL := fmt.Sprintf("INSERT INTO %s (id, timestamp, user, method, path, query, request_headers, response_headers, status_code, duration, request_body, response_body, error, middleware_trace, route_trace) VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)", quotedTbl)
	_ = orm.StatisticsHandler("", statsDB, func(db *sqlx.DB) error {
		q := db.Rebind(insertSQL)
		_, err := db.Exec(
			q,
			reqLog.ID,
			reqLog.Timestamp,
			subject,
			reqLog.Method,
			reqLog.Path,
			reqLog.Query,
			prepareJSON(reqLog.RequestHeaders),
			prepareJSON(reqLog.ResponseHeaders),
			reqLog.StatusCode,
			reqLog.Duration,
			reqLog.RequestBody,
			reqLog.ResponseBody,
			reqLog.Error,
			prepareJSON(reqLog.MiddlewareTrace),
			prepareJSON(reqLog.RouteTrace),
		)
		if err != nil {
			slog.Error("clickhouse insert failed", "error", err)
			return err
		}
		return nil
	})

	// Clean up old logs (no-op for ClickHouse; TTL can be managed by table settings)
	s.cleanup()
}

// cleanup removes old logs to maintain the capacity limit
func (s *ClickHouseStore) cleanup() {
	/*
			connector := clickhouse_interface.GetConnector(*s.config)

			countSQL := fmt.Sprintf("SELECT toString(count()) AS cnt FROM %s", s.tableName)
			vals := url.Values{"query": {countSQL}}
			resp, err := connector.GetClickhouseQuery(context.Background(), vals)
			if err != nil {
				log.Printf("cleanup: count query failed: %v", err)
				return
			}

			var cntWrapper struct {
				Data []struct {
					Cnt string `json:"cnt"`
				} `json:"data"`
			}
			if err := json.Unmarshal(resp.Body(), &cntWrapper); err != nil {
				log.Printf("cleanup: unmarshal count response failed: %v", err)
				return
			}
			if len(cntWrapper.Data) == 0 {
				return
			}
			totalCount, err := strconv.Atoi(cntWrapper.Data[0].Cnt)
			if err != nil {
				log.Printf("cleanup: parse count failed: %v", err)
				return
			}

			excess := totalCount - s.capacity
			if excess <= 0 {
				return
			}

			deleteSQL := fmt.Sprintf(`
		        ALTER TABLE %s
		        DELETE WHERE Id IN (
		            SELECT Id FROM %s
		            ORDER BY Timestamp ASC
		            LIMIT %d
		        )
		    `, s.tableName, s.tableName, excess)

			vals = url.Values{"query": {deleteSQL}}
			if _, err := connector.PostClickhouseQuery(context.Background(), []byte{}, vals); err != nil {
				log.Printf("cleanup: delete mutation failed: %v", err)
				return
			}

			log.Printf("cleanup: removed %d old rows from %s", excess, s.tableName)
	*/
}

// Get retrieves a specific request log by its ID using the statistics database connection
func (s *ClickHouseStore) Get(id string) (*model.RequestLog, bool) {
	statsDB := &s.config.Statistics.Database
	if statsDB == nil || statsDB.Driver == "" {
		return nil, false
	}

	type row struct {
		ID              string    `db:"id"`
		Timestamp       time.Time `db:"timestamp"`
		Method          string    `db:"method"`
		Path            string    `db:"path"`
		Query           string    `db:"query"`
		RequestHeaders  string    `db:"request_headers"`
		ResponseHeaders string    `db:"response_headers"`
		StatusCode      int       `db:"status_code"`
		Duration        int64     `db:"duration"`
		RequestBody     string    `db:"request_body"`
		ResponseBody    string    `db:"response_body"`
		Error           string    `db:"error"`
		MiddlewareTrace string    `db:"middleware_trace"`
		RouteTrace      string    `db:"route_trace"`
	}

	var r row
	err := orm.StatisticsHandler("", statsDB, func(db *sqlx.DB) error {
		quotedTbl := quoteCHQualifiedIdentifier(s.tableName)
		q := db.Rebind(fmt.Sprintf("SELECT id, timestamp, method, path, query, request_headers, response_headers, status_code, duration, request_body, response_body, error, middleware_trace, route_trace FROM %s WHERE id=?", quotedTbl))
		return db.Get(&r, q, id)
	})
	if err != nil {
		return nil, false
	}

	reqLog := &model.RequestLog{
		ID:           r.ID,
		Timestamp:    r.Timestamp,
		Method:       r.Method,
		Path:         r.Path,
		Query:        r.Query,
		StatusCode:   r.StatusCode,
		Duration:     r.Duration,
		RequestBody:  r.RequestBody,
		ResponseBody: r.ResponseBody,
		Error:        r.Error,
	}
	if err := json.Unmarshal([]byte(r.RequestHeaders), &reqLog.RequestHeaders); err != nil {
		slog.Warn("Failed to unmarshal request headers JSON", "id", r.ID, "error", err)
	}
	if err := json.Unmarshal([]byte(r.ResponseHeaders), &reqLog.ResponseHeaders); err != nil {
		slog.Warn("Failed to unmarshal response headers JSON", "id", r.ID, "error", err)
	}
	if err := json.Unmarshal([]byte(r.MiddlewareTrace), &reqLog.MiddlewareTrace); err != nil {
		slog.Warn("Failed to unmarshal middleware trace JSON", "id", r.ID, "error", err)
	}
	if err := json.Unmarshal([]byte(r.RouteTrace), &reqLog.RouteTrace); err != nil {
		slog.Warn("Failed to unmarshal route trace JSON", "id", r.ID, "error", err)
	}

	return reqLog, true
}

// GetAll fetches recent logs via the statistics database (max 100 rows)
func (s *ClickHouseStore) GetAll() []*model.RequestLog {
	statsDB := &s.config.Statistics.Database
	if statsDB == nil || statsDB.Driver == "" {
		return nil
	}

	type row struct {
		ID              string    `db:"id"`
		Timestamp       time.Time `db:"timestamp"`
		Method          string    `db:"method"`
		Path            string    `db:"path"`
		Query           string    `db:"query"`
		RequestHeaders  string    `db:"request_headers"`
		ResponseHeaders string    `db:"response_headers"`
		StatusCode      int       `db:"status_code"`
		Duration        int64     `db:"duration"`
		RequestBody     string    `db:"request_body"`
		ResponseBody    string    `db:"response_body"`
		Error           string    `db:"error"`
		MiddlewareTrace string    `db:"middleware_trace"`
		RouteTrace      string    `db:"route_trace"`
	}

	rows := []row{}
	_ = orm.StatisticsHandler("", statsDB, func(db *sqlx.DB) error {
		quotedTbl := quoteCHQualifiedIdentifier(s.tableName)
		q := db.Rebind(fmt.Sprintf("SELECT id, timestamp, method, path, query, request_headers, response_headers, status_code, duration, request_body, response_body, error, middleware_trace, route_trace FROM %s ORDER BY timestamp DESC LIMIT 100", quotedTbl))
		return db.Select(&rows, q)
	})

	result := make([]*model.RequestLog, 0, len(rows))
	for _, r := range rows {
		rl := &model.RequestLog{
			ID:           r.ID,
			Timestamp:    r.Timestamp,
			Method:       r.Method,
			Path:         r.Path,
			Query:        r.Query,
			StatusCode:   r.StatusCode,
			Duration:     r.Duration,
			RequestBody:  r.RequestBody,
			ResponseBody: r.ResponseBody,
			Error:        r.Error,
		}
		_ = json.Unmarshal([]byte(r.RequestHeaders), &rl.RequestHeaders)
		_ = json.Unmarshal([]byte(r.ResponseHeaders), &rl.ResponseHeaders)
		_ = json.Unmarshal([]byte(r.MiddlewareTrace), &rl.MiddlewareTrace)
		_ = json.Unmarshal([]byte(r.RouteTrace), &rl.RouteTrace)
		result = append(result, rl)
	}
	return result
}

// GetLatest returns the n most recent request logs using the statistics database
func (s *ClickHouseStore) GetLatest(n int) []*model.RequestLog {
	statsDB := &s.config.Statistics.Database
	if statsDB == nil || statsDB.Driver == "" {
		return nil
	}

	type row struct {
		ID              string    `db:"id"`
		Timestamp       time.Time `db:"timestamp"`
		Method          string    `db:"method"`
		Path            string    `db:"path"`
		Query           string    `db:"query"`
		RequestHeaders  string    `db:"request_headers"`
		ResponseHeaders string    `db:"response_headers"`
		StatusCode      int       `db:"status_code"`
		Duration        int64     `db:"duration"`
		RequestBody     string    `db:"request_body"`
		ResponseBody    string    `db:"response_body"`
		Error           string    `db:"error"`
		MiddlewareTrace string    `db:"middleware_trace"`
		RouteTrace      string    `db:"route_trace"`
	}

	rows := []row{}
	_ = orm.StatisticsHandler("", statsDB, func(db *sqlx.DB) error {
		quotedTbl := quoteCHQualifiedIdentifier(s.tableName)
		q := db.Rebind(fmt.Sprintf("SELECT id, timestamp, method, path, query, request_headers, response_headers, status_code, duration, request_body, response_body, error, middleware_trace, route_trace FROM %s ORDER BY timestamp DESC LIMIT ?", quotedTbl))
		return db.Select(&rows, q, n)
	})

	result := make([]*model.RequestLog, 0, len(rows))
	for _, r := range rows {
		rl := &model.RequestLog{
			ID:           r.ID,
			Timestamp:    r.Timestamp,
			Method:       r.Method,
			Path:         r.Path,
			Query:        r.Query,
			StatusCode:   r.StatusCode,
			Duration:     r.Duration,
			RequestBody:  r.RequestBody,
			ResponseBody: r.ResponseBody,
			Error:        r.Error,
		}
		_ = json.Unmarshal([]byte(r.RequestHeaders), &rl.RequestHeaders)
		_ = json.Unmarshal([]byte(r.ResponseHeaders), &rl.ResponseHeaders)
		_ = json.Unmarshal([]byte(r.MiddlewareTrace), &rl.MiddlewareTrace)
		_ = json.Unmarshal([]byte(r.RouteTrace), &rl.RouteTrace)
		result = append(result, rl)
	}
	return result
}

// Clear clears all stored request logs using the statistics database
func (s *ClickHouseStore) Clear() error {
	statsDB := &s.config.Statistics.Database
	if statsDB == nil || statsDB.Driver == "" {
		return nil
	}
	return orm.StatisticsHandler("", statsDB, func(db *sqlx.DB) error {
		quotedTbl := quoteCHQualifiedIdentifier(s.tableName)
		q := fmt.Sprintf("TRUNCATE TABLE %s", quotedTbl)
		_, err := db.Exec(q)
		return err
	})
}

// Close closes the database connection
func (s *ClickHouseStore) Close() error {
	return nil
}
