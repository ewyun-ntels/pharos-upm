package authhandler

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestExtractToken(t *testing.T) {
	t.Run("should extract token from Authorization header", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer header-token")

		token := extractToken(req)
		assert.Equal(t, "header-token", token)
	})

	t.Run("should extract token from cookie if header is missing", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.AddCookie(&http.Cookie{Name: "access_token", Value: "cookie-token"})

		token := extractToken(req)
		assert.Equal(t, "cookie-token", token)
	})

	t.Run("should prioritize header over cookie", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)
		req.Header.Set("Authorization", "Bearer header-token")
		req.AddCookie(&http.Cookie{Name: "access_token", Value: "cookie-token"})

		token := extractToken(req)
		assert.Equal(t, "header-token", token)
	})

	t.Run("should return empty string if both are missing", func(t *testing.T) {
		req, _ := http.NewRequest("GET", "/", nil)

		token := extractToken(req)
		assert.Equal(t, "", token)
	})
}
