package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"
	"os"
	"sort"
	"strings"
	"time"

	"github.com/pelletier/go-toml/v2"
)

type fileConfig struct {
	Server struct {
		Base         string `toml:"base"`
		ClientID     string `toml:"client_id"`
		ClientSecret string `toml:"client_secret"`
		Username     string `toml:"username"`
		Password     string `toml:"password"`
	} `toml:"server"`
	Clickhouse struct {
		Database string `toml:"database"`
		Schema   string `toml:"schema"`
		Host     string `toml:"host"`
		HTTPPort int    `toml:"http_port"`
		Username string `toml:"username"`
		Password string `toml:"password"`
	} `toml:"clickhouse"`
	Simulator struct {
		RuleName   string   `toml:"rule_name"`
		Duration   string   `toml:"duration"`
		Interval   int      `toml:"interval"`
		Poll       string   `toml:"poll"`
		LabelKey   string   `toml:"label_key"`
		Labels     []string `toml:"labels"`
		Datasource string   `toml:"datasource"`
	} `toml:"simulator"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type ruleRequest struct {
	AlertType string                 `json:"alert_type"`
	Rule      map[string]interface{} `json:"rule"`
}

type statusValue struct {
	Name     string            `json:"name"`
	Severity string            `json:"severity"`
	Status   string            `json:"status"`
	Labels   map[string]string `json:"labels"`
	Value    float64           `json:"value"`
	AlertId  string            `json:"alert_id"`
}

// keys returns sorted keys of a map[string]int for pretty logging
func keys(m map[string]int) []string {
	ks := make([]string, 0, len(m))
	for k := range m {
		ks = append(ks, k)
	}
	sort.Strings(ks)
	return ks
}

func getToken(baseURL, clientID, clientSecret, username, password string) (string, error) {
	// Enforce Resource Owner Password Credentials grant only
	if username == "" || password == "" {
		return "", fmt.Errorf("username/password required for password grant; provide via flags or config")
	}

	form := url.Values{}
	form.Set("grant_type", "password")
	form.Set("username", username)
	form.Set("password", password)
	if clientID != "" {
		form.Set("client_id", clientID)
	}
	// Some deployments may accept empty client_secret for public clients; include if provided
	if clientSecret != "" {
		form.Set("client_secret", clientSecret)
	}
	resp, err := http.PostForm(strings.TrimRight(baseURL, "/")+"/auth/token", form)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("password grant failed: %s: %s", resp.Status, string(b))
	}
	var tr tokenResponse
	if err := json.NewDecoder(resp.Body).Decode(&tr); err != nil {
		return "", err
	}
	if tr.AccessToken == "" {
		return "", fmt.Errorf("empty access_token in response")
	}
	return tr.AccessToken, nil
}

func authReq(method, urlStr, token string, body any) (*http.Response, error) {
	var rdr io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		rdr = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, urlStr, rdr)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if token != "" {
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return http.DefaultClient.Do(req)
}

func postRule(baseURL, token, ruleName string, intervalSec int, labelKey string, labels []string, datasourceName string) error {
	// Query alternates value 90 (occur) for 10s and 1 (clear) for next 10s, repeating.
	// Use arrayJoin to emit one row per label value so we can verify multiples simultaneously.
	if labelKey == "" {
		labelKey = "host"
	}
	if len(labels) == 0 {
		labels = []string{"sim"}
	}
	// Build arrayJoin list
	var b strings.Builder
	b.WriteString("[")
	for i, v := range labels {
		if i > 0 {
			b.WriteString(",")
		}
		b.WriteString("'")
		// naive escaping for single quotes by doubling
		b.WriteString(strings.ReplaceAll(v, "'", "''"))
		b.WriteString("'")
	}
	b.WriteString("]")
	array := b.String()
	// The label column must be named as labelKey and include an index per label for deterministic grouping
	// We generate mixed states per evaluation: roughly half labels occur and half do not, and active values differ per label.
	query := fmt.Sprintf(
		"SELECT formatDateTime(now(), '%%F %%T') as ts, toString(\n"+
			"    if( ((idx %% 2) = 0 AND (toUInt32(toUnixTimestamp(now())) %% 20) < 10)\n"+
			"        OR ((idx %% 2) = 1 AND (toUInt32(toUnixTimestamp(now())) %% 20) >= 10),\n"+
			"        80 + (idx %% 20),\n"+
			"        1 + (idx %% 5)\n"+
			"    )\n"+
			") as val, label as %s\n"+
			"FROM (SELECT %s AS arr)\n"+
			"ARRAY JOIN arr AS label, arrayEnumerate(arr) AS idx",
		labelKey, array,
	)

	// Build threshold with explicit id and include the same id in labels, so the
	// simulator can validate using threshold_id from Labels rather than relying on AlertId.
	thresholdId := "sim-high"
	rule := map[string]interface{}{
		"id":               "",
		"name":             ruleName,
		"description":      "alert simulator",
		"datasource":       datasourceName,
		"datasource_query": map[string]interface{}{"query": query, "time_label": "ts", "variable_label": "val"},
		"threshold": []map[string]interface{}{
			{"id": thresholdId, "condition": "is_above", "start_value": 80, "severity": "critical", "labels": map[string]string{"threshold_id": thresholdId}},
		},
		"query_delay_offset":  0,
		"evaluation_interval": fmt.Sprintf("*/%d * * * * *", intervalSec),
		"check_type":          "last",
		"notifications":       []string{},
	}
	payload := ruleRequest{AlertType: "query", Rule: rule}
	resp, err := authReq("POST", strings.TrimRight(baseURL, "/")+"/alert/rule", token, payload)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("post /alert/rule failed: %s: %s", resp.Status, string(b))
	}
	return nil
}

func getStatuses(baseURL, token string) ([]statusValue, error) {
	resp, err := authReq("GET", strings.TrimRight(baseURL, "/")+"/alert/status", token, nil)
	if err != nil {
		return nil, err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("get /alert/status failed: %s: %s", resp.Status, string(b))
	}
	var vals []statusValue
	if err := json.NewDecoder(resp.Body).Decode(&vals); err != nil {
		return nil, err
	}
	return vals, nil
}

func getRuleIdByName(baseURL, token, name string) (string, error) {
	resp, err := authReq("GET", strings.TrimRight(baseURL, "/")+"/alert/rule?detail=1", token, nil)
	if err != nil {
		return "", err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("get /alert/rule failed: %s: %s", resp.Status, string(b))
	}
	// Response is a list of objects; we only need to parse minimally for id and rule.name
	var list []struct {
		Rule map[string]any `json:"rule"`
		// id exposed only when not detailed? The DB fields exist; when detail, Rule map contains id and others.
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return "", err
	}
	for _, it := range list {
		if it.Rule == nil {
			continue
		}
		if rn, ok := it.Rule["name"].(string); ok && rn == name {
			if id, ok := it.Rule["id"].(string); ok && id != "" {
				return id, nil
			}
		}
	}
	return "", fmt.Errorf("rule %q not found", name)
}

func deleteRule(baseURL, token, id string) error {
	resp, err := authReq("DELETE", strings.TrimRight(baseURL, "/")+"/alert/rule/"+url.PathEscape(id), token, nil)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("delete /alert/rule/%s failed: %s: %s", id, resp.Status, string(b))
	}
	return nil
}

// cleanupRemainingStatuses removes any lingering alert_status rows for a rule name.
func cleanupRemainingStatuses(baseURL, token, ruleName string) error {
	resp, err := authReq("GET", strings.TrimRight(baseURL, "/")+"/alert/rule?detail=1", token, nil)
	if err != nil {
		return err
	}
	defer func(Body io.ReadCloser) {
		err := Body.Close()
		if err != nil {

		}
	}(resp.Body)
	if resp.StatusCode != http.StatusOK {
		b, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("get /alert/rule failed: %s: %s", resp.Status, string(b))
	}
	var list []struct {
		Rule   map[string]any `json:"rule"`
		Status []statusValue  `json:"status"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&list); err != nil {
		return err
	}
	for _, it := range list {
		if it.Rule == nil {
			continue
		}
		if rn, ok := it.Rule["name"].(string); ok && rn == ruleName {
			for _, s := range it.Status {
				_, derr := authReq("DELETE", strings.TrimRight(baseURL, "/")+"/alert/status/"+url.PathEscape(ruleName)+"/"+url.PathEscape(s.AlertId), token, nil)
				if derr != nil {
					log.Printf("warning: failed to delete lingering status id=%s: %v", s.AlertId, derr)
				}
			}
			break
		}
	}
	return nil
}

func loadConfig(path string) (*fileConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg fileConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func main() {
	configPath := flag.String("config", "./config.toml", "Path to simulator config TOML")
	flag.Parse()

	cfg, err := loadConfig(*configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// Validate required fields
	if strings.TrimSpace(cfg.Server.Base) == "" {
		log.Fatalf("config.server.base is required")
	}
	if strings.TrimSpace(cfg.Server.ClientID) == "" {
		log.Fatalf("config.server.client_id is required")
	}
	// Resource Owner Password Credentials required
	if strings.TrimSpace(cfg.Server.Username) == "" || strings.TrimSpace(cfg.Server.Password) == "" {
		log.Fatalf("config.server.username and config.server.password are required for password grant")
	}
	if cfg.Simulator.Interval <= 0 {
		cfg.Simulator.Interval = 10
	}
	if cfg.Simulator.RuleName == "" {
		cfg.Simulator.RuleName = fmt.Sprintf("sim-query-%d", time.Now().Unix())
	}
	pollDelay := 10 * time.Second
	if cfg.Simulator.Poll != "" {
		if d, err := time.ParseDuration(cfg.Simulator.Poll); err == nil {
			pollDelay = d
		}
	}
	duration := 10 * time.Minute
	if cfg.Simulator.Duration != "" {
		if d, err := time.ParseDuration(cfg.Simulator.Duration); err == nil {
			duration = d
		}
	}

	baseURL := strings.TrimSpace(cfg.Server.Base)
	clientID := strings.TrimSpace(cfg.Server.ClientID)
	clientSecret := cfg.Server.ClientSecret
	username := cfg.Server.Username
	password := cfg.Server.Password
	ruleName := cfg.Simulator.RuleName
	interval := cfg.Simulator.Interval
	// datasource name for rule; default to clickhouse-01 if not specified
	datasourceName := strings.TrimSpace(cfg.Simulator.Datasource)
	if datasourceName == "" {
		datasourceName = "clickhouse-01"
	}

	log.Printf("alert simulator starting: base=%s name=%s duration=%s interval=%ds", baseURL, ruleName, duration.String(), interval)

	token, err := getToken(baseURL, clientID, clientSecret, username, password)
	if err != nil {
		log.Fatalf("failed to get token: %v", err)
	}

	// Ensure no stale rule with the same name exists (from a previous run)
	if oldId, err := getRuleIdByName(baseURL, token, ruleName); err == nil && oldId != "" {
		if derr := deleteRule(baseURL, token, oldId); derr != nil {
			log.Printf("warning: failed to delete pre-existing rule id=%s: %v", oldId, derr)
		} else {
			log.Printf("deleted pre-existing rule id=%s for name=%s", oldId, ruleName)
		}
	}
	// Also remove any lingering alert_status rows for this rule name (in case previous run left them)
	if derr := cleanupRemainingStatuses(baseURL, token, ruleName); derr != nil {
		log.Printf("warning: failed to cleanup lingering statuses via /alert/rule?detail=1 for %s: %v", ruleName, derr)
	}
	// Fallback: directly inspect /alert/status and delete any rows with this name
	if allStatus, derr := getStatuses(baseURL, token); derr == nil {
		for _, s := range allStatus {
			if s.Name == ruleName {
				_, e := authReq("DELETE", strings.TrimRight(baseURL, "/")+"/alert/status/"+url.PathEscape(ruleName)+"/"+url.PathEscape(s.AlertId), token, nil)
				if e != nil {
					log.Printf("warning: failed to delete lingering status (direct) id=%s: %v", s.AlertId, e)
				}
			}
		}
	}

	if err := postRule(baseURL, token, ruleName, interval, cfg.Simulator.LabelKey, cfg.Simulator.Labels, datasourceName); err != nil {
		log.Fatalf("failed to register rule: %v", err)
	}
	log.Printf("rule %q registered. Monitoring for %s...", ruleName, duration.String())

	end := time.Now().Add(duration)
	var occurCount, clearCount int
	mismatches := 0

	// Warm-up: give the cron job time to run at least once before strict assertions.
	registeredAt := time.Now()
	warmedUp := false

	isOccurPhase := func(t time.Time) bool {
		return (t.Unix() % 20) < 10
	}

	for time.Now().Before(end) {
		vals, err := getStatuses(baseURL, token)
		if err != nil {
			log.Printf("get status error: %v", err)
			time.Sleep(pollDelay)
			continue
		}

		// Warm-up gate: wait until either we observe first alerts for this rule
		// or enough time has passed for at least one cron evaluation.
		if !warmedUp {
			seen := false
			for _, v := range vals {
				if v.Name == ruleName {
					seen = true
					break
				}
			}
			if !seen {
				minWait := time.Duration(interval)*time.Second + pollDelay
				if time.Since(registeredAt) < minWait {
					log.Printf("warming up... waiting for first evaluation window (wait up to %s)", minWait-time.Since(registeredAt))
					time.Sleep(pollDelay)
					continue
				}
			} else {
				warmedUp = true
			}
		}

		// Align expected phase with the last cron evaluation boundary to avoid edge mismatches
		now := time.Now()
		lastTick := (now.Unix() / int64(interval)) * int64(interval)
		evalTime := time.Unix(lastTick, 0)
		phaseOccur := isOccurPhase(evalTime)
		// Determine expected active labels by idx parity (idx starts at 1 as per arrayEnumerate)
		labels := cfg.Simulator.Labels
		if len(labels) == 0 {
			labels = []string{"sim"}
		}
		expectedActiveSet := make(map[string]bool)
		for i, lv := range labels {
			idx := i + 1
			parity := idx % 2
			if (parity == 0 && phaseOccur) || (parity == 1 && !phaseOccur) {
				expectedActiveSet[lv] = true
			}
		}
		expectedActive := len(expectedActiveSet)

		// Filter this rule's active alerts
		active := make([]statusValue, 0)
		for _, v := range vals {
			if v.Name == ruleName && v.Status == "alerting" {
				active = append(active, v)
			}
		}

		// Count by label value and gather values for diversity checks
		labelKey := cfg.Simulator.LabelKey
		if labelKey == "" {
			labelKey = "host"
		}
		byLabel := make(map[string]int)
		valueSet := make(map[float64]bool)
		for _, a := range active {
			val := a.Labels[labelKey]
			byLabel[val]++
			valueSet[a.Value] = true
		}

		pollOk := true
		if len(active) != expectedActive {
			mismatches++
			pollOk = false
			log.Printf("ASSERT FAIL: expected %d active alerts, got %d (phase=%v)", expectedActive, len(active), phaseOccur)
		}

		// Check membership per label: expected ones present exactly once; others absent
		for _, lv := range labels {
			if expectedActiveSet[lv] {
				if byLabel[lv] != 1 {
					mismatches++
					pollOk = false
					log.Printf("ASSERT FAIL: expected 1 active for %s=%s, got %d", labelKey, lv, byLabel[lv])
				}
			} else {
				if byLabel[lv] != 0 {
					mismatches++
					pollOk = false
					log.Printf("ASSERT FAIL: expected 0 active for %s=%s, got %d", labelKey, lv, byLabel[lv])
				}
			}
		}
		// Sanity: all active should have threshold_id
		for _, a := range active {
			if a.Labels["threshold_id"] == "" {
				mismatches++
				pollOk = false
				log.Printf("ASSERT FAIL: missing threshold_id in labels for alert id=%s labels=%v", a.AlertId, a.Labels)
			}
		}
		// Values among actives should not all be identical if more than one active
		if expectedActive > 1 && len(valueSet) <= 1 {
			mismatches++
			pollOk = false
			log.Printf("ASSERT FAIL: expected diverse values among actives, got single value set: %v", valueSet)
		}

		if pollOk {
			// Success log per poll with brief summary
			log.Printf("OK: phase=%v active=%d/%d labels=%v values_distinct=%d", phaseOccur, len(active), expectedActive, keys(byLabel), len(valueSet))
		}

		// increment counters for end summary
		if len(active) > 0 {
			occurCount += len(active)
		} else {
			clearCount++
		}

		time.Sleep(pollDelay)
	}

	// Cleanup: delete the rule
	id, err := getRuleIdByName(baseURL, token, ruleName)
	if err != nil {
		log.Printf("warning: could not resolve rule id for deletion: %v", err)
	} else {
		if err := deleteRule(baseURL, token, id); err != nil {
			log.Printf("warning: failed to delete rule: %v", err)
		} else {
			log.Printf("deleted rule id=%s", id)
		}
	}

	log.Printf("simulation finished. occur_samples=%d clear_samples=%d mismatches=%d", occurCount, clearCount, mismatches)
	if mismatches > 0 {
		log.Printf("RESULT: FAIL")
		os.Exit(2)
	}
	log.Printf("RESULT: SUCCESS")
	os.Exit(0)
}
