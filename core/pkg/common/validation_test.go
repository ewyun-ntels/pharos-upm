package common

import (
	"strings"
	"testing"
	"time"
)

func TestValidateUUID(t *testing.T) {
	v := NewValidator()

	// optional empty
	got := v.ValidateUUID("id", "", false)
	if got != "" || v.HasErrors() {
		t.Fatalf("optional empty should return empty with no errors; got=%q errors=%v", got, v.GetErrors())
	}

	// required empty
	v = NewValidator()
	got = v.ValidateUUID("id", "", true)
	if got != "" || !v.HasErrors() {
		t.Fatalf("required empty should set error; got=%q errors=%v", got, v.GetErrors())
	}

	// invalid UUID
	v = NewValidator()
	got = v.ValidateUUID("id", "not-a-uuid", true)
	if got != "" || !v.HasErrors() {
		t.Fatalf("invalid uuid should set error; got=%q errors=%v", got, v.GetErrors())
	}

	// valid UUID
	v = NewValidator()
	valid := "123e4567-e89b-12d3-a456-426614174000"
	got = v.ValidateUUID("id", valid, true)
	if got != valid || v.HasErrors() {
		t.Fatalf("valid uuid should pass; got=%q errors=%v", got, v.GetErrors())
	}
}

func TestValidateSimpleID(t *testing.T) {
	v := NewValidator()
	if got := v.ValidateSimpleID("sid", "abc_123-XYZ", true); got != "abc_123-XYZ" || v.HasErrors() {
		t.Fatalf("valid simple id failed: got=%q errors=%v", got, v.GetErrors())
	}

	v = NewValidator()
	if got := v.ValidateSimpleID("sid", "bad space", true); got != "" || !v.HasErrors() {
		t.Fatalf("invalid simple id should error: got=%q errors=%v", got, v.GetErrors())
	}
}

func TestValidateAlertName(t *testing.T) {
	v := NewValidator()
	in := "  Hello   World  "
	got := v.ValidateAlertName("name", in, true)
	if got != "Hello World" || v.HasErrors() {
		t.Fatalf("alert name should be sanitized and valid: got=%q errors=%v", got, v.GetErrors())
	}

	// too long
	v = NewValidator()
	long := strings.Repeat("a", 101)
	if got = v.ValidateAlertName("name", long, true); got != "" || !v.HasErrors() {
		t.Fatalf("too long alert name should error: got=%q errors=%v", got, v.GetErrors())
	}

	// invalid chars
	v = NewValidator()
	if got = v.ValidateAlertName("name", "bad@name", true); got != "" || !v.HasErrors() {
		t.Fatalf("invalid chars should error: got=%q errors=%v", got, v.GetErrors())
	}
}

func TestValidateResourceName(t *testing.T) {
	v := NewValidator()
	if got := v.ValidateResourceName("res", "valid_name-1", true); got != "valid_name-1" || v.HasErrors() {
		t.Fatalf("valid resource name failed: got=%q errors=%v", got, v.GetErrors())
	}

	v = NewValidator()
	if got := v.ValidateResourceName("res", "has.dot", true); got != "" || !v.HasErrors() {
		t.Fatalf("dot not allowed: got=%q errors=%v", got, v.GetErrors())
	}
}

func TestValidateFileName(t *testing.T) {
	v := NewValidator()
	if got := v.ValidateFileName("file", "good-file_name.txt", true); got != "good-file_name.txt" || v.HasErrors() {
		t.Fatalf("valid file name failed: got=%q errors=%v", got, v.GetErrors())
	}

	// dangerous patterns
	for _, s := range []string{"../secret", "a/b", "a\\b", ":bad"} {
		v = NewValidator()
		if got := v.ValidateFileName("file", s, true); got != "" || !v.HasErrors() {
			t.Fatalf("dangerous file name %q should error: got=%q errors=%v", s, got, v.GetErrors())
		}
	}
}

func TestValidateHostname(t *testing.T) {
	v := NewValidator()
	// valid IP
	if got := v.ValidateHostname("host", "192.168.0.1", true); got != "192.168.0.1" || v.HasErrors() {
		t.Fatalf("valid IP should pass: got=%q errors=%v", got, v.GetErrors())
	}

	// valid hostname (lowercased)
	v = NewValidator()
	if got := v.ValidateHostname("host", "My-Host.Example", true); got != "my-host.example" || v.HasErrors() {
		t.Fatalf("valid hostname should pass and lowercase: got=%q errors=%v", got, v.GetErrors())
	}

	// invalid hostname
	v = NewValidator()
	if got := v.ValidateHostname("host", "-bad-", true); got != "" || !v.HasErrors() {
		t.Fatalf("invalid hostname should error: got=%q errors=%v", got, v.GetErrors())
	}

	// too long (>253)
	v = NewValidator()
	tooLong := strings.Repeat("a", 254)
	if got := v.ValidateHostname("host", tooLong, true); got != "" || !v.HasErrors() {
		t.Fatalf("too long hostname should error: got=%q errors=%v", got, v.GetErrors())
	}
}

func TestValidateDateTime(t *testing.T) {
	v := NewValidator()
	// required empty
	zero := v.ValidateDateTime("ts", "", true)
	if !zero.IsZero() && !v.HasErrors() {
		t.Fatalf("required empty should set error")
	}

	// valid formats
	v = NewValidator()
	if ts := v.ValidateDateTime("ts", time.RFC3339, true); !ts.IsZero() && v.HasErrors() {
		t.Fatalf("using format string shouldn't be parsed; ensure we pass actual time string")
	}
	v = NewValidator()
	nowStr := time.Now().UTC().Format(time.RFC3339)
	if ts := v.ValidateDateTime("ts", nowStr, true); ts.IsZero() || v.HasErrors() {
		t.Fatalf("valid RFC3339 should parse: ts=%v errors=%v", ts, v.GetErrors())
	}

	// invalid
	v = NewValidator()
	if ts := v.ValidateDateTime("ts", "2020/01/01 12:00:00", true); !ts.IsZero() || !v.HasErrors() {
		t.Fatalf("invalid datetime should error")
	}
}

func TestValidateCount(t *testing.T) {
	v := NewValidator()
	if n := v.ValidateCount("cnt", "", 1, 10, true); n != 0 || !v.HasErrors() {
		t.Fatalf("required empty count should error")
	}

	v = NewValidator()
	if n := v.ValidateCount("cnt", "abc", 1, 10, true); n != 0 || !v.HasErrors() {
		t.Fatalf("non-integer should error")
	}

	v = NewValidator()
	if n := v.ValidateCount("cnt", "0", 1, 10, true); n != 0 || !v.HasErrors() {
		t.Fatalf("below min should error")
	}

	v = NewValidator()
	if n := v.ValidateCount("cnt", "11", 1, 10, true); n != 0 || !v.HasErrors() {
		t.Fatalf("above max should error")
	}

	v = NewValidator()
	if n := v.ValidateCount("cnt", "5", 1, 10, true); n != 5 || v.HasErrors() {
		t.Fatalf("within range should pass: n=%d errors=%v", n, v.GetErrors())
	}
}

func TestValidateEnum(t *testing.T) {
	v := NewValidator()
	allowed := []string{"a", "b", "c"}

	// optional empty ok
	if got := v.ValidateEnum("e", "", allowed, false); got != "" || v.HasErrors() {
		t.Fatalf("optional empty enum should be ok")
	}

	v = NewValidator()
	if got := v.ValidateEnum("e", "d", allowed, true); got != "" || !v.HasErrors() {
		t.Fatalf("invalid enum should error")
	}

	v = NewValidator()
	if got := v.ValidateEnum("e", "b", allowed, true); got != "b" || v.HasErrors() {
		t.Fatalf("valid enum should pass")
	}
}

func TestValidateJSONString(t *testing.T) {
	v := NewValidator()
	if got := v.ValidateJSONString("json", "", false); got != "" || v.HasErrors() {
		t.Fatalf("optional empty json should be ok")
	}

	v = NewValidator()
	if got := v.ValidateJSONString("json", "{bad}", true); got != "" || !v.HasErrors() {
		t.Fatalf("invalid json should error")
	}

	v = NewValidator()
	if got := v.ValidateJSONString("json", "{\"a\":1}", true); got != "{\"a\":1}" || v.HasErrors() {
		t.Fatalf("valid json should pass: got=%q errors=%v", got, v.GetErrors())
	}

	// too large (>1MB)
	v = NewValidator()
	big := "{" + "\"k\":" + "\"" + strings.Repeat("x", 1024*1024) + "\"}"
	if got := v.ValidateJSONString("json", big, true); got != "" || !v.HasErrors() {
		t.Fatalf("too large json should error")
	}
}

func TestValidateStringLength(t *testing.T) {
	v := NewValidator()
	if got := v.ValidateStringLength("s", "", 1, 5, true); got != "" || !v.HasErrors() {
		t.Fatalf("required empty should error")
	}

	v = NewValidator()
	if got := v.ValidateStringLength("s", "ab", 3, 5, true); got != "" || !v.HasErrors() {
		t.Fatalf("below min should error")
	}

	v = NewValidator()
	if got := v.ValidateStringLength("s", "abcdef", 1, 5, true); got != "" || !v.HasErrors() {
		t.Fatalf("above max should error")
	}

	v = NewValidator()
	if got := v.ValidateStringLength("s", "a\x01b\n\t", 1, 10, true); got != "ab\n\t" || v.HasErrors() {
		t.Fatalf("control chars should be stripped except newline/tab: got=%q errors=%v", got, v.GetErrors())
	}
}

func TestValidateQueryParam(t *testing.T) {
	v := NewValidator()
	// optional empty ok
	if got := v.ValidateQueryParam("q", "", false); got != "" || v.HasErrors() {
		t.Fatalf("optional empty query should be ok")
	}

	// URL decode
	v = NewValidator()
	if got := v.ValidateQueryParam("q", "hello%20world", true); got != "hello world" || v.HasErrors() {
		t.Fatalf("url decoded value expected; got=%q errors=%v", got, v.GetErrors())
	}

	// invalid encoding
	v = NewValidator()
	if got := v.ValidateQueryParam("q", "%ZZ", true); got != "" || !v.HasErrors() {
		t.Fatalf("invalid url encoding should error")
	}

	// invalid characters (e.g., semicolon not in whitelist)
	v = NewValidator()
	if got := v.ValidateQueryParam("q", "bad;drop", true); got != "" || !v.HasErrors() {
		t.Fatalf("invalid characters should error")
	}
}
