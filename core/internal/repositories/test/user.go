package test

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"ntels.com/pharos/core/internal/repositories"
)

type UserTestable interface {
	repositories.UserRepository
	SetUpDb() error
	TearDownDb() error
	SetUpUser(username string) error

	ClearTable() error
	GetUser(name string) (*repositories.UserEntity, error)
	GetUserCount() (int, error)
	GetPasswordHistories(username string) ([]string, error)
	GetLatestPasswordHistory(username string) (string, time.Time, error)
	CompareHashedPassword(hashedPassword, password []byte) error
}

func UserSuite(t *testing.T, testable UserTestable) {
	t.Run("User Repository Suite", func(t *testing.T) {
		t.Run("Create", func(t *testing.T) { UserCreate(t, testable) })
		t.Run("GetByName", func(t *testing.T) { UserGetByName(t, testable) })
		t.Run("ListAll", func(t *testing.T) { UserListAll(t, testable) })
		t.Run("ComparePassword", func(t *testing.T) { UserComparePassword(t, testable) })
		t.Run("UpdatePassword", func(t *testing.T) { UserUpdatePassword(t, testable) })
		t.Run("UpdatePasswordExpiredAt", func(t *testing.T) { UserUpdatePasswordExpiredAt(t, testable) })
		t.Run("UpdateBlocked", func(t *testing.T) { UserUpdateBlocked(t, testable) })
		t.Run("UpdateAttributes", func(t *testing.T) { UserUpdateAttributes(t, testable) })
		t.Run("IncreaseRetryCount", func(t *testing.T) { UserIncreaseRetryCount(t, testable) })
		t.Run("ResetRetryCount", func(t *testing.T) { UserResetRetryCount(t, testable) })
		t.Run("UpdatePrepare", func(t *testing.T) { UserUpdatePrepare(t, testable) })
		t.Run("DeleteByName", func(t *testing.T) { UserDeleteByName(t, testable) })
		t.Run("InsertPasswordHistory", func(t *testing.T) { UserInsertPasswordHistory(t, testable) })
		t.Run("DeletePasswordHistory", func(t *testing.T) { UserDeletePasswordHistory(t, testable) })
		t.Run("ListPasswordHistory", func(t *testing.T) { UserListPasswordHistory(t, testable) })
		t.Run("IsPasswordReuse", func(t *testing.T) { UserIsPasswordReuse(t, testable) })
	})
}

func UserCreate(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	passwordExpiredAt := time.Now().Add(24 * time.Hour).In(testTimeZone)
	u := &repositories.UserEntity{
		Username: "testuser",
		Attributes: map[string]any{
			"role": "tester",
		},
		PasswordExpiredAt: passwordExpiredAt,
	}
	password := []byte("testpassword")

	var countBefore int
	countBefore, err = testable.GetUserCount()
	if err != nil {
		t.Fatalf("failed to get user count: %v", err)
	}

	now := time.Now()
	err = testable.Create(context.Background(), u, password)
	if err != nil {
		t.Fatalf("failed to create user: %v", err)
	}
	afterCreate := time.Now()

	var countAfter int
	countAfter, err = testable.GetUserCount()
	if err != nil {
		t.Fatalf("failed to get user count: %v", err)
	}

	if countAfter != countBefore+1 {
		t.Fatalf("user count did not increase after creation: before=%d, after=%d", countBefore, countAfter)
	}

	createdUser, err := testable.GetUser("testuser")
	if err != nil {
		t.Fatalf("failed to get created user: %v", err)
	}
	if createdUser == nil {
		t.Fatalf("created user is nil")
	}
	if createdUser.Username != "testuser" {
		t.Fatalf("username mismatch: expected 'testuser', got '%s'", createdUser.Username)
	}
	if role, ok := createdUser.Attributes["role"]; !ok || role != "tester" {
		t.Fatalf("attributes mismatch: expected role 'tester', got '%v'", createdUser.Attributes["role"])
	}

	if createdUser.CreatedAt.Sub(now).Abs() > maximumAllowedTimeDeltaForNow {
		t.Fatalf("CreatedAt timestamp is not recent: %v, expected between %v, %v", createdUser.CreatedAt, now, afterCreate)
	}

	if createdUser.PasswordExpiredAt.IsZero() || createdUser.PasswordExpiredAt.Sub(passwordExpiredAt).Abs() >= time.Second {
		t.Fatalf("PasswordExpiredAt mismatch: expected ~%v, got %v", passwordExpiredAt, createdUser.PasswordExpiredAt)
	}

	err = testable.ComparePassword(context.Background(), u.Username, password)
	if err != nil {
		t.Fatalf("password check failed: %v", err)
	}
}

func UserGetByName(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	username := "testuser"
	err = testable.SetUpUser(username)
	if err != nil {
		t.Fatalf("failed to set up test user: %v", err)
	}

	u, err := testable.GetByName(context.Background(), username)
	if err != nil {
		t.Fatalf("failed to get user by name: %v", err)
	}
	if u == nil {
		t.Fatalf("user is nil")
	}
	if u.Username != username {
		t.Fatalf("사용자 이름 불일치: got %s, want %s", u.Username, "testuser")
	}

	_, err = testable.GetByName(context.Background(), "nonexistent")
	if err == nil {
		t.Fatalf("존재하지 않는 사용자에 대해 오류가 발생해야 합니다")
	} else if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("예상치 못한 오류: %v", err)
	}
}

func UserListAll(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	err = testable.ClearTable()
	if err != nil {
		t.Fatalf("failed to clear users table: %v", err)
	}

	users := []string{"user1", "user2", "user3"}
	for _, username := range users {
		err = testable.SetUpUser(username)
		if err != nil {
			t.Fatalf("failed to set up test user %s: %v", username, err)
		}
	}

	allUsers, err := testable.ListAll(context.Background())
	if err != nil {
		t.Fatalf("ListAll 실패: %v", err)
	} else if len(allUsers) != len(users) {
		t.Fatalf("사용자 수 불일치: got %d, want %d", len(allUsers), len(users))
	}

	userMap := make(map[string]bool)
	for _, u := range allUsers {
		userMap[u.Username] = true
	}
	for _, username := range users {
		if !userMap[username] {
			t.Fatalf("누락된 사용자: %s", username)
		}
	}
}

func UserComparePassword(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	username := "passuser"
	password := []byte("correctpassword")
	err = testable.SetUpUser(username)
	if err != nil {
		t.Fatalf("failed to set up test user: %v", err)
	}

	err = testable.UpdatePassword(context.Background(), username, password, time.Time{})
	if err != nil {
		t.Fatalf("failed to set password: %v", err)
	}

	err = testable.ComparePassword(context.Background(), username, password)
	if err != nil {
		t.Fatalf("password comparison failed for correct password: %v", err)
	}

	err = testable.ComparePassword(context.Background(), username, []byte("wrongpassword"))
	if err == nil {
		t.Fatalf("password comparison should fail for incorrect password")
	}

	err = testable.ComparePassword(context.Background(), "nonexistent", password)
	if err == nil {
		t.Fatalf("존재하지 않는 사용자에 대해 오류가 발생해야 합니다")
	}
}

func UserUpdatePassword(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	userName := "passworduser"
	err = testable.SetUpUser(userName)
	if err != nil {
		t.Fatalf("failed to set up test user: %v", err)
	}

	newPassword := []byte("newpassword")
	passwordExpiredAt := time.Now().Add(24 * time.Hour).In(testTimeZone)
	err = testable.UpdatePassword(context.Background(), userName, newPassword, passwordExpiredAt)
	if err != nil {
		t.Fatalf("failed to update password: %v", err)
	}

	err = testable.ComparePassword(context.Background(), userName, newPassword)
	if err != nil {
		t.Fatalf("password check failed after update: %v", err)
	}

	u, err := testable.GetUser(userName)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	} else if u == nil {
		t.Fatalf("user is nil")
	} else if u.PasswordExpiredAt.IsZero() || u.PasswordExpiredAt.Sub(passwordExpiredAt).Abs() >= time.Second {
		t.Fatalf("PasswordExpiredAt mismatch after update: expected ~%v, got %v", passwordExpiredAt, u.PasswordExpiredAt)
	}
}

func UserUpdatePasswordExpiredAt(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	username := "expireuser"
	err = testable.SetUpUser(username)
	if err != nil {
		t.Fatalf("failed to set up test user: %v", err)
	}

	expiredAt := time.Now().Add(24 * time.Hour).In(testTimeZone)
	err = testable.UpdatePasswordExpiredAt(context.Background(), username, expiredAt)
	if err != nil {
		t.Fatalf("failed to update password expiration: %v", err)
	}

	u, err := testable.GetUser(username)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	} else if u == nil {
		t.Fatalf("user is nil")
	} else if u.PasswordExpiredAt.IsZero() || u.PasswordExpiredAt.Sub(expiredAt).Abs() >= time.Second {
		t.Fatalf("password expiration time mismatch: got %v, want %v", u.PasswordExpiredAt, expiredAt)
	}

	err = testable.UpdatePasswordExpiredAt(context.Background(), username, time.Time{})
	if err != nil {
		t.Fatalf("failed to clear password expiration: %v", err)
	}

	u, err = testable.GetUser(username)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	} else if u == nil {
		t.Fatalf("user is nil")
	} else if !u.PasswordExpiredAt.IsZero() {
		t.Fatalf("password expiration time should be zero, got %v", u.PasswordExpiredAt)
	}

	err = testable.UpdatePasswordExpiredAt(context.Background(), "nonexistent", expiredAt)
	if err != nil {
		t.Fatalf("존재하지 않는 사용자에 대해 오류가 발생하지 않아야 합니다: %v", err)
	}
}

func UserUpdateBlocked(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	username := "blockuser"
	err = testable.SetUpUser(username)
	if err != nil {
		t.Fatalf("failed to set up test user: %v", err)
	}

	err = testable.UpdateBlocked(context.Background(), username, true)
	if err != nil {
		t.Fatalf("failed to block user: %v", err)
	}

	u, err := testable.GetUser(username)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	} else if u == nil {
		t.Fatalf("user is nil")
	} else if !u.Blocked {
		t.Fatalf("사용자가 차단되지 않음")
	}

	err = testable.UpdateBlocked(context.Background(), username, false)
	if err != nil {
		t.Fatalf("failed to unblock user: %v", err)
	}

	u, err = testable.GetUser(username)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	} else if u == nil {
		t.Fatalf("user is nil")
	} else if u.Blocked {
		t.Fatalf("사용자가 차단 해제되지 않음")
	}

	err = testable.UpdateBlocked(context.Background(), "nonexistent", true)
	if err == nil {
		t.Fatalf("존재하지 않는 사용자에 대해 오류가 발생해야 합니다")
	} else if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("예상치 못한 오류: %v", err)
	}
}

func UserUpdateAttributes(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	username := "attruser"
	err = testable.SetUpUser(username)
	if err != nil {
		t.Fatalf("failed to set up test user: %v", err)
	}

	attributes := map[string]any{
		"role": "admin",
		"age":  30,
	}
	err = testable.UpdateAttributes(context.Background(), username, attributes)
	if err != nil {
		t.Fatalf("failed to update attributes: %v", err)
	}

	u, err := testable.GetUser(username)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	} else if u == nil {
		t.Fatalf("user is nil")
	}

	aj, _ := json.Marshal(attributes)
	bj, _ := json.Marshal(u.Attributes)
	if !bytes.Equal(aj, bj) {
		t.Fatalf("attributes mismatch: expected '%v', got '%v'", attributes, u.Attributes)
	}

	err = testable.UpdateAttributes(context.Background(), "nonexistent", attributes)
	if err == nil {
		t.Fatalf("존재하지 않는 사용자에 대해 오류가 발생해야 합니다")
	} else if !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("예상치 못한 오류: %v", err)
	}
}

func UserIncreaseRetryCount(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	username := "retryuser"
	err = testable.SetUpUser(username)
	if err != nil {
		t.Fatalf("failed to set up test user: %v", err)
	}

	for i := 1; i <= 3; i++ {
		now := time.Now()
		err = testable.IncreaseRetryCount(context.Background(), username)
		if err != nil {
			t.Fatalf("failed to increase retry count: %v", err)
		}

		u, err := testable.GetUser(username)
		if err != nil {
			t.Fatalf("failed to get user: %v", err)
		} else if u == nil {
			t.Fatalf("user is nil")
		} else if u.Retry != i {
			t.Fatalf("Retry 불일치: got %d, want %d", u.Retry, i)
		} else if u.RetryAt.IsZero() {
			t.Fatalf("RetryAt가 설정되지 않음")
		} else if u.RetryAt.Sub(now).Abs() > maximumAllowedTimeDeltaForNow {
			t.Fatalf("RetryAt 타임스탬프가 최근이 아님: %v", u.RetryAt)
		}
	}

	err = testable.IncreaseRetryCount(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("존재하지 않는 사용자에 대해 오류가 발생하지 않아야 합니다: %v", err)
	}
}

func UserResetRetryCount(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	username := "resetuser"
	err = testable.SetUpUser(username)
	if err != nil {
		t.Fatalf("failed to set up test user: %v", err)
	}

	// Increase retry count a few times first
	for range 3 {
		err = testable.IncreaseRetryCount(context.Background(), username)
		if err != nil {
			t.Fatalf("failed to increase retry count: %v", err)
		}
	}

	err = testable.ResetRetryCount(context.Background(), username)
	if err != nil {
		t.Fatalf("failed to reset retry count: %v", err)
	}

	u, err := testable.GetUser(username)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	} else if u == nil {
		t.Fatalf("user is nil")
	} else if u.Retry != 0 {
		t.Fatalf("Retry 카운트 불일치: got %d, want %d", u.Retry, 0)
	} else if !u.RetryAt.IsZero() {
		t.Fatalf("RetryAt가 초기화되지 않았습니다")
	}

	err = testable.ResetRetryCount(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("존재하지 않는 사용자에 대해 오류가 발생하지 않아야 합니다: %v", err)
	}
}

func UserUpdatePrepare(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	username := "prepareuser"
	err = testable.SetUpUser(username)
	if err != nil {
		t.Fatalf("failed to set up test user: %v", err)
	}

	prepare := []string{"cmd1", "cmd2"}
	err = testable.UpdatePrepare(context.Background(), username, prepare)
	if err != nil {
		t.Fatalf("failed to update prepare: %v", err)
	}

	u, err := testable.GetUser(username)
	if err != nil {
		t.Fatalf("failed to get user: %v", err)
	} else if u == nil {
		t.Fatalf("user is nil")
	}

	if len(u.Prepare) != len(prepare) {
		t.Fatalf("Prepare 길이 불일치: got %d, want %d", len(u.Prepare), len(prepare))
	}
	for i, cmd := range prepare {
		if u.Prepare[i] != cmd {
			t.Fatalf("Prepare 항목 불일치 at index %d: got %s, want %s", i, u.Prepare[i], cmd)
		}
	}

	err = testable.UpdatePrepare(context.Background(), "nonexistent", prepare)
	if err != nil {
		t.Fatalf("존재하지 않는 사용자에 대해 오류가 발생하지 않아야 합니다: %v", err)
	}
}

func UserDeleteByName(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	username := "deleteuser"
	err = testable.SetUpUser(username)
	if err != nil {
		t.Fatalf("failed to set up test user: %v", err)
	}

	var countBefore int
	countBefore, err = testable.GetUserCount()
	if err != nil {
		t.Fatalf("failed to get user count: %v", err)
	}

	err = testable.DeleteByName(context.Background(), username)
	if err != nil {
		t.Fatalf("failed to delete user: %v", err)
	}

	var countAfter int
	countAfter, err = testable.GetUserCount()
	if err != nil {
		t.Fatalf("failed to get user count: %v", err)
	}

	if countAfter != countBefore-1 {
		t.Fatalf("user count did not decrease after deletion: before=%d, after=%d", countBefore, countAfter)
	}

	u, err := testable.GetUser(username)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		t.Fatalf("failed to get user after deletion: %v", err)
	} else if u != nil {
		t.Fatalf("user should be nil after deletion")
	}

	err = testable.DeleteByName(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("존재하지 않는 사용자에 대해 오류가 발생하지 않아야 합니다: %v", err)
	}
}

func UserInsertPasswordHistory(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	username := "historyuser"
	err = testable.SetUpUser(username)
	if err != nil {
		t.Fatalf("failed to set up test user: %v", err)
	}

	password := []byte("password")
	now := time.Now()
	err = testable.InsertPasswordHistory(context.Background(), username, password)
	if err != nil {
		t.Fatalf("Insert 실패: %v", err)
	}

	latestHash, createdAt, err := testable.GetLatestPasswordHistory(username)
	if err != nil {
		t.Fatalf("failed to get password histories: %v", err)
	}

	err = testable.CompareHashedPassword([]byte(latestHash), password)
	if err != nil {
		t.Fatalf("비밀번호 기록 불일치: %v", err)
	}
	if createdAt.Sub(now).Abs() > maximumAllowedTimeDeltaForNow {
		t.Fatalf("CreatedAt 타임스탬프가 최근이 아님: %v", createdAt)
	}

	err = testable.InsertPasswordHistory(context.Background(), "nonexistent", password)
	if err != nil {
		t.Fatalf("존재하지 않는 사용자에 대해 오류가 발생하지 않아야 합니다: %v", err)
	}
}

func UserDeletePasswordHistory(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	username := "delhistoryuser"
	err = testable.SetUpUser(username)
	if err != nil {
		t.Fatalf("failed to set up test user: %v", err)
	}

	passwords := []string{"pass1", "pass2", "pass3"}
	keepN := 2
	numDeleted := len(passwords) - keepN
	if numDeleted <= 0 {
		t.Fatalf("테스트를 위해 삭제할 비밀번호 수가 0 이하입니다")
	}
	var expectedToRemain []string
	for i, p := range passwords {
		if i >= numDeleted {
			expectedToRemain = append(expectedToRemain, p)
		}
		err = testable.InsertPasswordHistory(context.Background(), username, []byte(p))
		if err != nil {
			t.Fatalf("failed to insert password history: %v", err)
		}
		// SQLite + Windows 조합에서 정밀도 한계로 인해 동일한 타임스탬프로 기록되는 문제 방지
		time.Sleep(5 * time.Millisecond)
	}

	err = testable.DeletePasswordHistory(context.Background(), username, int32(keepN))
	if err != nil {
		t.Fatalf("failed to delete password history: %v", err)
	}

	histories, err := testable.GetPasswordHistories(username)
	if err != nil {
		t.Fatalf("failed to get password histories: %v", err)
	} else if len(histories) != keepN {
		t.Fatalf("비밀번호 기록 수 불일치 after deletion: got %d, want %d", len(histories), 1)
	}

	for i, pass := range expectedToRemain {
		err = testable.CompareHashedPassword([]byte(histories[i]), []byte(pass))
		if err != nil {
			t.Fatalf("남아있는 비밀번호 해시 불일치: got %s, want %s", histories[i], pass)
		}
	}
}

func UserListPasswordHistory(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	username := "listhistoryuser"
	err = testable.SetUpUser(username)
	if err != nil {
		t.Fatalf("failed to set up test user: %v", err)
	}

	passwords := []string{"hash1", "hash2", "hash3"}
	for _, p := range passwords {
		err = testable.InsertPasswordHistory(context.Background(), username, []byte(p))
		if err != nil {
			t.Fatalf("failed to insert password history: %v", err)
		}
		// SQLite + Windows 조합에서 정밀도 한계로 인해 동일한 타임스탬프로 기록되는 문제 방지
		time.Sleep(5 * time.Millisecond)
	}

	histories, err := testable.ListPasswordHistory(context.Background(), username)
	if err != nil {
		t.Fatalf("ListPasswordHistory 실패: %v", err)
	} else if len(histories) != len(passwords) {
		t.Fatalf("비밀번호 기록 수 불일치: got %d, want %d", len(histories), len(passwords))
	}

	for i, p := range passwords {
		err = testable.CompareHashedPassword([]byte(histories[i]), []byte(p))
		if err != nil {
			t.Fatalf("비밀번호 해시 불일치 at index %d: got %s, want %s", i, histories[i], p)
		}
	}

	histories, err = testable.ListPasswordHistory(context.Background(), "nonexistent")
	if err != nil {
		t.Fatalf("존재하지 않는 사용자에 대해 오류가 발생하지 않아야 합니다: %v", err)
	} else if len(histories) != 0 {
		t.Fatalf("존재하지 않는 사용자의 비밀번호 기록이 비어 있어야 합니다, got %d records", len(histories))
	}
}

func UserIsPasswordReuse(t *testing.T, testable UserTestable) {
	err := testable.SetUpDb()
	if err != nil {
		t.Fatalf("failed to set up test db: %v", err)
	}
	t.Cleanup(func() {
		if err := testable.TearDownDb(); err != nil {
			t.Errorf("failed to tear down test db: %v", err)
		}
	})

	username := "reuseuser"
	err = testable.SetUpUser(username)
	if err != nil {
		t.Fatalf("failed to set up test user: %v", err)
	}

	passwords := []string{"oldpass1", "oldpass2", "oldpass3"}
	for _, p := range passwords {
		err = testable.InsertPasswordHistory(context.Background(), username, []byte(p))
		if err != nil {
			t.Fatalf("failed to insert password history: %v", err)
		}
		// SQLite + Windows 조합에서 정밀도 한계로 인해 동일한 타임스탬프로 기록되는 문제 방지
		time.Sleep(5 * time.Millisecond)
	}

	tests := []struct {
		password string
		reuse    bool
	}{
		{"oldpass1", true},
		{"oldpass2", true},
		{"oldpass3", true},
		{"newpass", false},
	}

	for _, tt := range tests {
		reused, err := testable.IsPasswordReuse(context.Background(), username, tt.password)
		if err != nil {
			t.Fatalf("IsPasswordReuse 실패 for password %s: %v", tt.password, err)
		} else if reused != tt.reuse {
			t.Fatalf("비밀번호 재사용 감지 불일치 for password %s: got %v, want %v", tt.password, reused, tt.reuse)
		}
	}

	reused, err := testable.IsPasswordReuse(context.Background(), "nonexistent", "somepass")
	if err != nil {
		t.Fatalf("존재하지 않는 사용자에 대해 오류가 발생하지 않아야 합니다: %v", err)
	} else if reused {
		t.Fatalf("존재하지 않는 사용자의 비밀번호는 재사용되지 않아야 합니다")
	}
}
