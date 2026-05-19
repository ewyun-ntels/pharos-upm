package adapters

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"ntels.com/pharos/core/pkg/common"
)

func TestAuthAdapter_ToAPIAuthConfig(t *testing.T) {
	adapter := &AuthAdapter{}

	config := common.Config{
		User: common.UsersConfig{
			LoginBlockDuration:     3600,
			LoginRetryResetTimeout: 600,
			PasswordRule: common.PasswordRuleConfig{
				MinDigits:    1,
				MinLength:    8,
				MinLowercase: 1,
				MinSpecial:   1,
				MinUppercase: 1,
				PasswordTTL:  "90d",
			},
			UserRule: common.UserRuleConfig{
				EmailAllowed: true,
				MaxLength:    20,
				MinLength:    4,
			},
		},
	}

	result := adapter.ToAPIAuthConfig(config)

	assert.Equal(t, int64(3600), result.Common.LoginBlockDuration)
	assert.Equal(t, int64(600), result.Common.LoginRetryResetTimeout)

	assert.Equal(t, int64(1), result.Password.MinDigits)
	assert.Equal(t, int64(8), result.Password.MinLength)
	assert.Equal(t, int64(1), result.Password.MinLowercase)
	assert.Equal(t, int64(1), result.Password.MinSpecial)
	assert.Equal(t, int64(1), result.Password.MinUppercase)
	assert.NotNil(t, result.Password.PasswordTTL)
	// 90 days = 90 * 24 * 3600 = 7776000 seconds
	assert.Equal(t, int64(7776000), *result.Password.PasswordTTL)

	assert.Equal(t, true, result.User.EmailAllowed)
	assert.Equal(t, int64(20), result.User.MaxLength)
	assert.Equal(t, int64(4), result.User.MinLength)
}

func TestAuthAdapter_ToAPIAuthConfig_InvalidTTL(t *testing.T) {
	adapter := &AuthAdapter{}

	config := common.Config{
		User: common.UsersConfig{
			PasswordRule: common.PasswordRuleConfig{
				PasswordTTL: "invalid",
			},
		},
	}

	result := adapter.ToAPIAuthConfig(config)

	assert.NotNil(t, result.Password.PasswordTTL)
	assert.Equal(t, int64(0), *result.Password.PasswordTTL)
}
