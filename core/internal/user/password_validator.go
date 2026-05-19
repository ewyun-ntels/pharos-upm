package user

import (
	"fmt"
	"regexp"
)

// PasswordValidator abstracts the password complexity validation.
// This allows injecting alternative policies in tests or other deployments
// without changing the store implementation.
type PasswordValidator interface {
	Validate(password string) error
}

// RuleBasedPasswordValidator validates password complexity using simple counters
// for length, uppercase, lowercase, digits, and special characters.
type RuleBasedPasswordValidator struct {
	MinLength    int
	MinUppercase int
	MinLowercase int
	MinDigits    int
	MinSpecial   int
}

func newPasswordValidator(cfg StoreConfig) PasswordValidator {
	length := cfg.PasswordMinLength
	uppercase := cfg.PasswordMinUppercase
	lowercase := cfg.PasswordMinLowercase
	digits := cfg.PasswordMinDigits
	special := cfg.PasswordMinSpecial
	if length == 0 {
		length = 8
	}

	return &RuleBasedPasswordValidator{
		MinLength:    length,
		MinUppercase: uppercase,
		MinLowercase: lowercase,
		MinDigits:    digits,
		MinSpecial:   special,
	}
}

func (v *RuleBasedPasswordValidator) Validate(password string) error {
	if v == nil {
		return nil
	}
	if len(password) < v.MinLength {
		return fmt.Errorf("password must be at least %d characters long", v.MinLength)
	}

	upper := regexp.MustCompile(`[A-Z]`)
	lower := regexp.MustCompile(`[a-z]`)
	digits := regexp.MustCompile(`[0-9]`)
	special := regexp.MustCompile(`[!@#~$%^&*()_+\-=\[\]{};':"\\|,.<>/?]`)

	countUpper := len(upper.FindAllString(password, -1))
	countLower := len(lower.FindAllString(password, -1))
	countDigit := len(digits.FindAllString(password, -1))
	countSpecial := len(special.FindAllString(password, -1))

	if countUpper < v.MinUppercase {
		return fmt.Errorf("password must contain at least %d uppercase letter(s)", v.MinUppercase)
	}
	if countLower < v.MinLowercase {
		return fmt.Errorf("password must contain at least %d lowercase letter(s)", v.MinLowercase)
	}
	if countDigit < v.MinDigits {
		return fmt.Errorf("password must contain at least %d digit(s)", v.MinDigits)
	}
	if countSpecial < v.MinSpecial {
		return fmt.Errorf("password must contain at least %d special character(s)", v.MinSpecial)
	}

	return nil
}
