package store

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"regexp"
	"strings"

	_ "github.com/lib/pq"
	"ntels.com/pharos/third_party/GoVisual/internal/model"
)

// PostgresStore implements the Store interface with PostgreSQL as backend
type PostgresStore struct {
	db        *sql.DB
	tableName string
	capacity  int
}

// NewPostgresStore creates a new PostgreSQL-backed store
func NewPostgresStore(connStr, tableName string, capacity int) (*PostgresStore, error) {
	if capacity <= 0 {
		capacity = 100
	}

	// Validate table name (allow optional schema prefix) to prevent SQL injection via identifiers
	if !isValidPGTableName(tableName) {
		return nil, fmt.Errorf("invalid table name: must match [a-zA-Z0-9_]+ or schema.table with same rules")
	}

	// Connect to the database
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to PostgreSQL: %w", err)
	}

	// Test the connection
	if err := db.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping PostgreSQL: %w", err)
	}

	store := &PostgresStore{
		db:        db,
		tableName: tableName,
		capacity:  capacity,
	}

	// Create the table if it doesn't exist
	if err := store.createTable(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("failed to create table: %w", err)
	}

	return store, nil
}

// createTable creates the required table if it doesn't exist
// isValidPGTableName checks for safe postgres identifiers (optionally schema qualified)
func isValidPGTableName(name string) bool {
	// allow: table, schema.table (letters, digits, underscore only)
	re := regexp.MustCompile(`^[a-zA-Z0-9_]+(\.[a-zA-Z0-9_]+)?$`)
	return re.MatchString(name)
}

// quotePGQualifiedIdentifier safely quotes an optional schema-qualified identifier: schema.table
// It assumes the identifier has already been validated by isValidPGTableName.
func quotePGQualifiedIdentifier(name string) string {
	parts := strings.Split(name, ".")
	for i, p := range parts {
		// double-quote each identifier part and escape embedded quotes by doubling them
		p = strings.ReplaceAll(p, "\"", "\"\"")
		parts[i] = "\"" + p + "\""
	}
	return strings.Join(parts, ".")
}

func (s *PostgresStore) createTable() error {
	tbl := quotePGQualifiedIdentifier(s.tableName)
	query := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS %s (
			id TEXT PRIMARY KEY,
			timestamp TIMESTAMP WITH TIME ZONE,
			user TEXT,
			method TEXT,
			path TEXT,
			query TEXT,
			request_headers TEXT,
			response_headers TEXT,
			status_code INTEGER,
			duration BIGINT,
			request_body TEXT,
			response_body TEXT,
			error TEXT,
			middleware_trace TEXT,
			route_trace TEXT,
			created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
		)
		`, tbl)

	_, err := s.db.Exec(query)
	if err != nil {
		return err
	}

	// Create index on timestamp for faster retrieval
	indexName := fmt.Sprintf("%s_timestamp_idx", s.tableName)
	indexQuery := fmt.Sprintf("CREATE INDEX IF NOT EXISTS %s ON %s (timestamp DESC)", quotePGQualifiedIdentifier(indexName), tbl)
	_, err = s.db.Exec(indexQuery)

	return err
}

// Add adds a new request log to the store
func (s *PostgresStore) Add(reqLog *model.RequestLog) {
	// Prepare all JSON fields properly
	reqHeaders := prepareJSON(reqLog.RequestHeaders)
	respHeaders := prepareJSON(reqLog.ResponseHeaders)

	// Default to empty arrays/objects for JSON fields if they're nil or empty
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

	// Insert the log using a validated, quoted table identifier. Values remain parameterized.
	// The table name is validated in NewPostgresStore via isValidPGTableName.
	// #nosec G201 -- identifiers cannot be parameterized; we quote a validated identifier to avoid injection.
	tbl := quotePGQualifiedIdentifier(s.tableName)
	query := fmt.Sprintf(`
		INSERT INTO %s (
			id, timestamp, "user", method, path, query, request_headers, response_headers,
			status_code, duration, request_body, response_body, error, 
			middleware_trace, route_trace
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
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
		log.Printf("Failed to store request log in PostgreSQL: %v", err)
	}

	// Clean up old logs
	s.cleanup()
}

// prepareJSON ensures we have a valid JSON string
func prepareJSON(v any) string {
	if v == nil {
		return "{}"
	}

	data, err := json.Marshal(v)
	if err != nil {
		log.Printf("Failed to marshal JSON: %v", err)
		return "{}"
	}

	return string(data)
}

// cleanup removes old logs to maintain the capacity limit
func (s *PostgresStore) cleanup() {
	tbl := quotePGQualifiedIdentifier(s.tableName)
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
	deleteQuery := fmt.Sprintf(`
		DELETE FROM %s
		WHERE id IN (
			SELECT id FROM %s
			ORDER BY timestamp ASC
			LIMIT $1
		)
	`, tbl, tbl)

	_, err = s.db.Exec(deleteQuery, count-s.capacity)
	if err != nil {
		log.Printf("Failed to clean up old logs: %v", err)
	}
}

// Get retrieves a specific request log by its ID
func (s *PostgresStore) Get(id string) (*model.RequestLog, bool) {
	tbl := quotePGQualifiedIdentifier(s.tableName)
	query := fmt.Sprintf(`
		SELECT 
			id, timestamp, method, path, query, 
			COALESCE(request_headers, '{}'),
			COALESCE(response_headers, '{}'),
			status_code, duration, request_body, response_body, error, 
			COALESCE(middleware_trace, '[]'),
			COALESCE(route_trace, '{}')
		FROM %s
		WHERE id = $1
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
		log.Printf("Failed to get request log from PostgreSQL: %v", err)
		return nil, false
	}

	// Unmarshal all JSON fields
	_ = json.Unmarshal([]byte(reqHeadersStr), &reqLog.RequestHeaders)
	_ = json.Unmarshal([]byte(respHeadersStr), &reqLog.ResponseHeaders)
	_ = json.Unmarshal([]byte(middlewareTrace), &reqLog.MiddlewareTrace)
	_ = json.Unmarshal([]byte(routeTrace), &reqLog.RouteTrace)

	return &reqLog, true
}

// GetAll returns all stored request logs
func (s *PostgresStore) GetAll() []*model.RequestLog {
	tbl := quotePGQualifiedIdentifier(s.tableName)
	query := fmt.Sprintf(`
		SELECT 
			id, timestamp, method, path, query, 
			COALESCE(request_headers::text, '{}'),
			COALESCE(response_headers::text, '{}'),
			status_code, duration, request_body, response_body, error, 
			COALESCE(middleware_trace::text, '[]'),
			COALESCE(route_trace::text, '{}')
		FROM %s
		ORDER BY timestamp DESC
	`, tbl)

	return s.queryLogs(query)
}

// GetLatest returns the n most recent request logs
func (s *PostgresStore) GetLatest(n int) []*model.RequestLog {
	tbl := quotePGQualifiedIdentifier(s.tableName)
	query := fmt.Sprintf(`
		SELECT 
			id, timestamp, method, path, query, 
			COALESCE(request_headers::text, '{}'),
			COALESCE(response_headers::text, '{}'),
			status_code, duration, request_body, response_body, error, 
			COALESCE(middleware_trace::text, '[]'),
			COALESCE(route_trace::text, '{}')
		FROM %s
		ORDER BY timestamp DESC
		LIMIT $1
	`, tbl)

	return s.queryLogs(query, n)
}

// queryLogs executes a query and returns the resulting log entries
func (s *PostgresStore) queryLogs(query string, args ...any) []*model.RequestLog {
	rows, err := s.db.Query(query, args...)
	if err != nil {
		log.Printf("Failed to query logs from PostgreSQL: %v", err)
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

		// Unmarshal all JSON fields, ignoring errors
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

// Clear clears all stored request logs
func (s *PostgresStore) Clear() error {
	tbl := quotePGQualifiedIdentifier(s.tableName)
	query := fmt.Sprintf("TRUNCATE TABLE %s", tbl)
	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to clear logs: %w", err)
	}

	return nil
}

// Close closes the database connection
func (s *PostgresStore) Close() error {
	return s.db.Close()
}
