package auth

import (
	"errors"
	"github.com/ory/fosite"
	"net/http"
)

var (
	ErrNotImplemented = errors.New("not implemented")
)

var (
	ErrInvalidPassword = &fosite.RFC6749Error{
		ErrorField:       "invalid_request",
		DescriptionField: "password mismatch",
		CodeField:        http.StatusBadRequest,
	}
	ErrUserBlocked = &fosite.RFC6749Error{
		ErrorField:       "invalid_request",
		DescriptionField: "user is blocked",
		CodeField:        http.StatusBadRequest,
	}
)
