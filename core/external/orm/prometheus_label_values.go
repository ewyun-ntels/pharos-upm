package orm

import (
	"context"
	"fmt"
	"log/slog"
	"sort"
	"strings"
	"time"

	prometheus_api_v1 "github.com/prometheus/client_golang/api/prometheus/v1"
)

type grafanaLabelValuesQuery struct {
	label   string
	matches []string
}

func IsGrafanaLabelValuesQuery(query string) bool {
	_, ok, _ := parseGrafanaLabelValuesQuery(query)
	return ok
}

func makeDatabaseResponseForPrometheusLabelValues(
	ctx context.Context,
	v1api prometheus_api_v1.API,
	query string,
	started time.Time,
) (DatabaseResponse, bool, error) {
	parsed, ok, err := parseGrafanaLabelValuesQuery(query)
	if !ok || err != nil {
		return DatabaseResponse{}, ok, err
	}

	var startTime, endTime time.Time
	if tr, ok := getPrometheusTimeRange(ctx); ok {
		startTime = time.Unix(tr.StartTime, 0)
		endTime = time.Unix(tr.EndTime, 0)
	}

	values, warnings, err := v1api.LabelValues(ctx, parsed.label, parsed.matches, startTime, endTime)
	if err != nil {
		return DatabaseResponse{}, true, err
	}
	if len(warnings) > 0 {
		slog.Warn("prometheus label_values warnings", "warnings", warnings)
	}

	rows := make([]string, 0, len(values))
	for _, value := range values {
		rows = append(rows, string(value))
	}
	sort.Strings(rows)

	response := DatabaseResponse{
		SQL: query,
		Meta: []map[string]string{
			{"name": "value"},
			{"name": "label"},
		},
		Rows: int64(len(rows)),
	}
	response.Statistics.Elapsed = time.Since(started).Seconds()

	for _, value := range rows {
		response.Data = append(response.Data, map[string]any{
			"value": value,
			"label": value,
		})
	}

	return response, true, nil
}

func parseGrafanaLabelValuesQuery(query string) (grafanaLabelValuesQuery, bool, error) {
	q := strings.TrimSpace(query)
	if !strings.HasPrefix(q, "label_values") {
		return grafanaLabelValuesQuery{}, false, nil
	}

	rest := strings.TrimSpace(strings.TrimPrefix(q, "label_values"))
	if !strings.HasPrefix(rest, "(") || !strings.HasSuffix(rest, ")") {
		return grafanaLabelValuesQuery{}, true, fmt.Errorf("invalid label_values query: %s", query)
	}

	args, err := splitLabelValuesArgs(rest[1 : len(rest)-1])
	if err != nil {
		return grafanaLabelValuesQuery{}, true, err
	}

	switch len(args) {
	case 1:
		label := normalizePrometheusLabelArg(args[0])
		if label == "" {
			return grafanaLabelValuesQuery{}, true, fmt.Errorf("label_values label is empty")
		}
		return grafanaLabelValuesQuery{label: label}, true, nil
	case 2:
		match := strings.TrimSpace(args[0])
		label := normalizePrometheusLabelArg(args[1])
		if match == "" || label == "" {
			return grafanaLabelValuesQuery{}, true, fmt.Errorf("label_values metric and label are required")
		}
		return grafanaLabelValuesQuery{label: label, matches: []string{match}}, true, nil
	default:
		return grafanaLabelValuesQuery{}, true, fmt.Errorf("label_values expects 1 or 2 arguments, got %d", len(args))
	}
}

func splitLabelValuesArgs(input string) ([]string, error) {
	args := []string{}
	start := 0
	braceDepth := 0
	parenDepth := 0
	var quote rune
	escaped := false

	for i, ch := range input {
		if quote != 0 {
			if escaped {
				escaped = false
				continue
			}
			if ch == '\\' {
				escaped = true
				continue
			}
			if ch == quote {
				quote = 0
			}
			continue
		}

		switch ch {
		case '"', '\'':
			quote = ch
		case '{':
			braceDepth++
		case '}':
			if braceDepth == 0 {
				return nil, fmt.Errorf("unbalanced brace in label_values query")
			}
			braceDepth--
		case '(':
			parenDepth++
		case ')':
			if parenDepth == 0 {
				return nil, fmt.Errorf("unbalanced parenthesis in label_values query")
			}
			parenDepth--
		case ',':
			if braceDepth == 0 && parenDepth == 0 {
				args = append(args, strings.TrimSpace(input[start:i]))
				start = i + len(string(ch))
			}
		}
	}

	if quote != 0 {
		return nil, fmt.Errorf("unterminated quoted string in label_values query")
	}
	if braceDepth != 0 || parenDepth != 0 {
		return nil, fmt.Errorf("unbalanced selector in label_values query")
	}

	args = append(args, strings.TrimSpace(input[start:]))
	return args, nil
}

func normalizePrometheusLabelArg(value string) string {
	value = strings.TrimSpace(value)
	if len(value) >= 2 {
		first := value[0]
		last := value[len(value)-1]
		if (first == '"' && last == '"') || (first == '\'' && last == '\'') {
			value = strings.TrimSpace(value[1 : len(value)-1])
		}
	}
	return value
}
