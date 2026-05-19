package query

import "testing"

func TestThreshold_Validate_UnknownOperation(t *testing.T) {
	th := Threshold{Operation: "unknown", StartValue: 1, EndValue: 2}
	if err := th.Validate(); err == nil {
		t.Fatalf("expected error for unknown operation, got nil")
	}
}

func TestThreshold_Validate_InvalidRange(t *testing.T) {
	th := Threshold{Operation: OperationWithinRange, StartValue: 5, EndValue: 3}
	if err := th.Validate(); err == nil {
		t.Fatalf("expected error for invalid range start>end, got nil")
	}
}

func TestThreshold_Validate_ValidRange(t *testing.T) {
	th := Threshold{Operation: OperationWithinRange, StartValue: 3, EndValue: 5}
	if err := th.Validate(); err != nil {
		t.Fatalf("unexpected error for valid range: %v", err)
	}
}

func TestThreshold_Check_IsAbove(t *testing.T) {
	th := Threshold{Operation: OperationIsAbove, StartValue: 10}
	// strictly greater than
	cases := []struct {
		v   float64
		ok  bool
		msg string
	}{
		{10, false, "equal should be false"},
		{9.9999, false, "below should be false"},
		{10.0001, true, "above should be true"},
	}
	for _, c := range cases {
		if got := th.Check(c.v); got != c.ok {
			t.Fatalf("is_above %s: value %v -> %v, want %v", c.msg, c.v, got, c.ok)
		}
	}
}

func TestThreshold_Check_IsBelow(t *testing.T) {
	th := Threshold{Operation: OperationIsBelow, StartValue: 10}
	// strictly less than
	cases := []struct {
		v   float64
		ok  bool
		msg string
	}{
		{10, false, "equal should be false"},
		{10.0001, false, "above should be false"},
		{9.9999, true, "below should be true"},
	}
	for _, c := range cases {
		if got := th.Check(c.v); got != c.ok {
			t.Fatalf("is_below %s: value %v -> %v, want %v", c.msg, c.v, got, c.ok)
		}
	}
}

func TestThreshold_Check_WithinRange_Inclusive(t *testing.T) {
	th := Threshold{Operation: OperationWithinRange, StartValue: 10, EndValue: 20}
	cases := []struct {
		v   float64
		ok  bool
		msg string
	}{
		{9.9999, false, "just below should be false"},
		{10, true, "lower bound inclusive"},
		{15, true, "middle should be true"},
		{20, true, "upper bound inclusive"},
		{20.0001, false, "just above should be false"},
	}
	for _, c := range cases {
		if got := th.Check(c.v); got != c.ok {
			t.Fatalf("within_range %s: value %v -> %v, want %v", c.msg, c.v, got, c.ok)
		}
	}
}

func TestThreshold_Check_OutsideRange(t *testing.T) {
	th := Threshold{Operation: OperationOutsideRange, StartValue: 10, EndValue: 20}
	cases := []struct {
		v   float64
		ok  bool
		msg string
	}{
		{9.9999, true, "just below should be true"},
		{10, false, "lower bound is inside -> false"},
		{15, false, "inside should be false"},
		{20, false, "upper bound is inside -> false"},
		{20.0001, true, "just above should be true"},
	}
	for _, c := range cases {
		if got := th.Check(c.v); got != c.ok {
			t.Fatalf("outside_range %s: value %v -> %v, want %v", c.msg, c.v, got, c.ok)
		}
	}
}
