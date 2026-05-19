package repositories

import (
	"context"
	"time"
)

// UserEntity represents a user entity
type UserEntity struct {
	Username          string
	Attributes        map[string]any
	CreatedAt         time.Time
	Retry             int
	RetryAt           time.Time
	Blocked           bool
	Prepare           []string
	PasswordExpiredAt time.Time
	PasswordUpdatedAt time.Time
}

func (u *UserEntity) GetID() string {
	return u.Username
}
func (u *UserEntity) GetBlocked() bool {
	return u.Blocked
}
func (u *UserEntity) GetExtra() map[string]any {
	return u.Attributes
}
func (u *UserEntity) GetCreatedAt() time.Time {
	return u.CreatedAt
}
func (u *UserEntity) GetPasswordExpiresAt() *time.Time {
	if u.PasswordExpiredAt.IsZero() {
		return nil
	}
	return &u.PasswordExpiredAt
}
func (u *UserEntity) GetPrepare() []string {
	return u.Prepare
}
func (u *UserEntity) GetTemporaryBlockedExpiresAt(limit int, duration time.Duration) *time.Time {
	if u.IsTemporarilyBlocked(limit, duration) {
		expiredAt := u.RetryAt.Add(duration)
		return &expiredAt
	}
	return nil
}
func (u *UserEntity) IsRetryAtTimedOut(timeout time.Duration) bool {
	return !u.RetryAt.IsZero() && time.Now().After(u.RetryAt.Add(timeout))
}
func (u *UserEntity) IsTemporarilyBlocked(limit int, duration time.Duration) bool {
	return u.Retry >= limit && !u.RetryAt.IsZero() && time.Now().Before(u.RetryAt.Add(duration))
}
func (u *UserEntity) HasRetry() bool {
	return u.Retry != 0 && !u.RetryAt.IsZero()
}

// UserRepository defines the interface for user data access
type UserRepository interface {
	Create(ctx context.Context, user *UserEntity, password []byte) error
	GetByName(ctx context.Context, username string) (*UserEntity, error)
	ListAll(ctx context.Context) ([]*UserEntity, error)
	ComparePassword(ctx context.Context, username string, password []byte) error

	UpdatePassword(ctx context.Context, username string, password []byte, passwordExpiredAt time.Time) error
	UpdatePasswordExpiredAt(ctx context.Context, username string, expired time.Time) error
	UpdateBlocked(ctx context.Context, username string, blocked bool) error
	UpdateAttributes(ctx context.Context, username string, attributes map[string]any) error
	IncreaseRetryCount(ctx context.Context, username string) error
	ResetRetryCount(ctx context.Context, username string) error
	UpdatePrepare(ctx context.Context, username string, prepare []string) error

	DeleteByName(ctx context.Context, username string) error

	InsertPasswordHistory(ctx context.Context, username string, password []byte) error
	DeletePasswordHistory(ctx context.Context, username string, n int32) error
	ListPasswordHistory(ctx context.Context, username string) ([]string, error)
	IsPasswordReuse(ctx context.Context, username string, password string) (bool, error)
}
