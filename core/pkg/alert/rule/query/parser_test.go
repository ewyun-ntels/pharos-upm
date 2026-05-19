package query

import (
	"strings"
	"testing"
	"time"
)

func TestParseRow_Success(t *testing.T) {
	timeLabel := "ts"
	varLabel := "val"
	ts := time.Now().Format(time.DateTime)
	row := map[string]any{
		timeLabel: ts,
		varLabel:  "123.45",
		"host":    "srv1",
		"zone":    3,       // non-string should be stringified
		"active":  true,    // non-string should be stringified
		"note":    "hello", // extra string
	}

	parsedTs, value, labels, err := ParseRow(row, timeLabel, varLabel)
	if err != nil {
		t.Fatalf("ParseRow success expected, got error: %v", err)
	}
	if parsedTs.IsZero() {
		t.Fatalf("parsed timestamp should not be zero")
	}
	if value != 123.45 {
		t.Fatalf("expected value 123.45, got %v", value)
	}
	// Ensure base labels include all except time and variable labels
	if labels["host"] != "srv1" {
		t.Fatalf("expected host label, got: %v", labels["host"])
	}
	if labels["zone"] != "3" { // stringified
		t.Fatalf("expected zone=3 stringified, got: %v", labels["zone"])
	}
	if labels["active"] != "true" {
		t.Fatalf("expected active=true stringified, got: %v", labels["active"])
	}
	if _, ok := labels[timeLabel]; ok {
		t.Fatalf("time label should not be included in base labels: %+v", labels)
	}
	if _, ok := labels[varLabel]; ok {
		t.Fatalf("variable label should not be included in base labels: %+v", labels)
	}
}

func TestParseRow_Errors(t *testing.T) {
	timeLabel := "ts"
	varLabel := "val"
	ts := time.Now().Format(time.DateTime)

	// Missing time label
	_, _, _, err := ParseRow(map[string]any{varLabel: "1"}, timeLabel, varLabel)
	if err == nil || err.Error() != "missing time label" {
		t.Fatalf("expected missing time label, got %v", err)
	}

	// Time not a string
	_, _, _, err = ParseRow(map[string]any{timeLabel: 123, varLabel: "1"}, timeLabel, varLabel)
	if err == nil || err.Error() != "time label is not a string" {
		t.Fatalf("expected time label is not a string, got %v", err)
	}

	// Invalid time format
	_, _, _, err = ParseRow(map[string]any{timeLabel: "2020-01-01T00:00:00Z", varLabel: "1"}, timeLabel, varLabel)
	if err == nil || !strings.HasPrefix(err.Error(), "invalid time format") {
		t.Fatalf("expected invalid time format, got %v", err)
	}

	// Missing variable label
	_, _, _, err = ParseRow(map[string]any{timeLabel: ts}, timeLabel, varLabel)
	if err == nil || err.Error() != "missing variable label" {
		t.Fatalf("expected missing variable label, got %v", err)
	}

	// Variable not a string
	_, _, _, err = ParseRow(map[string]any{timeLabel: ts, varLabel: 1}, timeLabel, varLabel)
	if err == nil || err.Error() != "variable label is not a string" {
		t.Fatalf("expected variable label is not a string, got %v", err)
	}

	// Variable not parseable as float
	_, _, _, err = ParseRow(map[string]any{timeLabel: ts, varLabel: "abc"}, timeLabel, varLabel)
	if err == nil || err.Error()[:21] != "failed to parse float" {
		t.Fatalf("expected failed to parse float, got %v", err)
	}
}
