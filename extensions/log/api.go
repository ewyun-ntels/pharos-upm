package log

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
	"ntels.com/pharos/core/pkg/authhandler"
	"ntels.com/pharos/core/pkg/common"
)

type opensearchConfig struct {
	Address            string        `toml:"address"`
	Username           string        `toml:"username"`
	Password           string        `toml:"password"`
	Indices            []string      `toml:"indices"`
	InsecureSkipVerify bool          `toml:"insecure_skip_verify"`
	IndexSchema        []indexSchema `toml:"index_schema"`
}

type indexSchema struct {
	Pattern string      `toml:"pattern" json:"pattern"`
	Filters []filterDef `toml:"filters" json:"filters"`
}

type filterDef struct {
	Name            string `toml:"name"             json:"name"`
	Field           string `toml:"field"            json:"field"`
	Display         string `toml:"display"          json:"display"`
	CaseInsensitive bool   `toml:"case_insensitive" json:"case_insensitive"`
}

type logTomlConfig struct {
	Log struct {
		OpenSearch opensearchConfig `toml:"opensearch"`
	} `toml:"log"`
}

// Api proxies log search requests to OpenSearch.
type Api struct {
	cfg    opensearchConfig
	client *http.Client
}

func (a *Api) Init(configPath string, _ common.Config) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Warn("log: failed to read config file", "path", configPath, "error", err)
		return
	}
	var cfg logTomlConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		slog.Warn("log: failed to parse [log.opensearch] config", "error", err)
		return
	}
	a.cfg = cfg.Log.OpenSearch
	if a.cfg.Address == "" {
		slog.Warn("log: [log.opensearch] address is not configured")
	} else {
		slog.Info("log: opensearch address loaded", "address", a.cfg.Address)
	}
	if len(a.cfg.Indices) == 0 {
		a.cfg.Indices = []string{"app-*", "node-*"}
	}
	transport := &http.Transport{
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: a.cfg.InsecureSkipVerify, //nolint:gosec
		},
	}
	a.client = &http.Client{Timeout: 30 * time.Second, Transport: transport}
}

func (a *Api) Use() bool   { return true }
func (a *Api) Load() error { return nil }
func (a *Api) Unload()     {}

func (a *Api) GetRelativePath() string { return "/log" }

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	auth := authhandler.GetAuthenticationHandler(true)
	routes.GET("/config", auth, a.getConfig)
	routes.GET("/aggregations", auth, a.aggregations)
	routes.POST("/search", auth, a.search)
}

func (a *Api) getConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"indices":      a.cfg.Indices,
		"address":      a.cfg.Address,
		"index_schema": a.cfg.IndexSchema,
	})
}

type searchRequest struct {
	Index    string            `json:"index"`
	Query    string            `json:"query"`
	FromTime string            `json:"from_time"`
	ToTime   string            `json:"to_time"`
	From     int               `json:"from"`
	Size     int               `json:"size"`
	Filters  map[string]string `json:"filters"` // Dynamic filters based on index schema
}

func (a *Api) search(c *gin.Context) {
	if a.cfg.Address == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "opensearch address not configured in [log.opensearch]"})
		return
	}

	var req searchRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if req.Size <= 0 || req.Size > 1000 {
		req.Size = 500
	}
	if req.FromTime == "" {
		req.FromTime = "now-15m"
	}
	if req.ToTime == "" {
		req.ToTime = "now"
	}
	if req.Index == "" && len(a.cfg.Indices) > 0 {
		req.Index = a.cfg.Indices[0]
	}
	if !a.isValidIndex(req.Index) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid index"})
		return
	}

	osQuery := a.buildOSQuery(req)
	body, err := json.Marshal(osQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	index := strings.TrimSpace(req.Index)
	target := fmt.Sprintf("%s/%s/_search", strings.TrimRight(a.cfg.Address, "/"), index)

	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if a.cfg.Username != "" {
		httpReq.SetBasicAuth(a.cfg.Username, a.cfg.Password)
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		slog.Error("log: opensearch request failed", "url", target, "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	c.Header("Content-Type", "application/json")
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
}

func (a *Api) buildOSQuery(req searchRequest) map[string]interface{} {
	mustClauses := []interface{}{
		map[string]interface{}{
			"range": map[string]interface{}{
				"@timestamp": map[string]interface{}{
					"gte": req.FromTime,
					"lte": req.ToTime,
				},
			},
		},
	}

	q := strings.TrimSpace(req.Query)
	if q != "" && q != "*" {
		mustClauses = append(mustClauses, map[string]interface{}{
			"query_string": map[string]interface{}{
				"query":            q,
				"default_operator": "AND",
				"lenient":          true,
			},
		})
	}

	boolQuery := map[string]interface{}{"must": mustClauses}

	filters := a.buildTermFilters(req)
	if len(filters) > 0 {
		boolQuery["filter"] = filters
	}

	return map[string]interface{}{
		"query": map[string]interface{}{
			"bool": boolQuery,
		},
		"sort": []interface{}{
			map[string]interface{}{
				"@timestamp": map[string]interface{}{"order": "desc"},
			},
		},
		"from":             req.From,
		"size":             req.Size,
		"track_total_hits": true,
	}
}

// findSchemaForIndex returns the most-specific schema whose pattern prefix matches index.
// Longer prefix wins, so "app-debug-*" beats "app-*" for "app-debug-logs".
func (a *Api) findSchemaForIndex(index string) *indexSchema {
	best := -1
	bestLen := -1
	for i := range a.cfg.IndexSchema {
		prefix := strings.TrimSuffix(a.cfg.IndexSchema[i].Pattern, "*")
		if strings.HasPrefix(index, prefix) && len(prefix) > bestLen {
			best = i
			bestLen = len(prefix)
		}
	}
	if best < 0 {
		return nil
	}
	return &a.cfg.IndexSchema[best]
}

// isValidIndex checks that the requested index matches one of the configured patterns.
func (a *Api) isValidIndex(index string) bool {
	for _, idx := range a.cfg.Indices {
		if idx == index {
			return true
		}
		if strings.HasSuffix(idx, "*") && strings.HasPrefix(index, strings.TrimSuffix(idx, "*")) {
			return true
		}
	}
	return false
}

// buildFilterClause builds a single OpenSearch filter clause for a field/value pair.
func buildFilterClause(field, value string, caseInsensitive bool) map[string]interface{} {
	if caseInsensitive {
		// wildcard with case_insensitive:true handles "info"/"INFO"/"Info" uniformly.
		return map[string]interface{}{
			"wildcard": map[string]interface{}{
				field: map[string]interface{}{
					"value":            strings.ToLower(value),
					"case_insensitive": true,
				},
			},
		}
	}
	return map[string]interface{}{
		"term": map[string]interface{}{field: value},
	}
}

func (a *Api) buildTermFilters(req searchRequest) []interface{} {
	var filters []interface{}
	schema := a.findSchemaForIndex(req.Index)
	if schema == nil {
		return filters
	}
	for _, fd := range schema.Filters {
		value, ok := req.Filters[fd.Name]
		if !ok || value == "" {
			continue
		}
		filters = append(filters, buildFilterClause(fd.Field, value, fd.CaseInsensitive))
	}
	return filters
}

// aggregations returns distinct values for filter dropdowns based on index schema.
func (a *Api) aggregations(c *gin.Context) {
	if a.cfg.Address == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "opensearch address not configured"})
		return
	}

	index := c.Query("index")
	if index == "" && len(a.cfg.Indices) > 0 {
		index = a.cfg.Indices[0]
	}
	if !a.isValidIndex(index) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid index"})
		return
	}
	fromTime := c.DefaultQuery("from_time", "now-15m")
	toTime := c.DefaultQuery("to_time", "now")

	schema := a.findSchemaForIndex(index)
	if schema == nil {
		c.JSON(http.StatusOK, gin.H{})
		return
	}
	
	// Get filter values from query params (for cascading filters)
	appliedFilters := make(map[string]string)
	for _, fd := range schema.Filters {
		if val := c.Query(fd.Name); val != "" {
			appliedFilters[fd.Name] = val
		}
	}

	// Build query with time range and any applied filters
	mustClauses := []interface{}{
		map[string]interface{}{
			"range": map[string]interface{}{
				"@timestamp": map[string]interface{}{"gte": fromTime, "lte": toTime},
			},
		},
	}
	
	boolQuery := map[string]interface{}{"must": mustClauses}
	
	// Apply filters for cascading — reuse the same clause builder as search.
	if len(appliedFilters) > 0 {
		var filterClauses []interface{}
		for _, fd := range schema.Filters {
			if value, ok := appliedFilters[fd.Name]; ok {
				filterClauses = append(filterClauses, buildFilterClause(fd.Field, value, fd.CaseInsensitive))
			}
		}
		if len(filterClauses) > 0 {
			boolQuery["filter"] = filterClauses
		}
	}

	// Build aggregations for all filters defined in schema
	aggs := make(map[string]interface{})
	for _, fd := range schema.Filters {
		aggs[fd.Name] = map[string]interface{}{
			"terms": map[string]interface{}{
				"field": fd.Field,
				"size":  500,
			},
		}
	}

	aggQuery := map[string]interface{}{
		"query": map[string]interface{}{"bool": boolQuery},
		"aggs":  aggs,
		"size":  0,
	}

	body, err := json.Marshal(aggQuery)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	target := fmt.Sprintf("%s/%s/_search", strings.TrimRight(a.cfg.Address, "/"), strings.TrimSpace(index))
	httpReq, err := http.NewRequestWithContext(c.Request.Context(), http.MethodPost, target, bytes.NewReader(body))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	httpReq.Header.Set("Content-Type", "application/json")
	if a.cfg.Username != "" {
		httpReq.SetBasicAuth(a.cfg.Username, a.cfg.Password)
	}

	resp, err := a.client.Do(httpReq)
	if err != nil {
		slog.Error("log: aggregations request failed", "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	var raw map[string]interface{}
	if err := json.NewDecoder(resp.Body).Decode(&raw); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Extract bucket keys for each filter
	result := make(gin.H)
	for _, fd := range schema.Filters {
		result[fd.Name] = extractBucketKeys(raw, fd.Name)
	}
	
	c.JSON(http.StatusOK, result)
}

func extractBucketKeys(raw map[string]interface{}, aggName string) []string {
	aggs, ok := raw["aggregations"].(map[string]interface{})
	if !ok {
		return nil
	}
	agg, ok := aggs[aggName].(map[string]interface{})
	if !ok {
		return nil
	}
	buckets, ok := agg["buckets"].([]interface{})
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(buckets))
	for _, b := range buckets {
		bMap, ok := b.(map[string]interface{})
		if !ok {
			continue
		}
		keyVal := bMap["key"]
		if keyVal == nil {
			continue
		}
		var s string
		switch v := keyVal.(type) {
		case string:
			s = v
		case float64:
			// JSON numbers (e.g. PRIORITY: 6) are decoded as float64
			s = strconv.FormatFloat(v, 'f', -1, 64)
		default:
			s = fmt.Sprintf("%v", v)
		}
		if s != "" {
			keys = append(keys, s)
		}
	}
	return keys
}
