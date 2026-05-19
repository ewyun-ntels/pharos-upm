package user

import (
	"testing"
)

func TestRuleBasedPasswordValidator_Success(t *testing.T) {
	t.Parallel()

	v := &RuleBasedPasswordValidator{
		MinLength:    8,
		MinUppercase: 1,
		MinLowercase: 1,
		MinDigits:    1,
		MinSpecial:   1,
	}

	// Meets all requirements: length 10, 1 upper, 1 lower, 1 digit, 1 special
	pwd := "Aa1!aaaaaa"
	if err := v.Validate(pwd); err != nil {
		t.Fatalf("expected success, got error: %v", err)
	}
}

func TestRuleBasedPasswordValidator_LengthFailure(t *testing.T) {
	t.Parallel()

	v := &RuleBasedPasswordValidator{MinLength: 10}
	err := v.Validate("short")
	if err == nil {
		t.Fatalf("expected error for short password, got nil")
	}
	want := "password must be at least 10 characters long"
	if err.Error() != want {
		t.Fatalf("unexpected error: want %q got %q", want, err.Error())
	}
}

func TestRuleBasedPasswordValidator_ClassFailures(t *testing.T) {
	t.Parallel()

	// Require at least one of each class
	v := &RuleBasedPasswordValidator{
		MinLength:    4,
		MinUppercase: 1,
		MinLowercase: 1,
		MinDigits:    1,
		MinSpecial:   1,
	}

	// No uppercase
	if err := v.Validate("a1!a"); err == nil || err.Error() != "password must contain at least 1 uppercase letter(s)" {
		t.Fatalf("unexpected error for missing uppercase: %v", err)
	}

	// No lowercase
	if err := v.Validate("A1!A"); err == nil || err.Error() != "password must contain at least 1 lowercase letter(s)" {
		t.Fatalf("unexpected error for missing lowercase: %v", err)
	}

	// No digit
	if err := v.Validate("Aa!A"); err == nil || err.Error() != "password must contain at least 1 digit(s)" {
		t.Fatalf("unexpected error for missing digit: %v", err)
	}

	// No special
	if err := v.Validate("Aa1A"); err == nil || err.Error() != "password must contain at least 1 special character(s)" {
		t.Fatalf("unexpected error for missing special: %v", err)
	}
}

func TestRuleBasedPasswordValidator_SpecialCharsVariety(t *testing.T) {
	t.Parallel()

	v := &RuleBasedPasswordValidator{
		MinLength:    4,
		MinUppercase: 0,
		MinLowercase: 0,
		MinDigits:    0,
		MinSpecial:   2, // require two specials to test counting
	}

	// Contains two different specials from the defined class: ! and @
	if err := v.Validate("!!@@"); err != nil {
		t.Fatalf("expected success for two specials, got: %v", err)
	}
}

func TestRuleBasedPasswordValidator_BoundaryCounts(t *testing.T) {
	t.Parallel()

	// Exactly meets each minimum
	v := &RuleBasedPasswordValidator{
		MinLength:    5,
		MinUppercase: 1,
		MinLowercase: 1,
		MinDigits:    1,
		MinSpecial:   1,
	}

	// 1 upper (A), 1 lower (a), 1 digit(1), 1 special(!), length 5
	if err := v.Validate("Aa1!!"); err != nil {
		t.Fatalf("expected boundary success, got: %v", err)
	}
}

func TestRuleBasedPasswordValidator_NilReceiverIsNoop(t *testing.T) {
	t.Parallel()

	var v *RuleBasedPasswordValidator
	if err := v.Validate("anything"); err != nil {
		t.Fatalf("nil validator should accept any password, got: %v", err)
	}
}
