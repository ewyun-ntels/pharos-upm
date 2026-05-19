package user

import (
	"context"
	"database/sql"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/ory/fosite"
	"ntels.com/pharos/core/external/orm"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/migration/schema"
	"ntels.com/pharos/shared/types/role"
)

// sanitizeTestName converts test name to valid username format
// (lowercase letters, numbers, and underscores only)
func sanitizeTestName(testName string) string {
	// Replace forward slashes with underscores
	name := strings.ReplaceAll(testName, "/", "_")
	// Convert to lowercase
	name = strings.ToLower(name)
	return name
}

type TestableStore interface {
	Store
	reset() error
	setLoginConfig(blockDuration time.Duration, retryLimit int, retryResetTimeout time.Duration)
	setTemporaryBlock(username string, retry int) error
	setPasswordTTL(ttl time.Duration)
	getPasswordHistoryCount(username string) (int, error)
	setPasswordComplexity(minLength, minUppercase, minLowercase, minDigits, minSpecial int)
	setPasswordExpiredRules(rules []string)
}

type TestableRepoStore struct {
	*RepoStore

	db *sql.DB
}

func (s *TestableRepoStore) reset() error {
	if s.db != nil {
		var err error
		_, err = s.db.Exec("DELETE FROM users WHERE true")
		if err != nil {
			return err
		}
		_, err = s.db.Exec("DELETE FROM user_password_history WHERE true")
		if err != nil {
			return err
		}
	}
	return nil
}

func (s *TestableRepoStore) setLoginConfig(blockDuration time.Duration, retryLimit int, retryResetTimeout time.Duration) {
	s.LoginBlockDuration = blockDuration
	s.LoginRetryLimit = retryLimit
	s.LoginRetryResetTimeout = retryResetTimeout
}
func (s *TestableRepoStore) setTemporaryBlock(username string, retry int) error {
	ctx := context.Background()
	err := s.repo.ResetRetryCount(ctx, username)
	if err != nil {
		return err
	}
	for range retry {
		err = s.repo.IncreaseRetryCount(ctx, username)
		if err != nil {
			return err
		}
	}
	return nil
}
func (s *TestableRepoStore) setPasswordTTL(ttl time.Duration) {
	s.PasswordTTL = ttl
}
func (s *TestableRepoStore) getPasswordHistoryCount(username string) (int, error) {
	ctx := context.Background()
	history, err := s.repo.ListPasswordHistory(ctx, username)
	if err != nil {
		return 0, err
	}
	return len(history), nil
}
func (s *TestableRepoStore) setPasswordComplexity(minLength, minUppercase, minLowercase, minDigits, minSpecial int) {
	s.PasswordValidator = newPasswordValidator(StoreConfig{
		PasswordMinLength:    minLength,
		PasswordMinUppercase: minUppercase,
		PasswordMinLowercase: minLowercase,
		PasswordMinDigits:    minDigits,
		PasswordMinSpecial:   minSpecial,
	})
}
func (s *TestableRepoStore) setPasswordExpiredRules(rules []string) {
	s.PasswordExpiredRule = rules
}

func ListUsers(t *testing.T, testable TestableStore) {
	t.Helper()
	// 사용자명은 소문자, 숫자, 언더스코어만 허용
	prefix := sanitizeTestName(t.Name())
	users := []struct {
		username string
		password string
	}{
		{prefix + "_user1", "pass1"},
		{prefix + "_user2", "pass2"},
		{prefix + "_user3", "pass3"},
	}

	for _, u := range users {
		err := testable.CreateUser(u.username, u.password, nil, nil)
		if err != nil {
			t.Fatalf("CreateUser failed for %s: %v", u.username, err)
		}
	}

	retrievedUsers, err := testable.ListUsers()
	if err != nil {
		t.Fatalf("ListUsers failed: %v", err)
	}
	if len(retrievedUsers) < len(users) {
		t.Fatalf("사용자 수 부족: got %d, want at least %d", len(retrievedUsers), len(users))
	}

	userMap := make(map[string]bool)
	for _, user := range retrievedUsers {
		userMap[user.GetID()] = true
	}

	for _, user := range users {
		if !userMap[user.username] {
			t.Fatalf("사용자 %s가 누락되었습니다", user.username)
		}
	}
}
func FindByUsername(t *testing.T, testable TestableStore) {
	t.Helper()
	username := sanitizeTestName(t.Name()) + "_testuser"
	password := "testpass"

	err := testable.CreateUser(username, password, nil, nil)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	retrievedUser, err := testable.FindByUsername(username)
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}

	if retrievedUser.GetID() != username {
		t.Fatalf("Username mismatch: got %s, want %s", retrievedUser.GetID(), username)
	}

	_, err = testable.FindByUsername("nonexistent")
	if err == nil {
		t.Fatalf("존재하지 않는 사용자에 대해 오류가 발생해야 합니다")
	}

	if !errors.Is(err, fosite.ErrNotFound) {
		t.Fatalf("예상치 못한 오류: %v", err)
	}
}
func CreateUser(t *testing.T, testable TestableStore) {
	t.Helper()
	username := sanitizeTestName(t.Name()) + "_newuser"
	password := "securepassword"
	attributes := map[string]any{"key": "value"}

	err := testable.CreateUser(username, password, attributes, nil)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	retrievedUser, err := testable.FindByUsername(username)
	if err != nil {
		t.Fatalf("사용자가 데이터베이스에 생성되지 않았습니다")
	}

	if retrievedUser.GetID() != username {
		t.Fatalf("Username mismatch: got %s, want %s", retrievedUser.GetID(), username)
	}

	extra := retrievedUser.GetExtra()
	for k, v := range attributes {
		if extra[k] != v {
			t.Fatalf("Attributes mismatch for key %s: got %v, want %v", k, extra[k], v)
		}
	}

	err = testable.Authenticate(username, password)
	if err != nil {
		t.Fatalf("Authenticate failed: %v", err)
	}

	err = testable.CreateUser(username, password, attributes, nil)
	if err == nil {
		t.Fatalf("중복된 사용자 생성에 대해 오류가 발생해야 합니다")
	}

	if !errors.Is(err, ErrUserAlreadyExists) {
		t.Fatalf("예상치 못한 오류: %v", err)
	}

	prepare := []string{"password_change"}
	prepareUser := sanitizeTestName(t.Name()) + "_prepuser"
	err = testable.CreateUser(prepareUser, password, attributes, &prepare)
	if err != nil {
		t.Fatalf("failed to create user with prepare: %v", err)
	}

	retrievedUser, err = testable.FindByUsername(prepareUser)
	if err != nil {
		t.Fatalf("failed to retrieve user with prepare: %v", err)
	}

	for i, prep := range retrievedUser.GetPrepare() {
		if prep != prepare[i] {
			t.Fatalf("Prepare mismatch at index %d: got %s, want %s", i, prep, prepare[i])
		}
	}
}
func DeleteUser(t *testing.T, testable TestableStore) {
	t.Helper()
	username := sanitizeTestName(t.Name()) + "_tobedeleted"
	err := testable.CreateUser(username, "somepassword", nil, nil)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	err = testable.DeleteUser(username)
	if err != nil {
		t.Fatalf("DeleteUser failed: %v", err)
	}

	_, err = testable.FindByUsername(username)
	if err == nil {
		t.Fatalf("삭제된 사용자가 여전히 검색 가능합니다")
	} else if !errors.Is(err, fosite.ErrNotFound) {
		t.Fatalf("예상치 못한 오류: %v", err)
	}

	// 존재하지 않는 사용자 삭제 시도
	err = testable.DeleteUser("nonexistent")
	if err == nil {
		t.Fatalf("존재하지 않는 사용자 삭제에 대해 오류가 발생해야 합니다")
	}

	if !errors.Is(err, fosite.ErrNotFound) {
		t.Fatalf("예상치 못한 오류: %v", err)
	}
}
func ChangePassword(t *testing.T, testable TestableStore) {
	t.Helper()
	username := sanitizeTestName(t.Name()) + "_changepassuser"
	oldPassword := "oldpassword"
	newPassword := "newsecurepassword"

	testable.setPasswordTTL(time.Hour)

	err := testable.CreateUser(username, oldPassword, nil, nil)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	err = testable.ChangePassword(username, newPassword)
	if err != nil {
		t.Fatalf("ChangePassword failed: %v", err)
	}

	err = testable.Authenticate(username, newPassword)
	if err != nil {
		t.Fatalf("새 비밀번호로 인증에 실패했습니다: %v", err)
	}

	err = testable.Authenticate(username, oldPassword)
	if err == nil {
		t.Fatalf("이전 비밀번호로 인증이 성공해서는 안 됩니다")
	}

	retrievedUser, err := testable.FindByUsername(username)
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}

	if retrievedUser.GetPasswordExpiresAt() == nil {
		t.Fatalf("PasswordExpiredAt should be set after password change")
	}

	err = testable.ChangePassword("nonexistent", newPassword)
	if err == nil {
		t.Fatalf("존재하지 않는 사용자에 대해 ChangePassword가 오류를 발생시켜야 합니다")
	}
}
func SetAttributes(t *testing.T, testable TestableStore) {
	t.Helper()
	username := sanitizeTestName(t.Name()) + "_attruser"

	err := testable.CreateUser(username, "somepassword", nil, nil)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	attributes := map[string]any{"role": "admin", "active": true}
	err = testable.SetAttributes(username, attributes)
	if err != nil {
		t.Fatalf("SetAttributes failed: %v", err)
	}

	retrievedUser, err := testable.FindByUsername(username)
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}

	retrievedExtra := retrievedUser.GetExtra()
	for key, value := range attributes {
		if retrievedExtra[key] != value {
			t.Fatalf("Attribute %s mismatch: got %v, want %v", key, retrievedExtra[key], value)
		}
	}

	err = testable.SetAttributes("nonexistent", attributes)
	if err == nil {
		t.Fatalf("존재하지 않는 사용자에 대해 SetAttributes가 오류를 발생시켜야 합니다")
	}
}
func DecodeExtra(t *testing.T, testable TestableStore) {
	t.Helper()

	// Note: Composite role expansion is now handled upstream by MetadataService
	// DecodeExtra should preserve roles as-is without expansion

	// Nested structure: {"roles": {...}, "info": {...}}
	extra := map[string]any{
		"roles": map[string]any{
			"superadmin": true,
			"viewer":     false,
		},
		"info": map[string]any{
			"department": "IT",
		},
	}

	decoded, err := testable.DecodeExtra(extra)
	if err != nil {
		t.Fatalf("DecodeExtra failed: %v", err)
	}

	// Check roles are preserved as-is (no expansion)
	roles, ok := decoded["roles"].(map[string]any)
	if !ok {
		t.Fatalf("roles should be a map[string]interface{}")
	}

	// Roles should be unchanged (expansion happens upstream)
	if roles["superadmin"] != true {
		t.Fatalf("superadmin role should be preserved: %v", roles)
	}

	if roles["viewer"] != false {
		t.Fatalf("viewer role should be preserved: %v", roles)
	}

	// Check info preserved
	info, ok := decoded["info"].(map[string]any)
	if !ok {
		t.Fatalf("info should be a map[string]interface{}")
	}

	if info["department"] != "IT" {
		t.Fatalf("Non-role attribute should be preserved: %v", info)
	}
}
func SetBlock(t *testing.T, testable TestableStore) {
	t.Helper()
	username := sanitizeTestName(t.Name()) + "_blockeduser"

	err := testable.CreateUser(username, "somepassword", nil, nil)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	err = testable.SetBlock(username, true)
	if err != nil {
		t.Fatalf("SetBlock to true failed: %v", err)
	}

	retrievedUser, err := testable.FindByUsername(username)
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}

	if !retrievedUser.GetBlocked() {
		t.Fatalf("Blocked 상태가 true로 설정되지 않았습니다")
	}

	err = testable.SetBlock(username, false)
	if err != nil {
		t.Fatalf("SetBlock to false failed: %v", err)
	}

	retrievedUser, err = testable.FindByUsername(username)
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}

	if retrievedUser.GetBlocked() {
		t.Fatalf("Blocked 상태가 false로 설정되지 않았습니다")
	}

	err = testable.SetBlock("nonexistent", true)
	if err == nil {
		t.Fatalf("존재하지 않는 사용자에 대해 SetBlock이 오류를 발생시켜야 합니다")
	}
}
func Authenticate(t *testing.T, testable TestableStore) {
	t.Helper()
	username := sanitizeTestName(t.Name()) + "_authuser"
	password := "authpassword"
	wrongPassword := "wrongpassword"

	blockDuration := 5 * time.Second
	retryLimit := 3
	retryResetTimeout := 10 * time.Second
	testable.setLoginConfig(blockDuration, retryLimit, retryResetTimeout)

	err := testable.CreateUser(username, password, nil, nil)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	err = testable.Authenticate("nonexistent", "")
	if err == nil {
		t.Fatalf("존재하지 않는 사용자에 대해 Authenticate가 오류를 발생시켜야 합니다")
	} else if !errors.Is(err, fosite.ErrNotFound) {
		t.Fatalf("예상치 못한 오류: %v", err)
	}

	err = testable.SetBlock(username, true)
	if err != nil {
		t.Fatalf("SetBlock failed: %v", err)
	}

	err = testable.Authenticate(username, password)
	if err == nil {
		t.Fatalf("차단된 사용자에 대해 Authenticate가 오류를 발생시켜야 합니다")
	} else if !errors.Is(err, fosite.ErrInvalidRequest) {
		t.Fatalf("예상치 못한 오류: %v", err)
	}

	err = testable.SetBlock(username, false)
	if err != nil {
		t.Fatalf("SetBlock failed: %v", err)
	}

	// 차단된 사용자가 skip_temporarily_block 속성을 가질 때 인증이 성공해야 함
	err = testable.SetAttributes(username, map[string]any{
		string(role.AttrSkipTemporarilyBlock): true,
	})
	if err != nil {
		t.Fatalf("SetAttributes failed: %v", err)
	}
	err = testable.setTemporaryBlock(username, retryLimit)
	if err != nil {
		t.Fatalf("setTemporaryBlock failed: %v", err)
	}
	err = testable.Authenticate(username, password)
	if err != nil {
		t.Fatalf("skip_temporarily_block 속성이 있는 차단된 사용자의 인증이 실패했습니다: %v", err)
	}

	// password_retry_unlimited 속성을 가질 때 인증이 성공해야 함
	err = testable.SetAttributes(username, map[string]any{
		string(role.AttrPasswordRetryUnlimited): true,
	})
	if err != nil {
		t.Fatalf("SetAttributes failed: %v", err)
	}
	err = testable.setTemporaryBlock(username, retryLimit)
	if err != nil {
		t.Fatalf("setTemporaryBlock failed: %v", err)
	}
	err = testable.Authenticate(username, password)
	if err != nil {
		t.Fatalf("password_retry_unlimited 속성이 있는 사용자의 인증이 실패했습니다: %v", err)
	}

	// 속성이 없을 때 차단된 사용자는 인증이 실패해야 함
	err = testable.SetAttributes(username, nil)
	if err != nil {
		t.Fatalf("SetAttributes failed: %v", err)
	}
	err = testable.setTemporaryBlock(username, retryLimit)
	if err != nil {
		t.Fatalf("setTemporaryBlock failed: %v", err)
	}
	err = testable.Authenticate(username, password)
	if err == nil {
		t.Fatalf("일시적으로 차단된 사용자에 대해 Authenticate가 오류를 발생시켜야 합니다")
	} else if !errors.Is(err, fosite.ErrInvalidRequest) {
		t.Fatalf("예상치 못한 오류: %v", err)
	}

	// 시도 횟수가 제한에 미치지 못할 때 인증이 성공해야 함
	err = testable.setTemporaryBlock(username, retryLimit-1)
	if err != nil {
		t.Fatalf("unsetTemporaryBlock failed: %v", err)
	}
	err = testable.Authenticate(username, password)
	if err != nil {
		t.Fatalf("정상 사용자에 대해 Authenticate가 실패했습니다: %v", err)
	}

	// 재시도 카운트가 초기화되었는지 확인
	user, err := testable.FindByUsername(username)
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}
	if user.HasRetry() {
		t.Fatalf("정상 사용자에 대해 인증 후 재시도 카운트가 초기화되지 않았습니다")
	}

	blockDuration = 1 * time.Millisecond
	retryLimit = 3
	retryResetTimeout = 1 * time.Millisecond
	testable.setLoginConfig(blockDuration, retryLimit, retryResetTimeout)
	err = testable.setTemporaryBlock(username, retryLimit)
	if err != nil {
		t.Fatalf("setTemporaryBlock failed: %v", err)
	}
	time.Sleep(blockDuration + time.Millisecond)
	err = testable.Authenticate(username, password)
	if err != nil {
		t.Fatalf("재시도 타임아웃이 지난 사용자에 대해 Authenticate가 실패했습니다: %v", err)
	}

	err = testable.setTemporaryBlock(username, retryLimit)
	if err != nil {
		t.Fatalf("setTemporaryBlock failed: %v", err)
	}
	time.Sleep(retryResetTimeout + time.Millisecond)
	err = testable.Authenticate(username, wrongPassword)
	if err == nil {
		t.Fatalf("잘못된 비밀번호에 대해 Authenticate가 오류를 발생시켜야 합니다")
	}
	retrievedUser, err := testable.FindByUsername(username)
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}
	if retrievedUser.IsTemporarilyBlocked(retryLimit, blockDuration) {
		t.Fatalf("재시도 타임아웃이 지난 사용자에 대해 인증 후에도 일시적으로 차단 상태입니다")
	}
}
func InsertPasswordHistory(t *testing.T, testable TestableStore) {
	t.Helper()
	username := sanitizeTestName(t.Name()) + "_historyuser"
	err := testable.InsertPasswordHistory(username, "newpassword")
	if err != nil {
		t.Fatalf("InsertPasswordHistory failed: %v", err)
	}

	// 추가 검증은 CheckPasswordHistory 테스트에서 수행
}
func DeletePasswordHistory(t *testing.T, testable TestableStore) {
	t.Helper()
	username := sanitizeTestName(t.Name()) + "_historyuser"
	passwords := []string{"oldpassword1", "oldpassword2", "oldpassword3"}
	for _, pwd := range passwords {
		err := testable.InsertPasswordHistory(username, pwd)
		if err != nil {
			t.Fatalf("InsertPasswordHistory failed for %s: %v", pwd, err)
		}
	}

	expectedCountAfterDelete := 1
	err := testable.DeletePasswordHistory(username, int32(expectedCountAfterDelete))
	if err != nil {
		t.Fatalf("DeletePasswordHistory failed: %v", err)
	}

	count, err := testable.getPasswordHistoryCount(username)
	if err != nil {
		t.Fatalf("getPasswordHistoryCount failed: %v", err)
	}
	if count != expectedCountAfterDelete {
		t.Fatalf("Password history count mismatch after deletion: got %d, want %d", count, expectedCountAfterDelete)
	}
}
func CheckPasswordHistory(t *testing.T, testable TestableStore) {
	t.Helper()
	username := sanitizeTestName(t.Name()) + "_historyuser"
	passwords := []string{"oldpassword1", "oldpassword2", "oldpassword3"}
	for _, pwd := range passwords {
		err := testable.InsertPasswordHistory(username, pwd)
		if err != nil {
			t.Fatalf("InsertPasswordHistory failed for %s: %v", pwd, err)
		}
	}

	err := testable.CheckPasswordHistory(username, "newpassword")
	if err != nil {
		t.Fatalf("CheckPasswordHistory for new password failed: %v", err)
	}

	for _, oldPwd := range passwords {
		err = testable.CheckPasswordHistory(username, oldPwd)
		if err == nil {
			t.Fatalf("CheckPasswordHistory should fail for old password: %s", oldPwd)
		}
		if err.Error() != "password has been used before" {
			t.Fatalf("Unexpected error for old password %s: %v", oldPwd, err)
		}
	}
}
func UpdatePasswordExpiration(t *testing.T, testable TestableStore) {
	t.Helper()
	username := sanitizeTestName(t.Name()) + "_expireuser"

	err := testable.CreateUser(username, "somepassword", nil, nil)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	newExpiry := time.Now().Add(24 * time.Hour)
	err = testable.UpdatePasswordExpiration(username, &newExpiry)
	if err != nil {
		t.Fatalf("UpdatePasswordExpiration failed: %v", err)
	}

	retrievedUser, err := testable.FindByUsername(username)
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}

	if retrievedUser.GetPasswordExpiresAt() == nil || retrievedUser.GetPasswordExpiresAt().Sub(newExpiry).Abs() > time.Second {
		t.Fatalf("PasswordExpiresAt mismatch: got %v, want %v", retrievedUser.GetPasswordExpiresAt(), newExpiry)
	}

	// 만료 날짜를 nil로 설정
	err = testable.UpdatePasswordExpiration(username, nil)
	if err != nil {
		t.Fatalf("UpdatePasswordExpiration to nil failed: %v", err)
	}

	retrievedUser, err = testable.FindByUsername(username)
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}
	if retrievedUser.GetPasswordExpiresAt() != nil {
		t.Fatalf("PasswordExpiresAt should be nil after setting to nil, got %v", retrievedUser.GetPasswordExpiresAt())
	}

	err = testable.UpdatePasswordExpiration("nonexistent", &newExpiry)
	if err == nil {
		t.Fatalf("존재하지 않는 사용자에 대해 UpdatePasswordExpiration이 오류를 발생시켜야 합니다")
	} else if !errors.Is(err, fosite.ErrNotFound) {
		t.Fatalf("예상치 못한 오류: %v", err)
	}
}
func UpdatePrepare(t *testing.T, testable TestableStore) {
	t.Helper()
	username := sanitizeTestName(t.Name()) + "_prepareuser"

	err := testable.CreateUser(username, "somepassword", nil, nil)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	prepare := []string{"step1", "step2"}
	err = testable.UpdatePrepare(username, prepare)
	if err != nil {
		t.Fatalf("UpdatePrepare failed: %v", err)
	}

	retrievedUser, err := testable.FindByUsername(username)
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}
	retrievedPrepare := retrievedUser.GetPrepare()
	if len(retrievedPrepare) != len(prepare) {
		t.Fatalf("Prepare length mismatch: got %d, want %d", len(retrievedPrepare), len(prepare))
	}
	for i, prep := range prepare {
		if retrievedPrepare[i] != prep {
			t.Fatalf("Prepare mismatch at index %d: got %s, want %s", i, retrievedPrepare[i], prep)
		}
	}

	err = testable.UpdatePrepare(username, []string{})
	if err != nil {
		t.Fatalf("UpdatePrepare failed: %v", err)
	}
	retrievedUser, err = testable.FindByUsername(username)
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}
	retrievedPrepare = retrievedUser.GetPrepare()
	if len(retrievedPrepare) != 0 {
		t.Fatalf("Prepare should be empty: got %v", retrievedPrepare)
	}

	err = testable.UpdatePrepare(username, nil)
	if err != nil {
		t.Fatalf("UpdatePrepare failed: %v", err)
	}
	retrievedUser, err = testable.FindByUsername(username)
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}
	retrievedPrepare = retrievedUser.GetPrepare()
	if len(retrievedPrepare) != 0 {
		t.Fatalf("Prepare should be empty: got %v", retrievedPrepare)
	}

	err = testable.UpdatePrepare("nonexistent", []string{"step1", "step2"})
	if err == nil {
		t.Fatalf("존재하지 않는 사용자에 대해 UpdatePrepare가 오류를 발생시켜야 합니다")
	} else if !errors.Is(err, fosite.ErrNotFound) {
		t.Fatalf("예상치 못한 오류: %v", err)
	}
}
func Close(t *testing.T, _ TestableStore) {
	t.Helper()
	t.Skip("no operation to test")
}
func ValidatePasswordComplexity(t *testing.T, testable TestableStore) {
	t.Helper()
	minLength := 8
	minUppercase := 1
	minLowercase := 1
	minDigits := 1
	minSpecial := 1
	testable.setPasswordComplexity(minLength, minUppercase, minLowercase, minDigits, minSpecial)

	var err error
	validPasswords := []string{"Password1!", "Complex#123", "Valid$Pass9"}
	invalidPasswords := []string{"short1!", "nouppercase1!", "NOLOWERCASE1!", "NoDigits!", "NoSpecial1", "NoSpecialChar1"}

	for _, pwd := range validPasswords {
		err = testable.ValidatePasswordComplexity(pwd)
		if err != nil {
			t.Fatalf("Expected password %s to be valid, but got error: %v", pwd, err)
		}
	}

	for _, pwd := range invalidPasswords {
		err = testable.ValidatePasswordComplexity(pwd)
		if err == nil {
			t.Fatalf("Expected password %s to be invalid", pwd)
		}
	}
}
func GetLoginRetryLimit(t *testing.T, _ TestableStore) {
	t.Helper()
	t.Skip("no operation to test")
}
func GetLoginBlockDuration(t *testing.T, _ TestableStore) {
	t.Helper()
	t.Skip("no operation to test")
}
func GetPrepare(t *testing.T, testable TestableStore) {
	t.Helper()
	username := sanitizeTestName(t.Name()) + "_prepareuser"

	ttl := time.Minute
	testable.setPasswordTTL(ttl)
	prepare := []string{"step1", "step2"}

	err := testable.CreateUser(username, "somepassword", nil, &prepare)
	if err != nil {
		t.Fatalf("CreateUser failed: %v", err)
	}

	retrievedUser, err := testable.FindByUsername(username)
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}
	retrievedPrepare, err := testable.GetPrepare(retrievedUser)
	if err != nil {
		t.Fatalf("GetPrepare failed: %v", err)
	}
	if len(retrievedPrepare) != len(prepare) {
		t.Fatalf("Prepare length mismatch: got %d, want %d", len(retrievedPrepare), len(prepare))
	}
	for i, prep := range prepare {
		if retrievedPrepare[i] != prep {
			t.Fatalf("Prepare mismatch at index %d: got %s, want %s", i, retrievedPrepare[i], prep)
		}
	}

	past := time.Now().Add(-ttl)
	err = testable.UpdatePasswordExpiration(username, &past)
	if err != nil {
		t.Fatalf("UpdatePasswordExpiration failed: %v", err)
	}

	retrievedUser, err = testable.FindByUsername(username)
	if err != nil {
		t.Fatalf("FindByUsername failed: %v", err)
	}
	retrievedPrepare, err = testable.GetPrepare(retrievedUser)
	// after password expired, prepare should be the same with store's rule
	// if store's rule is empty, GetPrepare returns error
	if err == nil {
		t.Fatalf("GetPrepare should fail after password expired when store's rule is empty, got %v", retrievedPrepare)
	}

	rules := []string{"password_change"}
	testable.setPasswordExpiredRules(rules)
	retrievedPrepare, err = testable.GetPrepare(retrievedUser)
	if err != nil {
		t.Fatalf("GetPrepare failed after setting rules: %v", err)
	}
	if len(retrievedPrepare) != len(rules) {
		t.Fatalf("Prepare length mismatch after setting rules: got %d, want %d", len(retrievedPrepare), len(rules))
	}
	for i, prep := range rules {
		if retrievedPrepare[i] != prep {
			t.Fatalf("Prepare mismatch at index %d after setting rules: got %s, want %s", i, retrievedPrepare[i], prep)
		}
	}
}

func TestRepoStore(t *testing.T) {
	dbConfig := orm.DatabaseConfig{
		Driver: "sqlite",
		SQLite: orm.SQLiteConfig{
			Path: "file:sharedmem?mode=memory&cache=shared",
		},
	}

	db, err := sql.Open(dbConfig.Driver, dbConfig.SQLite.Path)
	if err != nil {
		t.Fatalf("failed to open database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	var keepConn *sql.Conn
	keepConn, err = db.Conn(context.Background())
	if err != nil {
		t.Fatalf("failed to get connection: %v", err)
	}
	t.Cleanup(func() {
		_ = keepConn.Close()
	})

	schema.DisableMigrateLog()
	err = schema.UpMasterBase(common.Config{
		Serve:    common.ServeConfig{},
		Database: dbConfig,
	})
	if err != nil {
		t.Fatalf("failed to run migrations: %v", err)
	}

	store, err := NewRepoStore(StoreConfig{
		Database: dbConfig,
	})
	if err != nil {
		t.Fatalf("NewRepoStore failed: %v", err)
	}

	testable := &TestableRepoStore{
		RepoStore: store,
		db:        db,
	}

	t.Run("ListUsers", func(t *testing.T) {
		ListUsers(t, testable)
	})
	t.Run("FindByUsername", func(t *testing.T) {
		FindByUsername(t, testable)
	})
	t.Run("CreateUser", func(t *testing.T) {
		CreateUser(t, testable)
	})
	t.Run("DeleteUser", func(t *testing.T) {
		DeleteUser(t, testable)
	})
	t.Run("ChangePassword", func(t *testing.T) {
		ChangePassword(t, testable)
	})
	t.Run("SetAttributes", func(t *testing.T) {
		SetAttributes(t, testable)
	})
	t.Run("DecodeExtra", func(t *testing.T) {
		DecodeExtra(t, testable)
	})
	t.Run("SetBlock", func(t *testing.T) {
		SetBlock(t, testable)
	})
	t.Run("Authenticate", func(t *testing.T) {
		Authenticate(t, testable)
	})
	t.Run("InsertPasswordHistory", func(t *testing.T) {
		InsertPasswordHistory(t, testable)
	})
	t.Run("DeletePasswordHistory", func(t *testing.T) {
		DeletePasswordHistory(t, testable)
	})
	t.Run("CheckPasswordHistory", func(t *testing.T) {
		CheckPasswordHistory(t, testable)
	})
	t.Run("UpdatePasswordExpiration", func(t *testing.T) {
		UpdatePasswordExpiration(t, testable)
	})
	t.Run("UpdatePrepare", func(t *testing.T) {
		UpdatePrepare(t, testable)
	})
	t.Run("Close", func(t *testing.T) {
		Close(t, testable)
	})
	t.Run("ValidatePasswordComplexity", func(t *testing.T) {
		ValidatePasswordComplexity(t, testable)
	})
	t.Run("GetLoginRetryLimit", func(t *testing.T) {
		GetLoginRetryLimit(t, testable)
	})
	t.Run("GetLoginBlockDuration", func(t *testing.T) {
		GetLoginBlockDuration(t, testable)
	})
	t.Run("GetPrepare", func(t *testing.T) {
		GetPrepare(t, testable)
	})
}
