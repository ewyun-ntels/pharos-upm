package user

import (
	"time"
)

type Store interface {
	ListUsers() ([]User, error)
	FindByUsername(username string) (User, error)
	CreateUser(username string, password string, attributes map[string]any, prepare *[]string) error
	DeleteUser(username string) error
	ChangePassword(username, password string) error

	SetAttributes(username string, attributes map[string]any) error
	DecodeExtra(attrs map[string]any) (map[string]any, error)

	SetBlock(username string, block bool) error
	Authenticate(username, password string) error
	InsertPasswordHistory(username, password string) error
	DeletePasswordHistory(username string, count int32) error
	CheckPasswordHistory(username string, password string) error
	UpdatePasswordExpiration(username string, expiredAt *time.Time) error
	UpdatePrepare(username string, prepare []string) error
	Close() error
	ValidatePasswordComplexity(password string) error

	GetLoginRetryLimit() int
	GetLoginBlockDuration() time.Duration
	GetPrepare(u User) ([]string, error)
}

type User interface {
	GetID() string
	GetBlocked() bool
	GetExtra() map[string]any
	GetCreatedAt() time.Time
	GetPasswordExpiresAt() *time.Time
	GetPrepare() []string
	GetTemporaryBlockedExpiresAt(limit int, duration time.Duration) *time.Time
	IsRetryAtTimedOut(timeout time.Duration) bool
	IsTemporarilyBlocked(limit int, duration time.Duration) bool
	HasRetry() bool
}

func NewStore(config StoreConfig) (Store, error) {
	return NewRepoStore(config) // Use the database config with SQLX config for settings
}
