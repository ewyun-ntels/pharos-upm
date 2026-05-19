package auth

import (
	"time"

	"github.com/ory/fosite"
)

type AccessToken struct {
	token string
	fosite.Requester
}

type RefreshToken struct {
	token string
	fosite.Requester
}

func (r *RefreshToken) isValid() bool {
	if r != nil {
		session := r.GetSession()
		if session != nil {
			exp := session.GetExpiresAt(fosite.RefreshToken).UTC()
			return exp.After(time.Now().UTC())
		}
	}

	return false
}
