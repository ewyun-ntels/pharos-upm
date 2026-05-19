package user

import (
	"time"

	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
)

type StoreConfig struct {
	Database orm.DatabaseConfig

	Prepare          []string
	UserEmailAllowed bool
	UserMinLength    int
	UserMaxLength    int

	LoginRetryLimit        int
	LoginBlockDuration     time.Duration
	LoginRetryResetTimeout time.Duration
	PasswordMinLength      int
	PasswordMinUppercase   int
	PasswordMinLowercase   int
	PasswordMinDigits      int
	PasswordMinSpecial     int
	PasswordTTL            string
	PasswordExpiredRule    []string
}

func GetUserStoreConfig(config common.Config) StoreConfig {
	return StoreConfig{
		Database: config.Database,

		Prepare:          config.User.Prepare,
		UserEmailAllowed: config.User.UserRule.EmailAllowed,
		UserMinLength:    config.User.UserRule.MinLength,
		UserMaxLength:    config.User.UserRule.MaxLength,

		LoginRetryLimit:        config.User.LoginRetryLimit,
		LoginBlockDuration:     time.Duration(config.User.LoginBlockDuration) * time.Second,
		LoginRetryResetTimeout: time.Duration(config.User.LoginRetryResetTimeout) * time.Second,
		PasswordMinLength:      config.User.PasswordRule.MinLength,
		PasswordMinUppercase:   config.User.PasswordRule.MinUppercase,
		PasswordMinLowercase:   config.User.PasswordRule.MinLowercase,
		PasswordMinDigits:      config.User.PasswordRule.MinDigits,
		PasswordMinSpecial:     config.User.PasswordRule.MinSpecial,
		PasswordTTL:            config.User.PasswordRule.PasswordTTL,
		PasswordExpiredRule:    config.User.PasswordRule.PasswordExpiredRule,
	}
}
