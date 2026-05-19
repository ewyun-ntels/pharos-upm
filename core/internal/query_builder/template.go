package query_builder

import (
	"fmt"
	"log/slog"
	"regexp"
	"strconv"
	"strings"
	"time"

	"github.com/flosch/pongo2/v6"
	"github.com/jinzhu/copier"
)

func parseExpression(input string) (int64, string, error) {
	re := regexp.MustCompile(`^(\d+)?\s*(.*)$`)

	input = strings.TrimSpace(input)
	matches := re.FindStringSubmatch(input)

	if len(matches) != 3 {
		return 0, "", fmt.Errorf("invalid format")
	}

	unit := matches[2]
	if unit == "Now" {
		return 0, "Now", nil
	}

	if matches[1] == "" {
		return 0, "", fmt.Errorf("missing number")
	}

	number, err := strconv.ParseInt(matches[1], 10, 64)
	if err != nil {
		return 0, "", fmt.Errorf("invalid number")
	}

	return number, unit, nil
}

func calculateUnixTimestamp(input string) (int64, error) {
	number, unit, err := parseExpression(input)
	if err != nil {
		return 0, err
	}

	now := time.Now()
	var targetTime time.Time

	switch unit {
	case "Now":
		targetTime = now
	case "Seconds ago":
		targetTime = now.Add(time.Duration(-number) * time.Second)
	case "Minutes ago":
		targetTime = now.Add(time.Duration(-number) * time.Minute)
	case "Hours ago":
		targetTime = now.Add(time.Duration(-number) * time.Hour)
	case "Days ago":
		targetTime = now.AddDate(0, 0, int(-number))
	case "Weeks ago":
		targetTime = now.AddDate(0, 0, int(-number*7))
	case "Months ago":
		targetTime = now.AddDate(0, int(-number), 0)
	case "Years ago":
		targetTime = now.AddDate(int(-number), 0, 0)

	case "Seconds from now":
		targetTime = now.Add(time.Duration(number) * time.Second)
	case "Minutes from now":
		targetTime = now.Add(time.Duration(number) * time.Minute)
	case "Hours from now":
		targetTime = now.Add(time.Duration(number) * time.Hour)
	case "Days from now":
		targetTime = now.AddDate(0, 0, int(number))
	case "Weeks from now":
		targetTime = now.AddDate(0, 0, int(number*7))
	case "Months from now":
		targetTime = now.AddDate(0, int(number), 0)
	case "Years from now":
		targetTime = now.AddDate(int(number), 0, 0)
	default:
		return 0, fmt.Errorf("invalid time unit")
	}

	return targetTime.Unix(), nil
}

func parseAndSetTime(context map[string]any, key string) {
	if value, exist := context[key]; exist {
		switch v := value.(type) {
		case string:
			parsedTime, err := calculateUnixTimestamp(v)
			if err != nil {
				slog.Error("calculateUnixTimestamp error",
					slog.String("key", key), slog.Any("value", value))
			}

			context[key] = parsedTime

		case int:
			context[key] = int64(v)
		case int32:
			context[key] = int64(v)
		case int64:
			// already correct
		case float32:
			context[key] = int64(v)
		case float64:
			// JSON number는 float64로 파싱되므로 int64로 변환
			context[key] = int64(v)

		default:
			slog.Warn("Skipping invalid time expression",
				slog.String("key", key), slog.Any("value", value))
		}
	}
}

func parseAndSetTimeMs(context map[string]any, key string) {
	if value, exist := context[key]; exist {
		switch v := value.(type) {
		case string:
			// 문자열이면 파싱 후 밀리초로 변환
			parsedTime, err := calculateUnixTimestamp(v)
			if err != nil {
				slog.Error("calculateUnixTimestamp error",
					slog.String("key", key), slog.Any("value", value))
			}

			context[key] = parsedTime * 1000 // 초 → 밀리초

		case int, int32, int64, float32, float64:
			// 숫자면 이미 밀리초라고 가정하고 그대로 유지
			// (프론트엔드에서 밀리초로 보낸 경우)

		default:
			slog.Warn("Skipping invalid time expression",
				slog.String("key", key), slog.Any("value", value))
		}
	}
}

func convertInternalContext(context map[string]any) (map[string]any, error) {
	copyContext := make(map[string]any, len(context))
	err := copier.Copy(&copyContext, context)
	if err != nil {
		return nil, err
	}

	parseAndSetTime(copyContext, "__start_time")
	parseAndSetTime(copyContext, "__end_time")
	parseAndSetTimeMs(copyContext, "__start_time_ms")
	parseAndSetTimeMs(copyContext, "__end_time_ms")

	return copyContext, nil
}

func GetQuery(query string, queryContext map[string]any) ([]byte, error) {
	convertContext, err := convertInternalContext(queryContext)
	if err != nil {
		return nil, err
	}

	tpl, err := pongo2.FromString(query)
	if err != nil {
		return nil, err
	}

	out, err := tpl.Execute(convertContext)
	if err != nil {
		return nil, err
	}

	return []byte(out), nil
}
