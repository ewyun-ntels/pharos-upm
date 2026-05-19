package query

import (
	"crypto/sha256"
	"encoding/hex"
	"log/slog"
	"maps"
	"sort"
	"strings"
	"time"
)

// MergeLabels returns a new map that merges base with all additional label maps.
// Later maps override earlier keys. The base map is not mutated.
func MergeLabels(base map[string]string, adds ...map[string]string) map[string]string {
	// Pre-size capacity hint
	capHint := len(base)
	for _, m := range adds {
		capHint += len(m)
	}
	out := make(map[string]string, capHint)
	maps.Copy(out, base)
	for _, m := range adds {
		maps.Copy(out, m)
	}
	return out
}

// SameTimestamp reports whether two timestamps are equal to the nanosecond.
func SameTimestamp(a, b time.Time) bool {
	return a.Equal(b)
}

// IsAfter reports whether a is strictly after b.
func IsAfter(a, b time.Time) bool {
	return a.After(b)
}

// LabelsFingerprint returns a deterministic fingerprint (sha256 hex) of the labels map.
// Keys are sorted; each entry formatted as k="v" and joined with '\n'.
func LabelsFingerprint(labels map[string]string) string {
	if len(labels) == 0 {
		return ""
	}
	keys := make([]string, 0, len(labels))
	for k := range labels {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	var b strings.Builder
	for i, k := range keys {
		if i > 0 {
			b.WriteByte('\n')
		}
		b.WriteString(k)
		b.WriteString("=")
		// simple quoting to stabilize representation
		b.WriteByte('"')
		b.WriteString(labels[k])
		b.WriteByte('"')
	}
	sum := sha256.Sum256([]byte(b.String()))
	return hex.EncodeToString(sum[:])
}

// LogAttrs returns standardized slog attributes for alert logs.
// Includes: rule, threshold_id, condition, labels_fingerprint.
// Threshold may be nil if not available (e.g., during clear handling); in that case omit it.
func LogAttrs(rule string, th *Threshold, labels map[string]string) []slog.Attr {
	attrs := []slog.Attr{sLogString("rule", rule)}
	// threshold metadata
	if th != nil {
		if th.Id != "" {
			attrs = append(attrs, sLogString("threshold_id", th.Id))
		}
		if th.Operation != "" {
			attrs = append(attrs, sLogString("condition", th.Operation))
		}
	}
	if fp := LabelsFingerprint(labels); fp != "" {
		attrs = append(attrs, sLogString("labels_fingerprint", fp))
	}
	return attrs
}

// AttrsToArgs converts []slog.Attr to []any for slog.* logging calls.
func AttrsToArgs(attrs []slog.Attr) []any {
	if len(attrs) == 0 {
		return nil
	}
	out := make([]any, 0, len(attrs))
	for _, a := range attrs {
		out = append(out, a)
	}
	return out
}

// small wrappers to avoid importing slog at call sites for simple values
func sLogString(k, v string) slog.Attr { return slog.String(k, v) }
