package adapters

import (
	pd "github.com/SiverPineValley/parseduration"
	"ntels.com/pharos/core/pkg/common"
	auth_types "ntels.com/pharos/shared/types/auth"
)

type AuthAdapter struct{}

func (a *AuthAdapter) ToAPIAuthConfig(config common.Config) auth_types.AuthConfig {
	duration, _ := pd.ParseDuration(config.User.PasswordRule.PasswordTTL)
	passwordTTL := int64(duration.Seconds())
	collectClientInfo := config.Auth.CollectClientInfo

	return auth_types.AuthConfig{
		Common: auth_types.Common{
			LoginBlockDuration:     int64(config.User.LoginBlockDuration),
			LoginRetryResetTimeout: int64(config.User.LoginRetryResetTimeout),
			CollectClientInfo:      &collectClientInfo,
		},
		Password: auth_types.PasswordPolicy{
			MinDigits:    int64(config.User.PasswordRule.MinDigits),
			MinLength:    int64(config.User.PasswordRule.MinLength),
			MinLowercase: int64(config.User.PasswordRule.MinLowercase),
			MinSpecial:   int64(config.User.PasswordRule.MinSpecial),
			MinUppercase: int64(config.User.PasswordRule.MinUppercase),
			PasswordTTL:  &passwordTTL,
		},
		User: auth_types.UserPolicy{
			EmailAllowed: config.User.UserRule.EmailAllowed,
			MaxLength:    int64(config.User.UserRule.MaxLength),
			MinLength:    int64(config.User.UserRule.MinLength),
		},
	}
}
