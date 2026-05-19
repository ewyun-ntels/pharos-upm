package authhandler

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite"
	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/pkg/common"
	auth_types "ntels.com/pharos/shared/types/auth"
)

func TestGetConfigInfoHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	config := common.Config{
		User: common.UsersConfig{
			LoginBlockDuration:     1800,
			LoginRetryResetTimeout: 300,
			PasswordRule: common.PasswordRuleConfig{
				MinDigits:    2,
				MinLength:    10,
				MinLowercase: 1,
				MinSpecial:   1,
				MinUppercase: 1,
				PasswordTTL:  "30d",
			},
			UserRule: common.UserRuleConfig{
				EmailAllowed: false,
				MaxLength:    30,
				MinLength:    5,
			},
		},
	}

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request, _ = http.NewRequest("GET", "/auth/config", nil)

	handler := GetConfigInfoHandler(config)
	handler(c)

	assert.Equal(t, http.StatusOK, w.Code)

	result, err := auth_types.UnmarshalAuthConfig(w.Body.Bytes())
	if err != nil {
		t.Fatalf("Failed to unmarshal response: %v", err)
	}

	assert.Equal(t, int64(1800), result.Common.LoginBlockDuration)
	assert.Equal(t, int64(300), result.Common.LoginRetryResetTimeout)
	assert.Equal(t, int64(2), result.Password.MinDigits)
	assert.Equal(t, int64(10), result.Password.MinLength)
	assert.Equal(t, false, result.User.EmailAllowed)
	assert.Equal(t, int64(30), result.User.MaxLength)
	// 30 days = 30 * 24 * 3600 = 2592000 seconds
	assert.Equal(t, int64(2592000), *result.Password.PasswordTTL)
}

func TestSetTokenCookie(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name            string
		config          common.Config
		tokenResponse   *fosite.AccessResponse
		expectedCookies []string
	}{
		{
			name: "should set access_token cookie",
			config: common.Config{
				Auth: common.AuthConfig{
					AccessTokenLifespan: 1 * time.Hour,
					SecureCookies:       true,
				},
			},
			tokenResponse: &fosite.AccessResponse{
				AccessToken: "test-access-token",
			},
			expectedCookies: []string{"access_token=test-access-token"},
		},
		{
			name: "should set both access and refresh token cookies",
			config: common.Config{
				Auth: common.AuthConfig{
					AccessTokenLifespan:  1 * time.Hour,
					RefreshTokenLifespan: 24 * time.Hour,
					SecureCookies:        false,
				},
			},
			tokenResponse: &fosite.AccessResponse{
				AccessToken: "test-access-token",
				Extra: map[string]any{
					"refresh_token": "test-refresh-token",
				},
			},
			expectedCookies: []string{"access_token=test-access-token", "refresh_token=test-refresh-token"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			c, _ := gin.CreateTestContext(w)
			c.Request, _ = http.NewRequest("POST", "/", nil)

			tokenResponse := fosite.NewAccessResponse()
			tokenResponse.AccessToken = tt.tokenResponse.AccessToken
			for k, v := range tt.tokenResponse.Extra {
				tokenResponse.SetExtra(k, v)
			}

			setTokenCookie(c, tt.config, tokenResponse)

			cookies := w.Header().Values("Set-Cookie")
			for _, expected := range tt.expectedCookies {
				found := false
				for _, actual := range cookies {
					if strings.Contains(actual, expected) {
						found = true
						break
					}
				}
				assert.True(t, found, "Expected cookie %s not found in %v", expected, cookies)
			}
		})
	}
}
