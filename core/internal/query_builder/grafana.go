package query_builder

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"strconv"
	"strings"
)

var grafanaVariablePattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?:\.([A-Za-z_][A-Za-z0-9_.]*))?(?::([^}]+))?\}|\[\[([A-Za-z_][A-Za-z0-9_]*)(?::([^\]]+))?\]\]|\$([A-Za-z_][A-Za-z0-9_]*)`)

func buildGrafana(query string, queryContext map[string]any) ([]byte, error) {
	convertContext, err := convertInternalContext(queryContext)
	if err != nil {
		return nil, err
	}

	query = replaceGrafanaTimeAliases(query, convertContext)

	out := grafanaVariablePattern.ReplaceAllStringFunc(query, func(match string) string {
		parts := grafanaVariablePattern.FindStringSubmatch(match)
		if len(parts) == 0 {
			return match
		}

		varName, fieldPath, format := grafanaVariableParts(parts)
		if varName == "" {
			return match
		}

		if fieldPath != "" {
			slog.Debug("Keeping Grafana field path variable as original token",
				"token", match,
				"variable", varName,
				"fieldPath", fieldPath,
			)
			return match
		}

		value, exists := convertContext[varName]
		if !exists {
			slog.Debug("Keeping unknown Grafana variable as original token",
				"token", match,
				"variable", varName,
			)
			return match
		}

		if !isSupportedGrafanaFormat(format) {
			slog.Debug("Keeping unsupported Grafana variable format as original token",
				"token", match,
				"variable", varName,
				"format", format,
			)
			return match
		}

		return formatGrafanaValue(value, format)
	})

	return []byte(out), nil
}

func grafanaVariableParts(parts []string) (varName string, fieldPath string, format string) {
	switch {
	case parts[1] != "":
		return parts[1], parts[2], strings.TrimSpace(parts[3])
	case parts[4] != "":
		return parts[4], "", strings.TrimSpace(parts[5])
	case parts[6] != "":
		return parts[6], "", ""
	default:
		return "", "", ""
	}
}

func replaceGrafanaTimeAliases(query string, context map[string]any) string {
	replacements := []struct {
		token string
		value string
	}{
		{"${__from:date:seconds}", grafanaTimeValue(context, "__start_time", "__start_time_ms", 1)},
		{"${__to:date:seconds}", grafanaTimeValue(context, "__end_time", "__end_time_ms", 1)},
		{"${__from:date:milliseconds}", grafanaTimeValue(context, "__start_time_ms", "__start_time", 1000)},
		{"${__to:date:milliseconds}", grafanaTimeValue(context, "__end_time_ms", "__end_time", 1000)},
		{"${__from}", grafanaTimeValue(context, "__start_time_ms", "__start_time", 1000)},
		{"${__to}", grafanaTimeValue(context, "__end_time_ms", "__end_time", 1000)},
		{"$__from", grafanaTimeValue(context, "__start_time_ms", "__start_time", 1000)},
		{"$__to", grafanaTimeValue(context, "__end_time_ms", "__end_time", 1000)},
	}

	for _, replacement := range replacements {
		if replacement.value == "" {
			continue
		}
		query = strings.ReplaceAll(query, replacement.token, replacement.value)
	}

	return query
}

func grafanaTimeValue(context map[string]any, primaryKey string, fallbackKey string, fallbackMultiplier int64) string {
	if value, exists := context[primaryKey]; exists {
		return scalarToString(value)
	}

	value, exists := context[fallbackKey]
	if !exists {
		return ""
	}

	number, ok := scalarToInt64(value)
	if !ok {
		return scalarToString(value)
	}

	return strconv.FormatInt(number*fallbackMultiplier, 10)
}

func isSupportedGrafanaFormat(format string) bool {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "glob", "csv", "raw", "regex", "singlequote", "sqlstring", "json", "doublequote", "pipe":
		return true
	default:
		return false
	}
}

func formatGrafanaValue(value any, format string) string {
	switch strings.ToLower(strings.TrimSpace(format)) {
	case "", "glob":
		return formatGrafanaGlob(value)
	case "csv", "raw":
		return strings.Join(valueStrings(value), ",")
	case "pipe":
		return strings.Join(valueStrings(value), "|")
	case "regex":
		return formatGrafanaRegex(value)
	case "singlequote":
		return quoteGrafanaValues(value, "'", func(s string) string {
			return strings.ReplaceAll(s, "'", `\'`)
		})
	case "sqlstring":
		return quoteGrafanaValues(value, "'", func(s string) string {
			return strings.ReplaceAll(s, "'", "''")
		})
	case "doublequote":
		return quoteGrafanaValues(value, `"`, func(s string) string {
			return strings.ReplaceAll(s, `"`, `\"`)
		})
	case "json":
		return formatGrafanaJSON(value)
	default:
		return scalarToString(value)
	}
}

func formatGrafanaGlob(value any) string {
	values := valueStrings(value)
	if len(values) == 0 {
		return ""
	}
	if len(values) == 1 {
		return values[0]
	}
	return "{" + strings.Join(values, ",") + "}"
}

func formatGrafanaRegex(value any) string {
	values := valueStrings(value)
	if len(values) == 0 {
		return ""
	}

	if len(values) == 1 && strings.Contains(values[0], "|") {
		values = strings.Split(values[0], "|")
	}

	escaped := make([]string, 0, len(values))
	for _, value := range values {
		if value == ".*" {
			escaped = append(escaped, value)
			continue
		}
		escaped = append(escaped, regexp.QuoteMeta(value))
	}

	if len(escaped) == 1 {
		return escaped[0]
	}
	return "(" + strings.Join(escaped, "|") + ")"
}

func quoteGrafanaValues(value any, quote string, escape func(string) string) string {
	values := valueStrings(value)
	quoted := make([]string, 0, len(values))
	for _, value := range values {
		quoted = append(quoted, quote+escape(value)+quote)
	}
	return strings.Join(quoted, ",")
}

func formatGrafanaJSON(value any) string {
	bytes, err := json.Marshal(value)
	if err != nil {
		return scalarToString(value)
	}
	return string(bytes)
}

func valueStrings(value any) []string {
	if value == nil {
		return []string{""}
	}

	rv := reflect.ValueOf(value)
	if rv.Kind() == reflect.Slice || rv.Kind() == reflect.Array {
		if rv.Type().Elem().Kind() == reflect.Uint8 {
			return []string{scalarToString(value)}
		}

		values := make([]string, 0, rv.Len())
		for i := 0; i < rv.Len(); i++ {
			values = append(values, scalarToString(rv.Index(i).Interface()))
		}
		return values
	}

	return []string{scalarToString(value)}
}

func scalarToString(value any) string {
	switch v := value.(type) {
	case nil:
		return ""
	case string:
		return v
	case fmt.Stringer:
		return v.String()
	case bool:
		return strconv.FormatBool(v)
	case int:
		return strconv.FormatInt(int64(v), 10)
	case int8:
		return strconv.FormatInt(int64(v), 10)
	case int16:
		return strconv.FormatInt(int64(v), 10)
	case int32:
		return strconv.FormatInt(int64(v), 10)
	case int64:
		return strconv.FormatInt(v, 10)
	case uint:
		return strconv.FormatUint(uint64(v), 10)
	case uint8:
		return strconv.FormatUint(uint64(v), 10)
	case uint16:
		return strconv.FormatUint(uint64(v), 10)
	case uint32:
		return strconv.FormatUint(uint64(v), 10)
	case uint64:
		return strconv.FormatUint(v, 10)
	case float32:
		return strconv.FormatFloat(float64(v), 'f', -1, 32)
	case float64:
		return strconv.FormatFloat(v, 'f', -1, 64)
	default:
		return fmt.Sprint(v)
	}
}

func scalarToInt64(value any) (int64, bool) {
	switch v := value.(type) {
	case int:
		return int64(v), true
	case int8:
		return int64(v), true
	case int16:
		return int64(v), true
	case int32:
		return int64(v), true
	case int64:
		return v, true
	case uint:
		return int64(v), true
	case uint8:
		return int64(v), true
	case uint16:
		return int64(v), true
	case uint32:
		return int64(v), true
	case uint64:
		if v > uint64(^uint64(0)>>1) {
			return 0, false
		}
		return int64(v), true
	case float32:
		return int64(v), true
	case float64:
		return int64(v), true
	default:
		return 0, false
	}
}
