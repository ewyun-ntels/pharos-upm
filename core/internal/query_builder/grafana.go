package query_builder

import (
	"encoding/json"
	"fmt"
	"log/slog"
	"reflect"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var grafanaVariablePattern = regexp.MustCompile(`\$\{([A-Za-z_][A-Za-z0-9_]*)(?:\.([A-Za-z_][A-Za-z0-9_.]*))?(?::([^}]+))?\}|\[\[([A-Za-z_][A-Za-z0-9_]*)(?::([^\]]+))?\]\]|\$([A-Za-z_][A-Za-z0-9_]*)`)
var promDurationPartPattern = regexp.MustCompile(`^([0-9]+(?:\.[0-9]+)?)(ms|s|m|h|d|w)`)

// Pharos does not yet pass a Prometheus datasource scrape interval, so use the
// 30s Grafana-style panel min step convention as the default fallback.
const defaultGrafanaScrapeInterval = 30 * time.Second

func buildGrafana(query string, queryContext map[string]any) ([]byte, error) {
	convertContext, err := convertInternalContext(queryContext)
	if err != nil {
		return nil, err
	}
	convertContext = enrichGrafanaBuiltins(convertContext)

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

func enrichGrafanaBuiltins(context map[string]any) map[string]any {
	if context == nil {
		context = map[string]any{}
	}

	rateValue, hasRate := context["__rate_interval"]
	rateMSValue, hasRateMS := context["__rate_interval_ms"]

	if hasRate && !hasRateMS {
		if duration, ok := parsePromDuration(rateValue); ok {
			context["__rate_interval_ms"] = duration.Milliseconds()
			hasRateMS = true
		}
	}

	if !hasRate && hasRateMS {
		if duration, ok := parseGrafanaMilliseconds(rateMSValue); ok {
			context["__rate_interval"] = formatPromDuration(duration)
			hasRate = true
		}
	}

	if hasRate && hasRateMS {
		return context
	}

	interval, ok := parsePromDuration(context["__interval"])
	if !ok {
		interval = 0
	}

	scrape, ok := parsePromDuration(context["__scrape_interval"])
	if !ok {
		scrape = defaultGrafanaScrapeInterval
	}

	rate := interval + scrape
	minRate := 4 * scrape
	if rate < minRate {
		rate = minRate
	}

	if !hasRate {
		context["__rate_interval"] = formatPromDuration(rate)
	}
	if !hasRateMS {
		context["__rate_interval_ms"] = rate.Milliseconds()
	}

	return context
}

func parseGrafanaMilliseconds(value any) (time.Duration, bool) {
	milliseconds, ok := scalarToFloat64(value)
	if !ok || milliseconds < 0 {
		return 0, false
	}
	return time.Duration(milliseconds * float64(time.Millisecond)), true
}

func parsePromDuration(value any) (time.Duration, bool) {
	switch v := value.(type) {
	case nil:
		return 0, false
	case time.Duration:
		if v < 0 {
			return 0, false
		}
		return v, true
	case string:
		return parsePromDurationString(v)
	default:
		seconds, ok := scalarToFloat64(value)
		if !ok || seconds < 0 {
			return 0, false
		}
		return time.Duration(seconds * float64(time.Second)), true
	}
}

func parsePromDurationString(value string) (time.Duration, bool) {
	value = strings.TrimSpace(value)
	if value == "" {
		return 0, false
	}

	if seconds, err := strconv.ParseFloat(value, 64); err == nil {
		if seconds < 0 {
			return 0, false
		}
		return time.Duration(seconds * float64(time.Second)), true
	}

	if duration, err := time.ParseDuration(value); err == nil && duration >= 0 {
		return duration, true
	}

	remaining := strings.ToLower(value)
	var total time.Duration
	for remaining != "" {
		match := promDurationPartPattern.FindStringSubmatchIndex(remaining)
		if match == nil || match[0] != 0 {
			return 0, false
		}

		quantity, err := strconv.ParseFloat(remaining[match[2]:match[3]], 64)
		if err != nil || quantity < 0 {
			return 0, false
		}

		unit, ok := promDurationUnit(remaining[match[4]:match[5]])
		if !ok {
			return 0, false
		}

		total += time.Duration(quantity * float64(unit))
		remaining = remaining[match[1]:]
	}

	return total, true
}

func promDurationUnit(unit string) (time.Duration, bool) {
	switch unit {
	case "ms":
		return time.Millisecond, true
	case "s":
		return time.Second, true
	case "m":
		return time.Minute, true
	case "h":
		return time.Hour, true
	case "d":
		return 24 * time.Hour, true
	case "w":
		return 7 * 24 * time.Hour, true
	default:
		return 0, false
	}
}

func formatPromDuration(duration time.Duration) string {
	if duration < time.Second {
		return "1s"
	}

	seconds := int64(duration / time.Second)
	if duration%time.Second != 0 {
		seconds++
	}

	units := []struct {
		suffix  string
		seconds int64
	}{
		{"w", int64((7 * 24 * time.Hour) / time.Second)},
		{"d", int64((24 * time.Hour) / time.Second)},
		{"h", int64(time.Hour / time.Second)},
		{"m", int64(time.Minute / time.Second)},
		{"s", 1},
	}

	var builder strings.Builder
	for _, unit := range units {
		if seconds < unit.seconds {
			continue
		}
		count := seconds / unit.seconds
		seconds %= unit.seconds
		builder.WriteString(strconv.FormatInt(count, 10))
		builder.WriteString(unit.suffix)
	}

	if builder.Len() == 0 {
		return "1s"
	}
	return builder.String()
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

func scalarToFloat64(value any) (float64, bool) {
	switch v := value.(type) {
	case int:
		return float64(v), true
	case int8:
		return float64(v), true
	case int16:
		return float64(v), true
	case int32:
		return float64(v), true
	case int64:
		return float64(v), true
	case uint:
		return float64(v), true
	case uint8:
		return float64(v), true
	case uint16:
		return float64(v), true
	case uint32:
		return float64(v), true
	case uint64:
		return float64(v), true
	case float32:
		return float64(v), true
	case float64:
		return v, true
	case json.Number:
		number, err := v.Float64()
		return number, err == nil
	case string:
		number, err := strconv.ParseFloat(strings.TrimSpace(v), 64)
		return number, err == nil
	default:
		return 0, false
	}
}
