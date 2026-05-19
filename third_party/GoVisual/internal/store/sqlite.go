package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	"ntels.com/pharos/third_party/GoVisual/internal/model"
	// Don't import and register SQLite driver automatically
	// _ "github.com/ncruces/go-sqlite3/driver"
	// _ "github.com/ncruces/go-sqlite3/embed"
)

// SQLiteStore implements the Store interface with SQLite as backend
type SQLiteStore struct {
	db        *sql.DB
	tableName string
	capacity  int
	// Add a flag to track if we own the connection
	ownsConnection bool
}

// isValidTableName checks if a table name contains only alphanumeric and underscore characters
func isValidTableName(tableName string) bool {
	match, _ := regexp.MatchString(`^[a-zA-Z0-9_]+$`, tableName)
	return match
}

// quoteSQLiteIdent safely quotes a SQLite identifier using double quotes
func quoteSQLiteIdent(name string) string {
	// escape embedded quotes by doubling them
	escaped := strings.ReplaceAll(name, "\"", "\"\"")
	return "\"" + escaped + "\""
}

// NewSQLiteStore creates a new SQLite-backed store
func NewSQLiteStore(dbPath, tableName string, capacity int) (*SQLiteStore, error) {
	if capacity <= 0 {
		capacity = 100
	}

	// Validate table name to prevent SQL injection
	if !isValidTableName(tableName) {
		return nil, fmt.Errorf("invalid table name: table name can only contain letters, numbers, and underscores")
	}

	// Connect to the database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open SQLite DB: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite DB: %w", err)
	}

	store := &SQLiteStore{
		db:             db,
		tableName:      tableName,
		capacity:       capacity,
		ownsConnection: true,
	}

	// Create the table if it doesn't exist
	if err := store.createTable(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return store, nil
}

// NewSQLiteStoreWithDB creates a new SQLite store with an existing database connection
func NewSQLiteStoreWithDB(db *sql.DB, tableName string, capacity int) (*SQLiteStore, error) {
	if db == nil {
		return nil, fmt.Errorf("database connection cannot be nil")
	}

	if capacity <= 0 {
		capacity = 100
	}

	// Validate table name to prevent SQL injection
	if !isValidTableName(tableName) {
		return nil, fmt.Errorf("invalid table name: table name can only contain letters, numbers, and underscores")
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping SQLite DB: %w", err)
	}

	store := &SQLiteStore{
		db:             db,
		tableName:      tableName,
		capacity:       capacity,
		ownsConnection: false,
	}

	// Create the table if it doesn't exist
	if err := store.createTable(); err != nil {
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return store, nil
}

// createTable creates the required table if it doesn't exist
func (s *SQLiteStore) createTable() error {
	tbl := quoteSQLiteIdent(s.tableName)
	// #nosec G201 -- using Sprintf to inject a validated, quoted table name. Table name is strictly validated by isValidTableName and quoted via quoteSQLiteIdent to prevent injection; parameters cannot be used for identifiers.
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			timestamp DATETIME,
			user TEXT,
			method TEXT,
			path TEXT,
			query TEXT,
			request_headers TEXT,
			response_headers TEXT,
			status_code INTEGER,
			duration INTEGER,
			request_body TEXT,
			response_body TEXT,
			error TEXT,
			middleware_trace TEXT,
			route_trace TEXT,
			created_at DATETIME DEFAULT CURRENT_TIMESTAMP
		)
		`, tbl)

	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}

	// Create index on timestamp for faster retrieval
	indexName := fmt.Sprintf("%s_timestamp_idx", s.tableName)
	// #nosec G201 -- injecting validated, quoted identifier names (index and table). No user input; identifiers can’t be parameterized.
	indexQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s(timestamp DESC)", quoteSQLiteIdent(indexName), tbl)
	_, err = s.db.Exec(indexQuery)

	return err
}

// Add adds a new request log to the store
func (s *SQLiteStore) Add(reqLog *model.RequestLog) {
	reqHeaders := prepareJSON(reqLog.RequestHeaders)
	respHeaders := prepareJSON(reqLog.ResponseHeaders)

	middlewareTrace := "[]"
	if len(reqLog.MiddlewareTrace) > 0 {
		if data, err := json.Marshal(reqLog.MiddlewareTrace); err == nil {
			middlewareTrace = string(data)
		}
	}

	routeTrace := "{}"
	if reqLog.RouteTrace != nil {
		if data, err := json.Marshal(reqLog.RouteTrace); err == nil {
			routeTrace = string(data)
		}
	}

	tbl := quoteSQLiteIdent(s.tableName)
	// #nosec G201 -- injecting validated, quoted table name only; all data values are parameterized with placeholders.
	query := fmt.Sprintf(`
		INSERT OR REPLACE INTO %s (
			id, timestamp, user, method, path, query, request_headers, response_headers,
			status_code, duration, request_body, response_body, error, 
			middleware_trace, route_trace
		) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, tbl)

	subject, _ := getJwtSubject(reqLog.RequestHeaders)
	_, err := s.db.Exec(
		query,
		reqLog.ID,
		reqLog.Timestamp,
		subject,
		reqLog.Method,
		reqLog.Path,
		reqLog.Query,
		reqHeaders,
		respHeaders,
		reqLog.StatusCode,
		reqLog.Duration,
		reqLog.RequestBody,
		reqLog.ResponseBody,
		reqLog.Error,
		middlewareTrace,
		routeTrace,
	)

	if err != nil {
		log.Printf("Failed to store request log in SQLite: %v", err)
	}

	s.cleanup()
}

// cleanup removes old logs to maintain the capacity limit
func (s *SQLiteStore) cleanup() {
	tbl := quoteSQLiteIdent(s.tableName)
	// #nosec G201 -- injecting validated, quoted table name. No user-controlled input; cannot parameterize identifiers.
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM %s", tbl)
	var count int
	err := s.db.QueryRow(countQuery).Scan(&count)
	if err != nil {
		log.Printf("Failed to count logs: %v", err)
		return
	}

	if count <= s.capacity {
		return
	}

	// Delete oldest logs
	// #nosec G201 -- injecting validated, quoted table name (twice). LIMIT remains parameterized; identifiers cannot be parameterized but inputs are validated and quoted.
	deleteQuery := fmt.Sprintf(`
		DELETE FROM %s
		WHERE id IN (
			SELECT id FROM %s
			ORDER BY created_at ASC, timestamp ASC
			LIMIT ?
		)
	`, tbl, tbl)

	_, err = s.db.Exec(deleteQuery, count-s.capacity)
	if err != nil {
		log.Printf("Failed to clean up old logs: %v", err)
	}
}

// Get retrieves a specific request log by its ID
func (s *SQLiteStore) Get(id string) (*model.RequestLog, bool) {
	tbl := quoteSQLiteIdent(s.tableName)
	// #nosec G201 -- injecting validated, quoted table name into a static query; the id is parameterized with a placeholder.
	query := fmt.Sprintf(`
		SELECT 
			id, timestamp, method, path, query, 
			COALESCE(request_headers, '{}'),
			COALESCE(response_headers, '{}'),
			status_code, duration, request_body, response_body, error, 
			COALESCE(middleware_trace, '[]'),
			COALESCE(route_trace, '{}')
		FROM %s
		WHERE id = ?
	`, tbl)

	var (
		reqLog          model.RequestLog
		reqHeadersStr   string
		respHeadersStr  string
		middlewareTrace string
		routeTrace      string
	)

	err := s.db.QueryRow(query, id).Scan(
		&reqLog.ID,
		&reqLog.Timestamp,
		&reqLog.Method,
		&reqLog.Path,
		&reqLog.Query,
		&reqHeadersStr,
		&respHeadersStr,
		&reqLog.StatusCode,
		&reqLog.Duration,
		&reqLog.RequestBody,
		&reqLog.ResponseBody,
		&reqLog.Error,
		&middlewareTrace,
		&routeTrace,
	)

	if err != nil {
		if err == sql.ErrNoRows {
			return nil, false
		}
		log.Printf("Failed to get request log from SQLite: %v", err)
		return nil, false
	}

	_ = json.Unmarshal([]byte(reqHeadersStr), &reqLog.RequestHeaders)
	_ = json.Unmarshal([]byte(respHeadersStr), &reqLog.ResponseHeaders)
	_ = json.Unmarshal([]byte(middlewareTrace), &reqLog.MiddlewareTrace)
	_ = json.Unmarshal([]byte(routeTrace), &reqLog.RouteTrace)

	return &reqLog, true
}

// GetAll returns all stored request logs
func (s *SQLiteStore) GetAll() []*model.RequestLog {
	tbl := quoteSQLiteIdent(s.tableName)
	// #nosec G201 -- injecting validated, quoted table name; entire query uses static columns and order.
	query := fmt.Sprintf(`
		SELECT 
			id, timestamp, method, path, query, 
			COALESCE(request_headers, '{}'),
			COALESCE(response_headers, '{}'),
			status_code, duration, request_body, response_body, error, 
			COALESCE(middleware_trace, '[]'),
			COALESCE(route_trace, '{}')
		FROM %s
		ORDER BY timestamp DESC
	`, tbl)

	return s.queryLogs(query)
}

// GetLatest returns the n most recent request logs
func (s *SQLiteStore) GetLatest(n int) []*model.RequestLog {
	tbl := quoteSQLiteIdent(s.tableName)
	// #nosec G201 -- injecting validated, quoted table name; LIMIT is parameterized. Identifiers cannot be bound.
	query := fmt.Sprintf(`
		SELECT 
			id, timestamp, method, path, query, 
			COALESCE(request_headers, '{}'),
			COALESCE(response_headers, '{}'),
			status_code, duration, request_body, response_body, error, 
			COALESCE(middleware_trace, '[]'),
			COALESCE(route_trace, '{}')
		FROM %s
		ORDER BY timestamp DESC
		LIMIT ?
	`, tbl)

	return s.queryLogs(query, n)
}

// queryLogs executes a query and returns the resulting log entries
func (s *SQLiteStore) queryLogs(query string, args ...any) []*model.RequestLog {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("Failed to query logs from SQLite: %v", err)
		return nil
	}
	defer func() { _ = rows.Close() }()

	var logs []*model.RequestLog

	for rows.Next() {
		var (
			reqLog          model.RequestLog
			reqHeadersStr   string
			respHeadersStr  string
			middlewareTrace string
			routeTrace      string
		)

		err := rows.Scan(
			&reqLog.ID,
			&reqLog.Timestamp,
			&reqLog.Method,
			&reqLog.Path,
			&reqLog.Query,
			&reqHeadersStr,
			&respHeadersStr,
			&reqLog.StatusCode,
			&reqLog.Duration,
			&reqLog.RequestBody,
			&reqLog.ResponseBody,
			&reqLog.Error,
			&middlewareTrace,
			&routeTrace,
		)

		if err != nil {
			log.Printf("Failed to scan row: %v", err)
			continue
		}

		_ = json.Unmarshal([]byte(reqHeadersStr), &reqLog.RequestHeaders)
		_ = json.Unmarshal([]byte(respHeadersStr), &reqLog.ResponseHeaders)
		_ = json.Unmarshal([]byte(middlewareTrace), &reqLog.MiddlewareTrace)
		_ = json.Unmarshal([]byte(routeTrace), &reqLog.RouteTrace)

		logs = append(logs, &reqLog)
	}

	if err := rows.Err(); err != nil {
		log.Printf("Error iterating over rows: %v", err)
	}

	return logs
}

// Clear removes all logs from the store
func (s *SQLiteStore) Clear() error {
	tbl := quoteSQLiteIdent(s.tableName)
	// #nosec G201 -- injecting validated, quoted table name in a static DELETE statement.
	query := fmt.Sprintf("DELETE FROM %s", tbl)
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to clear logs: %w", err)
	}

	return nil
}

// Close closes the database connection
func (s *SQLiteStore) Close() error {
	// Only close the connection if we own it
	if s.ownsConnection {
		return s.db.Close()
	}
	return nil
}
