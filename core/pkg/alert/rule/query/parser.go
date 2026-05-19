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
	if timeStr, ok := timeRaw.(string); ok {
		parsedTime, err := time.Parse(time.DateTime, timeStr)
		if err != nil {
			return time.Time{}, 0, nil, fmt.Errorf("invalid time format: %w", err)
		}
		ts = parsedTime
	} else {
		return time.Time{}, 0, nil, errors.New("time label is not a string")
	}

	// Parse value
	varRaw, ok := row[variableLabel]
	if !ok {
		return time.Time{}, 0, nil, errors.New("missing variable label")
	}

	varStr, ok := varRaw.(string)
	if !ok {
		return time.Time{}, 0, nil, errors.New("variable label is not a string")
	}
	decVal, err := strconv.ParseFloat(varStr, 64)
	if err != nil {
		return time.Time{}, 0, nil, fmt.Errorf("failed to parse float: %w", err)
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
