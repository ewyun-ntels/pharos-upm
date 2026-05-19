package sqlite

import (
	"context"
	"database/sql"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jmoiron/sqlx"
	"golang.org/x/crypto/bcrypt"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/internal/repositories"
)

var _ repositories.UserRepository = (*userRepository)(nil)

// userRepository implements the repositories.UserRepository interface
// using SQLite as the backend.
type userRepository struct {
	repositories.BaseRepository
}

func newUserRepository(config orm.DatabaseConfig) *userRepository {
	return &userRepository{
		BaseRepository: repositories.NewBaseRepository(config),
	}
}

type sqliteUser struct {
	Username          string     `db:"username"`
	PasswordHash      []byte     `db:"password_hash"`
	Extra             []byte     `db:"extra"`
	CreatedAt         time.Time  `db:"created_at"`
	Retry             int        `db:"retry"`
	RetryAt           *time.Time `db:"retry_at"`
	Blocked           bool       `db:"blocked"`
	Prepare           []byte     `db:"prepare"`
	PasswordExpiredAt *time.Time `db:"password_expired_at"`
	PasswordUpdatedAt *time.Time `db:"password_updated_at"`
}

func (repo *userRepository) toUser(dbUser sqliteUser) (*repositories.UserEntity, error) {
	attributes, err := repo.UnmarshalAttributes(dbUser.Extra)
	if err != nil {
		return nil, err
	}

	var prepare []string
	prepare, err = repo.UnmarshalPrepare(dbUser.Prepare)
	if err != nil {
		return nil, err
	}

	u := &repositories.UserEntity{
		Username:   dbUser.Username,
		Attributes: attributes,
		CreatedAt:  dbUser.CreatedAt,
		Retry:      dbUser.Retry,
		Blocked:    dbUser.Blocked,
		Prepare:    prepare,
	}

	if dbUser.RetryAt != nil {
		u.RetryAt = *dbUser.RetryAt
	}

	if dbUser.PasswordExpiredAt != nil {
		u.PasswordExpiredAt = *dbUser.PasswordExpiredAt
	}

	if dbUser.PasswordUpdatedAt != nil {
		u.PasswordUpdatedAt = *dbUser.PasswordUpdatedAt
	}

	return u, nil
}

func (repo *userRepository) generateHash(password []byte) ([]byte, error) {
	if password == nil {
		return nil, nil
	}
	return bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
}

func (repo *userRepository) compareHash(hash, password []byte) error {
	return bcrypt.CompareHashAndPassword(hash, password)
}

func (repo *userRepository) execUpdate(ctx context.Context, query string, args ...any) error {
	return repo.Handler(func(db *sqlx.DB) error {
		res, err := db.ExecContext(ctx, query, args...)
		if err != nil {
			return err
		}

		n, _ := res.RowsAffected()
		if n == 0 {
			return sql.ErrNoRows
		}
		return nil
	})
}

// Create inserts a new user into the database.
func (repo *userRepository) Create(ctx context.Context, u *repositories.UserEntity, password []byte) error {
	query := `
	INSERT INTO users
	   (username, password_hash,
	    extra, created_at,
	    retry, retry_at,
	    blocked, password_expired_at, prepare)
	VALUES ($1, $2,
	       $3, strftime('%Y-%m-%d %H:%M:%f', 'now'),
	       0, null,
	       false, $4, $5)`

	passwordHash, err := repo.generateHash(password)
	if err != nil {
		return err
	}

	extra, err := repo.MarshalAttributes(u.Attributes)
	if err != nil {
		return err
	}

	prepare, err := repo.MarshalPrepare(u.Prepare)
	if err != nil {
		return err
	}

	passwordExpiredAt := toNullableTime(u.PasswordExpiredAt)

	return repo.Handler(func(db *sqlx.DB) error {
		_, err = db.ExecContext(ctx, query,
			u.Username, passwordHash,
			extra, passwordExpiredAt, prepare)
		if err != nil {
			return err
		}
		return nil
	})
}

// GetByName retrieves a user by username.
func (repo *userRepository) GetByName(ctx context.Context, username string) (*repositories.UserEntity, error) {
	var dbUser sqliteUser
	query := `
SELECT
    username,
    password_hash,
    created_at,
    retry,
    retry_at,
    extra,
    blocked,
    prepare,
    password_expired_at
FROM users
WHERE username = $1`

	err := repo.Handler(func(db *sqlx.DB) error {
		return db.GetContext(ctx, &dbUser, query, username)
	})
	if err != nil {
		return nil, err
	}

	return repo.toUser(dbUser)
}

// ListAll retrieves all users from the database.
func (repo *userRepository) ListAll(ctx context.Context) ([]*repositories.UserEntity, error) {
	query := `
SELECT
    users.username,
    users.extra,
    users.created_at,
    users.blocked,
    users.prepare,
    users.retry,
    users.retry_at,
    users.password_expired_at
FROM users
`
	var dbUsers []sqliteUser
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.SelectContext(ctx, &dbUsers, query)
	})
	if err != nil {
		return nil, err
	}

	var users []*repositories.UserEntity
	for _, dbUser := range dbUsers {
		var u *repositories.UserEntity
		u, err = repo.toUser(dbUser)
		if err != nil {
			return nil, err
		}
		users = append(users, u)
	}
	return users, nil
}

// ComparePassword checks if the provided password matches the stored password hash.
func (repo *userRepository) ComparePassword(ctx context.Context, username string, password []byte) error {
	var storedHash []byte
	query := "SELECT password_hash FROM users WHERE username = $1"

	err := repo.Handler(func(db *sqlx.DB) error {
		return db.GetContext(ctx, &storedHash, query, username)
	})
	if err != nil {
		return err
	}

	return repo.compareHash(storedHash, password)
}

// UpdatePassword updates a user's password.
func (repo *userRepository) UpdatePassword(ctx context.Context, username string, password []byte, passwordExpiredAt time.Time) error {
	query := `
UPDATE users SET
    password_hash = $1,
    password_updated_at = strftime('%Y-%m-%d %H:%M:%f', 'now'),
    password_expired_at = $2
 WHERE username = $3`

	passwordHash, err := repo.generateHash(password)
	if err != nil {
		return err
	}

	expiredAt := toNullableTime(passwordExpiredAt)

	return repo.execUpdate(ctx, query, passwordHash, expiredAt, username)
}

// UpdatePasswordExpiredAt updates a user's password expiration time.
func (repo *userRepository) UpdatePasswordExpiredAt(ctx context.Context, username string, expiredAt time.Time) error {
	query := `
UPDATE users SET
	password_expired_at = $1
WHERE username = $2`

	t := toNullableTime(expiredAt)

	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.ExecContext(ctx, query, t, username)
		return err
	})
}

// UpdateBlocked updates a user's blocked status.
func (repo *userRepository) UpdateBlocked(ctx context.Context, username string, blocked bool) error {
	query := `
UPDATE users SET
	blocked = $1
WHERE username = $2`

	return repo.execUpdate(ctx, query, blocked, username)
}

// UpdateAttributes updates a user's attributes.
func (repo *userRepository) UpdateAttributes(ctx context.Context, username string, attributes map[string]any) error {
	query := `
UPDATE users SET
	extra = $1
WHERE username = $2`

	extra, err := repo.MarshalAttributes(attributes)
	if err != nil {
		return err
	}

	return repo.execUpdate(ctx, query, extra, username)
}

// IncreaseRetryCount increments a user's retry count and sets the retry timestamp.
func (repo *userRepository) IncreaseRetryCount(ctx context.Context, username string) error {
	query := `
UPDATE users SET
	retry = coalesce(retry, 0) + 1,
	retry_at = strftime('%Y-%m-%d %H:%M:%f', 'now')
WHERE username = $1`

	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.ExecContext(ctx, query, username)
		return err
	})
}

// ResetRetryCount resets a user's retry count and clears the retry timestamp.
func (repo *userRepository) ResetRetryCount(ctx context.Context, username string) error {
	query := `
UPDATE users SET
	retry = 0,
	retry_at = null
WHERE username = $1`

	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.ExecContext(ctx, query, username)
		return err
	})
}

// UpdatePrepare updates a user's prepare commands.
func (repo *userRepository) UpdatePrepare(ctx context.Context, username string, prepare []string) error {
	query := `
UPDATE users SET
	prepare = $1
WHERE username = $2`

	prepareData, err := repo.MarshalPrepare(prepare)
	if err != nil {
		return err
	}

	return repo.Handler(func(db *sqlx.DB) error {
		_, err = db.ExecContext(ctx, query, prepareData, username)
		return err
	})
}

// DeleteByName deletes a user by username.
func (repo *userRepository) DeleteByName(ctx context.Context, username string) error {
	query := "DELETE FROM users WHERE username = $1"
	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.ExecContext(ctx, query, username)
		return err
	})
}

// InsertPasswordHistory adds a new password history record for the specified username.
func (repo *userRepository) InsertPasswordHistory(ctx context.Context, username string, password []byte) error {
	query := `
INSERT INTO user_password_history
    (id, username, password_hash, created_at)
VALUES ($1, $2, $3, strftime('%Y-%m-%d %H:%M:%f', 'now'))`

	id := uuid.NewString()

	hash, err := repo.generateHash(password)
	if err != nil {
		return err
	}

	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.ExecContext(ctx, query, id, username, hash)
		return err
	})
}

// DeletePasswordHistory ensures that only the most recent 'n' password history records
// for the specified username are retained, deleting older records.
func (repo *userRepository) DeletePasswordHistory(ctx context.Context, username string, n int32) error {
	query := `
DELETE FROM user_password_history
WHERE username = $1
AND id NOT IN (
    SELECT id
    FROM user_password_history
    WHERE username = $1
    ORDER BY created_at DESC
    LIMIT $2
)`

	return repo.Handler(func(db *sqlx.DB) error {
		_, err := db.ExecContext(ctx, query, username, n)
		return err
	})
}

// ListPasswordHistory retrieves all password hashes associated with the specified username.
func (repo *userRepository) ListPasswordHistory(ctx context.Context, username string) ([]string, error) {
	var histories []string
	// 명시적으로 ORDER BY를 지정하여 일관된 결과를 반환
	query := `
SELECT password_hash
FROM user_password_history
WHERE username = $1
ORDER BY created_at
`
	err := repo.Handler(func(db *sqlx.DB) error {
		return db.SelectContext(ctx, &histories, query, username)
	})
	if err != nil {
		return nil, err
	}
	return histories, nil
}

// IsPasswordReuse checks if the provided password has been used before by the specified username.
func (repo *userRepository) IsPasswordReuse(ctx context.Context, username string, password string) (bool, error) {
	histories, err := repo.ListPasswordHistory(ctx, username)
	if err != nil {
		return false, err
	}

	for _, hash := range histories {
		err = repo.compareHash([]byte(hash), []byte(password))
		if err == nil {
			return true, nil
		} else if !errors.Is(err, bcrypt.ErrMismatchedHashAndPassword) {
			return false, err
		}
	}

	return false, nil
}
