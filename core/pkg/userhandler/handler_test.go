package userhandler

import (
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite/token/jwt"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/internal/user"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/shared/types/role"
)

type fakeUser struct {
	id      string
	blocked bool
	created time.Time
	extra   map[string]any
	prepare []string
	pwdExp  *time.Time
	retry   int
	retryAt *time.Time
}

func (f *fakeUser) GetID() string                    { return f.id }
func (f *fakeUser) GetBlocked() bool                 { return f.blocked }
func (f *fakeUser) GetExtra() map[string]any         { return f.extra }
func (f *fakeUser) GetCreatedAt() time.Time          { return f.created }
func (f *fakeUser) GetPasswordExpiresAt() *time.Time { return f.pwdExp }
func (f *fakeUser) GetPrepare() []string             { return f.prepare }
func (f *fakeUser) GetTemporaryBlockedExpiresAt(limit int, duration time.Duration) *time.Time {
	if f.retry >= limit && f.retryAt != nil {
		exp := f.retryAt.Add(duration)
		return &exp
	}
	return nil
}
func (f *fakeUser) IsRetryAtTimedOut(timeout time.Duration) bool {
	if f.retryAt == nil {
		return false
	}
	return time.Now().After(f.retryAt.Add(timeout))
}
func (f *fakeUser) IsTemporarilyBlocked(limit int, duration time.Duration) bool {
	return f.retry >= limit && f.retryAt != nil && time.Now().Before(f.retryAt.Add(duration))
}
func (f *fakeUser) HasRetry() bool {
	return f.retry != 0 && f.retryAt != nil
}

type fakeStore struct {
	users            map[string]user.User
	retryLimit       int
	blockDuration    time.Duration
	effectivePrepare map[string][]string
}

func (s *fakeStore) ListUsers() ([]user.User, error) {
	arr := make([]user.User, 0, len(s.users))
	for _, u := range s.users {
		arr = append(arr, u)
	}
	return arr, nil
}
func (s *fakeStore) FindByUsername(username string) (user.User, error) {
	u, ok := s.users[username]
	if !ok {
		return nil, errors.New("not found")
	}
	return u, nil
}
func (s *fakeStore) CreateUser(string, string, map[string]any, *[]string) error { return nil }
func (s *fakeStore) DeleteUser(string) error                                    { return nil }
func (s *fakeStore) ChangePassword(string, string) error                        { return nil }
func (s *fakeStore) SetAttributes(string, map[string]any) error                 { return nil }
func (s *fakeStore) DecodeExtra(attrs map[string]any) (map[string]any, error) {
	return attrs, nil
}
func (s *fakeStore) SetBlock(string, bool) error                       { return nil }
func (s *fakeStore) Authenticate(string, string) error                 { return nil }
func (s *fakeStore) InsertPasswordHistory(string, string) error        { return nil }
func (s *fakeStore) DeletePasswordHistory(string, int32) error         { return nil }
func (s *fakeStore) CheckPasswordHistory(string, string) error         { return nil }
func (s *fakeStore) UpdatePasswordExpiration(string, *time.Time) error { return nil }
func (s *fakeStore) UpdatePrepare(string, []string) error              { return nil }
func (s *fakeStore) Close() error                                      { return nil }
func (s *fakeStore) ValidatePasswordComplexity(string) error           { return nil }
func (s *fakeStore) GetLoginRetryLimit() int                           { return s.retryLimit }
func (s *fakeStore) GetLoginBlockDuration() time.Duration              { return s.blockDuration }
func (s *fakeStore) GetPrepare(u user.User) ([]string, error) {
	if s.effectivePrepare != nil {
		if prepare, ok := s.effectivePrepare[u.GetID()]; ok {
			return prepare, nil
		}
	}
	return u.GetPrepare(), nil
}

// helper to build gin with our routes but bypassing auth middleware by directly invoking handler methods
func newTestRouter(s *fakeStore) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	// test middleware to inject JWT claims from headers
	r.Use(func(c *gin.Context) {
		sub := c.GetHeader("X-Subject")
		claims := &jwt.JWTClaims{Subject: sub}

		// JWTExtra 타입을 사용하여 명확한 구조 정의
		extra := &common.JWTExtra{
			Roles: make(map[string]bool),
			Info:  make(map[string]any),
		}

		if c.GetHeader("X-Role-Super-Admin") == "true" {
			extra.Roles[string(role.RoleSuperAdmin)] = true
		}
		if c.GetHeader("X-Role-Manage-User") == "true" {
			// Manage user includes all CRUD permissions
			extra.Roles[string(role.RoleUserCreate)] = true
			extra.Roles[string(role.RoleUserRead)] = true
			extra.Roles[string(role.RoleUserUpdate)] = true
			extra.Roles[string(role.RoleUserDelete)] = true
		}

		extraMap, err := extra.ToMap()
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to convert extra"})
			return
		}

		claims.Extra = extraMap
		c.Set(common.ContextKeyJWTClaims, claims)
		c.Next()
	})
	// mount directly without auth middlewares by binding endpoints to handler methods
	h := NewHandler(s, nil, nil).(*handler)
	r.GET("/me", func(c *gin.Context) { h.GetMe(c) })
	r.GET("/user", func(c *gin.Context) { h.GetUserList(c) })
	r.GET("/user/:username", func(c *gin.Context) { h.GetUser(c) })
	return r
}

func Test_getUserDetail(t *testing.T) {
	now := time.Now().Add(-time.Hour)
	retryAt := time.Now().Add(-10 * time.Minute)
	// DB 구조와 동일하게: {"roles": {...}, "info": {...}}
	extra := map[string]any{
		"roles": map[string]any{},
		"info":  map[string]any{"k": "v"},
	}
	fu := &fakeUser{id: "alice", blocked: true, created: now, extra: extra, prepare: []string{"p"}, pwdExp: nil, retry: 5, retryAt: &retryAt}
	d := getUserDetail(fu, 3, 30*time.Minute, nil)
	require.Equal(t, "alice", d.Name)
	require.Equal(t, true, *d.Blocked)
	require.NotNil(t, d.CreatedAt)
	require.NotNil(t, d.Attributes)
	require.Equal(t, "v", d.Attributes.Info["k"])
	require.NotNil(t, d.TemporaryBlockedExpiresAt)
}

func Test_getUserDetail_WithCompositeRole(t *testing.T) {
	// role service가 없는 경우
	now := time.Now()
	fu := &fakeUser{id: "admin", created: now, extra: map[string]any{"admin_role": true}}
	d := getUserDetail(fu, 3, 30*time.Minute, nil)
	require.Equal(t, "admin", d.Name)
	require.Nil(t, d.Roles) // roleService가 nil이므로 roles도 nil

	// 추후 role service가 있는 경우 테스트 추가 가능
}

func Test_getEffectiveUserDetail_UsesStorePrepare(t *testing.T) {
	now := time.Now()
	fu := &fakeUser{id: "expired", created: now, extra: map[string]any{
		"roles": map[string]any{},
		"info":  map[string]any{},
	}}
	s := &fakeStore{
		users:            map[string]user.User{"expired": fu},
		retryLimit:       3,
		blockDuration:    30 * time.Minute,
		effectivePrepare: map[string][]string{"expired": {user.PreparePasswordChange}},
	}
	h := NewHandler(s, nil, nil).(*handler)

	d := h.getEffectiveUserDetail(fu)
	require.Equal(t, []string{user.PreparePasswordChange}, d.Prepare)
	require.Equal(t, []string{"Password change required"}, d.PrepareLabels)
}

func Test_GetMe_and_GetUserList_Basic(t *testing.T) {
	// seed store with two users: alice (subject) non-admin, bob regular, root is super_admin
	now := time.Now()
	retryAt := time.Now()
	alice := &fakeUser{id: "alice", created: now, extra: map[string]any{
		"roles": map[string]any{},
		"info":  map[string]any{},
	}, retry: 0, retryAt: &retryAt}
	bob := &fakeUser{id: "bob", created: now, extra: map[string]any{
		"roles": map[string]any{},
		"info":  map[string]any{},
	}, retry: 0}
	root := &fakeUser{id: "root", created: now, extra: map[string]any{
		"roles": map[string]any{string(role.RoleSuperAdmin): true},
		"info":  map[string]any{},
	}, retry: 0}
	s := &fakeStore{users: map[string]user.User{"alice": alice, "bob": bob, "root": root}, retryLimit: 3, blockDuration: 30 * time.Minute}
	r := newTestRouter(s)

	// call GET /me with Subject=alice
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("X-Subject", "alice")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var details Details
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &details))
	require.Equal(t, "alice", details.Name)

	// super admin lists users: now includes self with is_me flag
	req2 := httptest.NewRequest(http.MethodGet, "/user", nil)
	req2.Header.Set("X-Subject", "alice")
	req2.Header.Set("X-Role-Super-Admin", "true")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusOK, rec2.Code)
	var list ListUserResponse
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &list))
	// should contain alice, bob and root
	names := []string{}
	var aliceFound bool
	for _, u := range list.Users {
		names = append(names, u.Name)
		if u.Name == "alice" && u.IsMe != nil && *u.IsMe {
			aliceFound = true
		}
	}
	require.ElementsMatch(t, []string{"alice", "bob", "root"}, names)
	require.True(t, aliceFound, "alice should have is_me=true")
}

func Test_GetUser_Authorization(t *testing.T) {
	now := time.Now()
	s := &fakeStore{users: map[string]user.User{"bob": &fakeUser{id: "bob", created: now, extra: map[string]any{
		"roles": map[string]any{},
		"info":  map[string]any{},
	}}}, retryLimit: 3, blockDuration: 10 * time.Minute}
	r := newTestRouter(s)

	// Without manage or super-admin role, should return limited details (name only)
	req := httptest.NewRequest(http.MethodGet, "/user/bob", nil)
	req.Header.Set("X-Subject", "alice")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
	var d Details
	require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &d))
	require.Equal(t, "bob", d.Name)
	require.Nil(t, d.CreatedAt)

	// With manage role, should get full details
	req2 := httptest.NewRequest(http.MethodGet, "/user/bob", nil)
	req2.Header.Set("X-Subject", "admin")
	req2.Header.Set("X-Role-Manage-User", "true")
	rec2 := httptest.NewRecorder()
	r.ServeHTTP(rec2, req2)
	require.Equal(t, http.StatusOK, rec2.Code)
	var d2 Details
	require.NoError(t, json.Unmarshal(rec2.Body.Bytes(), &d2))
	require.NotNil(t, d2.CreatedAt)
}
