package user

import (
	"context"
	"database/sql"
	"errors"
	"log/slog"
	"time"

	pd "github.com/SiverPineValley/parseduration"
	"github.com/ory/fosite"
	"ntels.com/pharos/core/internal/repositories"
	"ntels.com/pharos/core/internal/repositories/factory"
	roleTypes "ntels.com/pharos/shared/types/role"
)

var _ Store = (*RepoStore)(nil)

type RepoStore struct {
	repo repositories.UserRepository

	Prepare []string

	LoginRetryLimit        int
	LoginBlockDuration     time.Duration
	LoginRetryResetTimeout time.Duration
	PasswordTTL            time.Duration
	PasswordExpiredRule    []string

	// UsernameValidator enforces username complexity; defaults to usernameValidator
	UsernameValidator UsernameValidator

	// PasswordValidator enforces password complexity; defaults to RuleBasedPasswordValidator
	PasswordValidator PasswordValidator
}

func NewRepoStore(cfg StoreConfig) (*RepoStore, error) {
	uv := newUsernameValidator(cfg)
	pv := newPasswordValidator(cfg)

	store := &RepoStore{
		UsernameValidator: uv,
		PasswordValidator: pv,

		Prepare: cfg.Prepare,

		LoginRetryLimit:        cfg.LoginRetryLimit,
		LoginBlockDuration:     cfg.LoginBlockDuration,
		LoginRetryResetTimeout: cfg.LoginRetryResetTimeout,

		PasswordExpiredRule: cfg.PasswordExpiredRule,
	}

	store.Prepare = cfg.Prepare
	for _, prepare := range store.Prepare {
		switch prepare {
		case PreparePasswordChange:
			continue
		default:
			panic("unknown user prepare(" + prepare + ")")
		}
	}

	if store.LoginRetryLimit == 0 {
		store.LoginRetryLimit = 5
	}
	if store.LoginBlockDuration == 0 {
		store.LoginBlockDuration = 15 * time.Minute
	}
	if store.LoginRetryResetTimeout == 0 {
		store.LoginRetryResetTimeout = 15 * time.Minute
	}

	if cfg.PasswordTTL != "" {
		duration, err := pd.ParseDuration(cfg.PasswordTTL)
		if err != nil {
			return nil, err
		}
		store.PasswordTTL = duration
	}

	for _, prepare := range store.PasswordExpiredRule {
		switch prepare {
		case PreparePasswordChange:
			continue
		default:
			panic("unknown user password_expired_rule(" + prepare + ")")
		}
	}

	var err error
	store.repo, err = factory.NewUserRepository(factory.RepositoryOptions{
		DatabaseConfig: cfg.Database,
	})
	if err != nil {
		return nil, err
	}

	return store, nil
}

func (store *RepoStore) newPasswordExpiredAt() time.Time {
	if store.PasswordTTL != 0 {
		return time.Now().UTC().Add(store.PasswordTTL)
	}
	return time.Time{}
}

func (store *RepoStore) ListUsers() ([]User, error) {
	ctx := context.Background()
	users, err := store.repo.ListAll(ctx)
	if err != nil {
		slog.Error("db.Select failed", "error", err)
		return nil, err
	}

	// to convert [](*repositories.UserEntity) to []User
	var iusers []User
	for _, u := range users {
		iusers = append(iusers, u)
	}

	return iusers, nil
}
func (store *RepoStore) FindByUsername(username string) (User, error) {
	ctx := context.Background()
	u, err := store.repo.GetByName(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, fosite.ErrNotFound
		}
		return nil, err
	}

	return u, nil
}
func (store *RepoStore) CreateUser(username string, password string, attributes map[string]any, prepare *[]string) error {
	ctx := context.Background()

	err := store.UsernameValidator.Validate(username)
	if err != nil {
		slog.Error("invalid username", "error", err)
		return err
	}

	_, err = store.FindByUsername(username)
	if err == nil {
		return ErrUserAlreadyExists
	} else if !errors.Is(err, fosite.ErrNotFound) {
		return err
	}

	targetPrepare := store.Prepare
	if prepare != nil {
		targetPrepare = *prepare
	}

	newUser := &repositories.UserEntity{
		Username:   username,
		Attributes: attributes,
		Prepare:    targetPrepare,

		PasswordExpiredAt: store.newPasswordExpiredAt(),
	}

	err = store.repo.Create(ctx, newUser, []byte(password))
	if err != nil {
		slog.Error("create user failed", "error", err)
		return err
	}
	return nil
}
func (store *RepoStore) DeleteUser(username string) error {
	ctx := context.Background()
	_, err := store.FindByUsername(username)
	if err != nil {
		return err
	}

	return store.repo.DeleteByName(ctx, username)
}
func (store *RepoStore) ChangePassword(username, password string) error {
	ctx := context.Background()
	err := store.repo.UpdatePassword(ctx, username, []byte(password), store.newPasswordExpiredAt())
	if err != nil {
		slog.Error("update user password failed", "error", err)
		return err
	}

	return nil
}
func (store *RepoStore) SetAttributes(username string, attributes map[string]any) error {
	ctx := context.Background()
	err := store.repo.UpdateAttributes(ctx, username, attributes)
	if err != nil {
		slog.Error("update user extra failed", "error", err)
		return err
	}
	return nil
}
func (store *RepoStore) DecodeExtra(attrs map[string]any) (map[string]any, error) {
	// DB stores attributes in nested structure: {"roles": {...}, "info": {...}}
	// Extract roles, apply CompositeRole expansion, then reconstruct

	rolesMap, ok := attrs["roles"].(map[string]any)
	if !ok || rolesMap == nil {
		// No roles to expand
		return attrs, nil
	}

	// Reconstruct attrs with expanded roles
	result := make(map[string]any)
	for k, v := range attrs {
		if k == "roles" {
			result["roles"] = rolesMap
		} else {
			result[k] = v
		}
	}

	return result, nil
}

func (store *RepoStore) SetBlock(username string, block bool) error {
	ctx := context.Background()

	err := store.repo.UpdateBlocked(ctx, username, block)
	if err != nil {
		slog.Error("update user blocked failed", "error", err)
		return err
	}
	return nil
}
func (store *RepoStore) Authenticate(username, password string) error {
	ctx := context.Background()
	intErrInvalidReqOrigin := *fosite.ErrInvalidRequest
	intErrInvalidReqOrigin.HintField = ""
	intErrNotFoundOrigin := *fosite.ErrNotFound
	intInvalidReqErr := &intErrInvalidReqOrigin
	intNotFoundErr := &intErrNotFoundOrigin
	u, err := store.FindByUsername(username)
	if err != nil {
		intNotFoundErr.DescriptionField = "user not found"
		return intNotFoundErr
	}

	if u.GetBlocked() {
		slog.Error("user is blocked", "username", username)
		intInvalidReqErr.DescriptionField = "user is blocked"
		return intInvalidReqErr
	}

	attrs := u.GetExtra()
	attrs, err = store.DecodeExtra(attrs)
	if err != nil {
		slog.Error("DecodeExtra failed", "error", err)
		intInvalidReqErr.DescriptionField = "internal server error"
		return intInvalidReqErr
	}

	skipTemporarilyBlock := false
	if val, exists := attrs[string(roleTypes.AttrSkipTemporarilyBlock)]; exists {
		if boolVal, ok := val.(bool); ok {
			skipTemporarilyBlock = boolVal
		}
	}

	passwordRetryUnlimited := false
	if val, exists := attrs[string(roleTypes.AttrPasswordRetryUnlimited)]; exists {
		if boolVal, ok := val.(bool); ok {
			passwordRetryUnlimited = boolVal
		}
	}

	if passwordRetryUnlimited || u.IsRetryAtTimedOut(store.LoginRetryResetTimeout) {
		err = store.repo.ResetRetryCount(ctx, username)
		if err != nil {
			slog.Error("reset retry count failed", "error", err)
		}
	} else if !skipTemporarilyBlock && u.IsTemporarilyBlocked(store.LoginRetryLimit, store.LoginBlockDuration) {
		slog.Error("user is temporarily blocked", "username", username)
		err = store.repo.IncreaseRetryCount(ctx, username)
		if err != nil {
			slog.Error("IncreaseRetryCount failed", "error", err)
		}
		intInvalidReqErr.DescriptionField = "user is temporarily blocked"
		return intInvalidReqErr
	}

	err = store.repo.ComparePassword(ctx, username, []byte(password))
	if err != nil {
		slog.Error("ComparePassword failed", "error", err)
		err = store.repo.IncreaseRetryCount(ctx, username)
		if err != nil {
			slog.Error("IncreaseRetryCount failed", "error", err)
			return fosite.ErrNotFound
		}
		intInvalidReqErr.DescriptionField = "password mismatch"
		return intInvalidReqErr
	}

	err = store.repo.ResetRetryCount(ctx, username)
	if err != nil {
		slog.Error("ResetRetryCount failed", "error", err)
	}

	return nil
}
func (store *RepoStore) InsertPasswordHistory(username, password string) error {
	ctx := context.Background()
	err := store.repo.InsertPasswordHistory(ctx, username, []byte(password))
	if err != nil {
		slog.Error("insert password history failed", "error", err)
		return err
	}

	return nil
}
func (store *RepoStore) DeletePasswordHistory(username string, count int32) error {
	ctx := context.Background()
	err := store.repo.DeletePasswordHistory(ctx, username, count)
	if err != nil {
		slog.Error("delete password history failed", "error", err)
		return err
	}
	return nil
}
func (store *RepoStore) CheckPasswordHistory(username string, password string) error {
	ctx := context.Background()
	used, err := store.repo.IsPasswordReuse(ctx, username, password)
	if err != nil {
		slog.Error("check password history failed", "error", err)
		return err
	}
	if used {
		return errors.New("password has been used before")
	}

	return nil
}
func (store *RepoStore) UpdatePasswordExpiration(username string, expiredAt *time.Time) error {
	ctx := context.Background()

	_, err := store.repo.GetByName(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fosite.ErrNotFound
		}
		return err
	}

	var t time.Time
	if expiredAt != nil {
		t = *expiredAt
	}
	err = store.repo.UpdatePasswordExpiredAt(ctx, username, t)
	if err != nil {
		slog.Error("update user password expired at failed", "error", err)
		return err
	}
	return nil
}
func (store *RepoStore) UpdatePrepare(username string, prepare []string) error {
	ctx := context.Background()
	_, err := store.repo.GetByName(ctx, username)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return fosite.ErrNotFound
		}
		return err
	}

	err = store.repo.UpdatePrepare(ctx, username, prepare)
	if err != nil {
		return err
	}
	return nil
}
func (store *RepoStore) Close() error { return nil }
func (store *RepoStore) ValidatePasswordComplexity(password string) error {
	return store.PasswordValidator.Validate(password)
}
func (store *RepoStore) GetLoginRetryLimit() int              { return store.LoginRetryLimit }
func (store *RepoStore) GetLoginBlockDuration() time.Duration { return store.LoginBlockDuration }
func (store *RepoStore) GetPrepare(u User) ([]string, error) {
	prepare := u.GetPrepare()
	if u.GetPasswordExpiresAt() != nil && u.GetPasswordExpiresAt().Before(time.Now().UTC()) {
		if len(store.PasswordExpiredRule) == 0 {
			passwordErr := *fosite.ErrInvalidRequest

			passwordErr.DescriptionField = "password expired"
			return nil, &passwordErr
		}
		prepare = store.PasswordExpiredRule
	}

	return prepare, nil
}
