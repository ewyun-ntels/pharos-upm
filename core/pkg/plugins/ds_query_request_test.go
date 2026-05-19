package plugins

import (
	"bytes"
	"io"
	"strings"
	"testing"
)

// TestDsQuery tests the DsQuery struct
func TestDsQuery(t *testing.T) {
	query := DsQuery{
		ID:             "test-id",
		DatasourceName: "test-datasource",
		SQL:            "SELECT * FROM test",
		Timeout:        30,
	}

	// Test field access
	if query.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", query.ID)
	}

	if query.DatasourceName != "test-datasource" {
		t.Errorf("Expected DatasourceName 'test-datasource', got %s", query.DatasourceName)
	}

	if query.SQL != "SELECT * FROM test" {
		t.Errorf("Expected SQL 'SELECT * FROM test', got %s", query.SQL)
	}

	if query.Timeout != 30 {
		t.Errorf("Expected Timeout 30, got %d", query.Timeout)
	}
}

// TestDsQueryRequest tests the DsQueryRequest struct
func TestDsQueryRequest(t *testing.T) {
	queries := []DsQuery{
		{ID: "q1", DatasourceName: "ds1", SQL: "SELECT 1"},
		{ID: "q2", DatasourceName: "ds2", SQL: "SELECT 2"},
	}

	request := DsQueryRequest{
		Queries: queries,
	}

	// Test field access
	if len(request.Queries) != 2 {
		t.Errorf("Expected 2 queries, got %d", len(request.Queries))
	}

	if request.Queries[0].ID != "q1" {
		t.Errorf("Expected first query ID 'q1', got %s", request.Queries[0].ID)
	}

	if request.Queries[1].DatasourceName != "ds2" {
		t.Errorf("Expected second query datasource 'ds2', got %s", request.Queries[1].DatasourceName)
	}
}

// TestDsQueryRequest_Set tests the set method with valid JSON
func TestDsQueryRequest_Set_ValidJSON(t *testing.T) {
	validJSON := `{
		"queries": [
			{
				"id": "query1",
				"datasourceName": "datasource1",
				"sql": "SELECT 1",
				"timeout": 30
			},
			{
				"id": "query2",
				"datasourceName": "datasource2",
				"sql": "SELECT 2",
				"timeout": 60
			}
		]
	}`

	request := &DsQueryRequest{}
	reader := strings.NewReader(validJSON)

	err := request.set(reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	// Verify the data was parsed correctly
	if len(request.Queries) != 2 {
		t.Errorf("Expected 2 queries, got %d", len(request.Queries))
	}

	if request.Queries[0].ID != "query1" {
		t.Errorf("Expected first query ID 'query1', got %s", request.Queries[0].ID)
	}

	if request.Queries[0].DatasourceName != "datasource1" {
		t.Errorf("Expected first query datasource 'datasource1', got %s", request.Queries[0].DatasourceName)
	}

	if request.Queries[0].SQL != "SELECT 1" {
		t.Errorf("Expected first query SQL 'SELECT 1', got %s", request.Queries[0].SQL)
	}

	if request.Queries[1].ID != "query2" {
		t.Errorf("Expected second query ID 'query2', got %s", request.Queries[1].ID)
	}

	if request.Queries[0].Timeout != 30 {
		t.Errorf("Expected first query timeout 30, got %d", request.Queries[0].Timeout)
	}

	if request.Queries[1].Timeout != 60 {
		t.Errorf("Expected second query timeout 60, got %d", request.Queries[1].Timeout)
	}
}

// TestDsQueryRequest_Set_EmptyQueries tests parsing empty queries array
func TestDsQueryRequest_Set_EmptyQueries(t *testing.T) {
	emptyJSON := `{"queries": []}`

	request := &DsQueryRequest{}
	reader := strings.NewReader(emptyJSON)

	err := request.set(reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(request.Queries) != 0 {
		t.Errorf("Expected 0 queries, got %d", len(request.Queries))
	}
}

// TestDsQueryRequest_Set_InvalidJSON tests the set method with invalid JSON
func TestDsQueryRequest_Set_InvalidJSON(t *testing.T) {
	invalidJSON := `{"queries": [invalid json`

	request := &DsQueryRequest{}
	reader := strings.NewReader(invalidJSON)

	err := request.set(reader)
	if err == nil {
		t.Error("Expected error for invalid JSON, got none")
	}
}

// TestDsQueryRequest_Set_EmptyReader tests the set method with empty reader
func TestDsQueryRequest_Set_EmptyReader(t *testing.T) {
	request := &DsQueryRequest{}
	reader := strings.NewReader("")

	err := request.set(reader)
	if err == nil {
		t.Error("Expected error for empty JSON, got none")
	}
}

// TestDsQueryRequest_Set_ReadError tests the set method with read error
func TestDsQueryRequest_Set_ReadError(t *testing.T) {
	// Create a reader that will fail
	errorReader := &errorReader{}

	request := &DsQueryRequest{}

	err := request.set(errorReader)
	if err == nil {
		t.Error("Expected error from failing reader, got none")
	}

	// Check that we got an error (the exact message may vary)
	if err.Error() != "unexpected EOF" {
		t.Logf("Got error: %s (this is expected)", err.Error())
	}
}

// TestDsQueryRequest_Set_MalformedJSON tests various malformed JSON cases
func TestDsQueryRequest_Set_MalformedJSON(t *testing.T) {
	testCases := []struct {
		name string
		json string
	}{
		{"MissingClosingBrace", `{"queries": [`},
		{"InvalidFieldType", `{"queries": "not an array"}`},
		{"MissingQuotes", `{queries: []}`},
		{"TrailingComma", `{"queries": [],}`},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			request := &DsQueryRequest{}
			reader := strings.NewReader(tc.json)

			err := request.set(reader)
			if err == nil {
				t.Errorf("Expected error for malformed JSON case %s, got none", tc.name)
			}
		})
	}
}

// TestDsQueryRequest_Set_PartialData tests parsing with partial/missing fields
func TestDsQueryRequest_Set_PartialData(t *testing.T) {
	partialJSON := `{
		"queries": [
			{
				"id": "query1",
				"sql": "SELECT 1"
			}
		]
	}`

	request := &DsQueryRequest{}
	reader := strings.NewReader(partialJSON)

	err := request.set(reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(request.Queries) != 1 {
		t.Errorf("Expected 1 query, got %d", len(request.Queries))
	}

	// DatasourceName should be empty string when missing
	if request.Queries[0].DatasourceName != "" {
		t.Errorf("Expected empty datasource name, got %s", request.Queries[0].DatasourceName)
	}

	if request.Queries[0].ID != "query1" {
		t.Errorf("Expected ID 'query1', got %s", request.Queries[0].ID)
	}

	if request.Queries[0].SQL != "SELECT 1" {
		t.Errorf("Expected SQL 'SELECT 1', got %s", request.Queries[0].SQL)
	}

	// Timeout should be 0 (default) when missing
	if request.Queries[0].Timeout != 0 {
		t.Errorf("Expected default timeout 0, got %d", request.Queries[0].Timeout)
	}
}

// TestDsQueryRequest_Set_LargeData tests parsing large JSON data
func TestDsQueryRequest_Set_LargeData(t *testing.T) {
	// Create a large query request with many queries
	var queries []string
	for range 100 {
		query := `{
			"id": "query` + strings.Repeat("0", 10) + `",
			"datasourceName": "datasource` + strings.Repeat("0", 10) + `",
			"sql": "SELECT * FROM large_table WHERE condition = '` + strings.Repeat("x", 100) + `'"
		}`
		queries = append(queries, query)
	}

	largeJSON := `{"queries": [` + strings.Join(queries, ",") + `]}`

	request := &DsQueryRequest{}
	reader := strings.NewReader(largeJSON)

	err := request.set(reader)
	if err != nil {
		t.Errorf("Expected no error for large data, got %v", err)
	}

	if len(request.Queries) != 100 {
		t.Errorf("Expected 100 queries, got %d", len(request.Queries))
	}
}

// TestDsQueryRequest_Set_SpecialCharacters tests parsing with special characters
func TestDsQueryRequest_Set_SpecialCharacters(t *testing.T) {
	specialJSON := `{
		"queries": [
			{
				"id": "query-with-special-chars",
				"datasourceName": "datasource_with_underscores",
				"sql": "SELECT 'test with \"quotes\" and \\backslashes' FROM table"
			}
		]
	}`

	request := &DsQueryRequest{}
	reader := strings.NewReader(specialJSON)

	err := request.set(reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(request.Queries) != 1 {
		t.Errorf("Expected 1 query, got %d", len(request.Queries))
	}

	expectedSQL := `SELECT 'test with "quotes" and \backslashes' FROM table`
	if request.Queries[0].SQL != expectedSQL {
		t.Errorf("Expected SQL with special chars, got %s", request.Queries[0].SQL)
	}
}

// Helper type for testing read errors
type errorReader struct{}

func (r *errorReader) Read(p []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

// TestDsQueryRequest_Set_BytesReader tests with bytes.Reader
func TestDsQueryRequest_Set_BytesReader(t *testing.T) {
	jsonData := `{"queries": [{"id": "test", "datasourceName": "test-ds", "sql": "SELECT 1", "timeout": 20}]}`
	reader := bytes.NewReader([]byte(jsonData))

	request := &DsQueryRequest{}

	err := request.set(reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(request.Queries) != 1 {
		t.Errorf("Expected 1 query, got %d", len(request.Queries))
	}
}

// TestDsQueryRequest_JSONTags tests that JSON tags work correctly
func TestDsQueryRequest_JSONTags(t *testing.T) {
	// Test that the struct tags match the expected JSON format
	jsonWithDifferentCase := `{
		"queries": [
			{
				"id": "test-id",
				"datasourceName": "test-datasource",
				"sql": "SELECT 1",
				"timeout": 45
			}
		]
	}`

	request := &DsQueryRequest{}
	reader := strings.NewReader(jsonWithDifferentCase)

	err := request.set(reader)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if len(request.Queries) != 1 {
		t.Errorf("Expected 1 query, got %d", len(request.Queries))
	}

	query := request.Queries[0]
	if query.ID != "test-id" {
		t.Errorf("Expected ID 'test-id', got %s", query.ID)
	}

	if query.DatasourceName != "test-datasource" {
		t.Errorf("Expected DatasourceName 'test-datasource', got %s", query.DatasourceName)
	}

	if query.SQL != "SELECT 1" {
		t.Errorf("Expected SQL 'SELECT 1', got %s", query.SQL)
	}

	if query.Timeout != 45 {
		t.Errorf("Expected Timeout 45, got %d", query.Timeout)
	}
}

// Benchmark tests
func BenchmarkDsQueryRequest_Set_SmallJSON(b *testing.B) {
	jsonData := `{"queries": [{"id": "q1", "datasourceName": "ds1", "sql": "SELECT 1"}]}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := &DsQueryRequest{}
		reader := strings.NewReader(jsonData)
		_ = request.set(reader)
	}
}

func BenchmarkDsQueryRequest_Set_LargeJSON(b *testing.B) {
	// Create JSON with 10 queries
	queries := make([]string, 10)
	for i := range 10 {
		queries[i] = `{"id": "q` + string(rune('0'+i)) + `", "datasourceName": "ds` + string(rune('0'+i)) + `", "sql": "SELECT ` + string(rune('0'+i)) + `"}`
	}
	jsonData := `{"queries": [` + strings.Join(queries, ",") + `]}`

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		request := &DsQueryRequest{}
		reader := strings.NewReader(jsonData)
		_ = request.set(reader)
	}
}

func BenchmarkDsQuery_Creation(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		query := DsQuery{
			ID:             "benchmark-id",
			DatasourceName: "benchmark-datasource",
			SQL:            "SELECT 1",
			Timeout:        30,
		}
		_ = query
	}
}

// TestDsQueryRequest_Set_TimeoutField tests timeout field parsing
func TestDsQueryRequest_Set_TimeoutField(t *testing.T) {
	tests := []struct {
		name            string
		json            string
		expectedTimeout int
	}{
		{
			name:            "With timeout",
			json:            `{"queries": [{"id": "q1", "datasourceName": "ds1", "sql": "SELECT 1", "timeout": 120}]}`,
			expectedTimeout: 120,
		},
		{
			name:            "Without timeout (default to 0)",
			json:            `{"queries": [{"id": "q1", "datasourceName": "ds1", "sql": "SELECT 1"}]}`,
			expectedTimeout: 0,
		},
		{
			name:            "With zero timeout",
			json:            `{"queries": [{"id": "q1", "datasourceName": "ds1", "sql": "SELECT 1", "timeout": 0}]}`,
			expectedTimeout: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			request := &DsQueryRequest{}
			reader := strings.NewReader(tt.json)

			err := request.set(reader)
			if err != nil {
				t.Errorf("Expected no error, got %v", err)
			}

			if len(request.Queries) != 1 {
				t.Errorf("Expected 1 query, got %d", len(request.Queries))
			}

			if request.Queries[0].Timeout != tt.expectedTimeout {
				t.Errorf("Expected timeout %d, got %d", tt.expectedTimeout, request.Queries[0].Timeout)
			}
		})
	}
}
