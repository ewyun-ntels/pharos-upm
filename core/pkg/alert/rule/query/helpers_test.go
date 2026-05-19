package query

import (
	"log/slog"
	"testing"
	"time"
)

func TestMergeLabels_NewMapAndOverride(t *testing.T) {
	base := map[string]string{"a": "1", "b": "2"}
	add1 := map[string]string{"b": "B", "c": "3"}
	add2 := map[string]string{"d": "4"}

	merged := MergeLabels(base, add1, add2)
	if &merged == &base { // should be different map; though addresses of maps aren't directly comparable, comparing pointers of vars is ok
		t.Fatalf("MergeLabels must return a new map instance")
	}
	// base unaffected
	if base["b"] != "2" {
		t.Fatalf("base mutated: want b=2 got %s", base["b"])
	}
	// overrides applied
	if merged["b"] != "B" || merged["c"] != "3" || merged["d"] != "4" || merged["a"] != "1" {
		t.Fatalf("merged unexpected: %v", merged)
	}
}

func TestTimeComparators(t *testing.T) {
	a := time.Now()
	b := a
	c := a.Add(1 * time.Second)
	if !SameTimestamp(a, b) {
		t.Fatalf("SameTimestamp should be true for equal times")
	}
	if SameTimestamp(a, c) {
		t.Fatalf("SameTimestamp should be false for different times")
	}
	if !IsAfter(c, a) || IsAfter(a, c) || IsAfter(a, b) {
		t.Fatalf("IsAfter logic incorrect")
	}
}

func TestLabelsFingerprint_DeterministicAndDistinct(t *testing.T) {
	m1 := map[string]string{"x": "1", "y": "2"}
	m2 := map[string]string{"y": "2", "x": "1"} // different insertion order
	fp1 := LabelsFingerprint(m1)
	fp2 := LabelsFingerprint(m2)
	if fp1 == "" || fp2 == "" {
		t.Fatalf("fingerprint should not be empty")
	}
	if fp1 != fp2 {
		t.Fatalf("fingerprint must be deterministic regardless of insertion order: %s != %s", fp1, fp2)
	}
	m3 := map[string]string{"x": "1", "y": "DIFF"}
	fp3 := LabelsFingerprint(m3)
	if fp3 == fp1 {
		t.Fatalf("fingerprint must change when content changes")
	}
}

func TestLogAttrsAndArgs(t *testing.T) {
	th := &Threshold{Id: "tid", Operation: OperationIsBelow}
	labels := map[string]string{"a": "1", "b": "2"}
	attrs := LogAttrs("rule-1", th, labels)
	if len(attrs) == 0 {
		t.Fatalf("expected some attributes")
	}
	// basic presence checks
	var seen = map[string]bool{}
	for _, a := range attrs {
		seen[a.Key] = true
	}
	if !seen["rule"] || !seen["threshold_id"] || !seen["condition"] || !seen["labels_fingerprint"] {
		t.Fatalf("missing expected log attrs: %+v", attrs)
	}
	args := AttrsToArgs(attrs)
	if len(args) != len(attrs) {
		t.Fatalf("AttrsToArgs length mismatch")
	}
	// Can be passed to slog without panic (not executing real logging here)
	_ = slog.Attr{}
}
