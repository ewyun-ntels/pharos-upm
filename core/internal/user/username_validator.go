package user

import (
	"errors"
	"fmt"
	"net/mail"
	"regexp"
)

// UsernameValidator validates username complexity.
type UsernameValidator interface {
	Validate(username string) error
}

// usernameValidator is a rule-based implementation of UsernameValidator.
type usernameValidator struct {
	UserEmailAllowed bool
	UserMinLength    int
	UserMaxLength    int
}

func newUsernameValidator(cfg StoreConfig) UsernameValidator {
	minLength := cfg.UserMinLength
	maxLength := cfg.UserMaxLength

	// Set defaults if not provided
	if minLength == 0 {
		minLength = 4
	}
	if maxLength == 0 {
		maxLength = 254
	}

	return &usernameValidator{
		UserEmailAllowed: cfg.UserEmailAllowed,
		UserMinLength:    minLength,
		UserMaxLength:    maxLength,
	}
}

func (v *usernameValidator) Validate(username string) error {
	if v.UserEmailAllowed {
		emailLike := isEmail(username)
		if emailLike {
			return nil
		}
	}

	if len(username) < v.UserMinLength || len(username) > v.UserMaxLength {
		return fmt.Errorf("ID must be between %d and %d characters", v.UserMinLength, v.UserMaxLength)
	}

	matched, err := regexp.MatchString(`^[a-z0-9_]+$`, username)
	if err != nil {
		return err
	}
	if !matched {
		return errors.New("username must contain only lowercase letters, numbers, or underscores")
	}

	return nil
}

func isEmail(s string) bool {
	_, err := mail.ParseAddress(s)
	return err == nil
}
