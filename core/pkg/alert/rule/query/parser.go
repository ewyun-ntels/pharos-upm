package query

import (
	"errors"
	"fmt"
	"strconv"
	"time"
)

// ParseRow parses a generic row coming from datasource into:
// - timestamp (using timeLabel)
// - numeric value (using variableLabel)
// - baseLabels (all other columns stringified)
// It standardizes error messages for consistency across callers.
func ParseRow(row map[string]any, timeLabel, variableLabel string) (time.Time, float64, map[string]string, error) {
	// Parse time
	timeRaw, ok := row[timeLabel]
	if !ok {
		return time.Time{}, 0, nil, errors.New("missing time label")
	}

	var ts time.Time
	switch v := timeRaw.(type) {
	case time.Time:
		ts = v
	case string:
		parsedTime, err := time.Parse(time.DateTime, v)
		if err != nil {
			return time.Time{}, 0, nil, fmt.Errorf("invalid time format: %w", err)
		}
		ts = parsedTime
	default:
		return time.Time{}, 0, nil, fmt.Errorf("time label has unsupported type: %T", v)
	}

	// Parse value
	varRaw, ok := row[variableLabel]
	if !ok {
		return time.Time{}, 0, nil, errors.New("missing variable label")
	}

	var decVal float64
	switch v := varRaw.(type) {
	case float64:
		decVal = v
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return time.Time{}, 0, nil, fmt.Errorf("failed to parse float: %w", err)
		}
		decVal = parsed
	default:
		// model.SampleValue and other numeric types
		parsed, err := strconv.ParseFloat(fmt.Sprint(v), 64)
		if err != nil {
			return time.Time{}, 0, nil, fmt.Errorf("variable label has unsupported type %T: %w", v, err)
		}
		decVal = parsed
	}

	// Base labels from row (pre-size capacity to reduce reallocations)
	capHint := max(len(row)-2, 0)
	baseLabels := make(map[string]string, capHint)
	for key, value := range row {
		if key != timeLabel && key != variableLabel {
			baseLabels[key] = fmt.Sprint(value)
		}
	}

	return ts, decVal, baseLabels, nil
}
