package store

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"

	"github.com/golang-jwt/jwt/v5"
)

// getJwtSubject extracts subject from Authorization header
// Supports both Bearer JWT tokens and Basic Auth for OAuth client authentication
func getJwtSubject(header http.Header) (string, error) {
	auth := header.Get("Authorization")
	if auth == "" {
		return "", nil
	}

	parts := strings.Split(auth, " ")
	if len(parts) != 2 {
		return "", errors.New("invalid Authorization header format")
	}

	authType := strings.ToLower(parts[0])
	authValue := parts[1]

	switch authType {
	case "bearer":
		// JWT Bearer token - extract user subject
		return extractJWTSubject(authValue)
	case "basic":
		// Basic Auth - but we don't consider client_id as user for logging
		// Return empty string to indicate no user found
		return "", nil
	default:
		return "", errors.New("unsupported authorization type: " + authType)
	}
}

// extractJWTSubject extracts subject from JWT token
func extractJWTSubject(token string) (string, error) {
	jwtP := jwt.Parser{}
	parsedToken, _, err := jwtP.ParseUnverified(token, jwt.MapClaims{})
	if err != nil {
		return "", err
	}

	return parsedToken.Claims.GetSubject()
}

// extractBasicAuthClientId extracts client_id from Basic Auth header
// Used for OAuth client authentication in refresh token grants
// Basic Auth is treated as no user for logging purposes, but this function can be used to extract the client_id if needed.
func extractBasicAuthClientId(encoded string) (string, error) {
	decoded, err := base64.StdEncoding.DecodeString(encoded)
	if err != nil {
		return "", errors.New("invalid base64 encoding in Basic Auth")
	}

	credentials := strings.SplitN(string(decoded), ":", 2)
	if len(credentials) != 2 {
		return "", errors.New("invalid Basic Auth format, expected client_id:client_secret")
	}

	clientId := credentials[0]
	if clientId == "" {
		return "", errors.New("client_id is empty")
	}

	// For user registration purposes, use client_id as the identifier
	// You might want to prefix it to distinguish from regular users
	return "client:" + clientId, nil
}
