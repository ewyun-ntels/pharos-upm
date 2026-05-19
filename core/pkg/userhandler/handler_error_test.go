package userhandler

import (
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite"
	"github.com/ory/fosite/token/jwt"
	"github.com/stretchr/testify/require"
	"ntels.com/pharos/core/internal/user"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/shared/types/role"
)

// newHandlerForTest creates a handler for testing with nil metadataService
func newHandlerForTest(store user.Store) *handler {
	return NewHandler(store, nil, nil).(*handler)
}

type errableStore struct {
	fakeStore
	listErr         error
	findErr         error
	createErr       error
	deleteErr       error
	setBlockErr     error
	setAttrErr      error
	validatePwdErr  error
	checkHistErr    error
	changePwdErr    error
	updatePwdExpErr error
}

func (s *errableStore) ListUsers() ([]user.User, error) {
	if s.listErr != nil {
		return nil, s.listErr
	}
	return s.fakeStore.ListUsers()
}
func (s *errableStore) FindByUsername(username string) (user.User, error) {
	if s.findErr != nil {
		return nil, s.findErr
	}
	return s.fakeStore.FindByUsername(username)
}
func (s *errableStore) CreateUser(u, p string, a map[string]any, prep *[]string) error {
	if s.createErr != nil {
		return s.createErr
	}
	return s.fakeStore.CreateUser(u, p, a, prep)
}
func (s *errableStore) DeleteUser(u string) error {
	if s.deleteErr != nil {
		return s.deleteErr
	}
	return s.fakeStore.DeleteUser(u)
}
func (s *errableStore) SetBlock(u string, b bool) error {
	if s.setBlockErr != nil {
		return s.setBlockErr
	}
	return s.fakeStore.SetBlock(u, b)
}
func (s *errableStore) SetAttributes(u string, a map[string]any) error {
	if s.setAttrErr != nil {
		return s.setAttrErr
	}
	return s.fakeStore.SetAttributes(u, a)
}
func (s *errableStore) ValidatePasswordComplexity(p string) error {
	if s.validatePwdErr != nil {
		return s.validatePwdErr
	}
	return s.fakeStore.ValidatePasswordComplexity(p)
}
func (s *errableStore) CheckPasswordHistory(u, p string) error {
	if s.checkHistErr != nil {
		return s.checkHistErr
	}
	return s.fakeStore.CheckPasswordHistory(u, p)
}
func (s *errableStore) ChangePassword(u, p string) error {
	if s.changePwdErr != nil {
		return s.changePwdErr
	}
	return s.fakeStore.ChangePassword(u, p)
}
func (s *errableStore) UpdatePasswordExpiration(u string, t *time.Time) error {
	if s.updatePwdExpErr != nil {
		return s.updatePwdExpErr
	}
	return s.fakeStore.UpdatePasswordExpiration(u, t)
}

func newRouterWithClaims(h *handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
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
		if c.GetHeader("X-Temporary-User") == "true" {
			extra.Roles[string(role.AttrTemporaryUser)] = true
		}
		if c.GetHeader("X-Prepare-Password-Change") == "true" {
			extra.Info[user.PreparePasswordChange] = true
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
	// bind routes
	r.GET("/me", func(c *gin.Context) { h.GetMe(c) })
	r.GET("/user", func(c *gin.Context) { h.GetUserList(c) })
	r.GET("/user/:username", func(c *gin.Context) { h.GetUser(c) })
	r.DELETE("/user/:username", func(c *gin.Context) { h.DeleteUser(c) })
	r.PUT("/user/:username/block", func(c *gin.Context) { h.Block(c) })
	r.PUT("/user/:username/password", func(c *gin.Context) { h.ChangePasswordFromAdmin(c) })
	r.PUT("/user/:username/password-expired-at", func(c *gin.Context) { h.UpdatePasswordExpiration(c) })
	r.PUT("/user/:username/attributes", func(c *gin.Context) { h.SetAttributes(c) })
	r.POST("/user", func(c *gin.Context) { h.CreateUser(c) })
	r.PUT("/me/password", func(c *gin.Context) { h.ChangePassword(c) })
	return r
}

func Test_GetUserList_and_InitErrors(t *testing.T) {
	now := time.Now()
	s := &errableStore{fakeStore: fakeStore{users: map[string]user.User{"me": &fakeUser{id: "me", created: now, extra: map[string]any{"roles": map[string]any{}, "info": map[string]any{}}}}}, listErr: errors.New("boom")}
	h := newHandlerForTest(s)
	r := newRouterWithClaims(h)
	// list users error should return 500
	req := httptest.NewRequest(http.MethodGet, "/user", nil)
	req.Header.Set("X-Role-Super-Admin", "true")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	// GetMe: missing subject -> 400
	req = httptest.NewRequest(http.MethodGet, "/me", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
}

func Test_CheckSuperUser_Permissions(t *testing.T) {
	now := time.Now()
	// DB 구조와 동일: {"roles": {...}, "info": {...}}
	rootExtra := map[string]any{
		"roles": map[string]any{string(role.RoleSuperAdmin): true},
		"info":  map[string]any{},
	}
	root := &fakeUser{id: "root", created: now, extra: rootExtra}
	s := &errableStore{fakeStore: fakeStore{users: map[string]user.User{"root": root}}}
	h := newHandlerForTest(s)
	r := newRouterWithClaims(h)
	// manager but not super-admin tries to change password of super-admin -> 403
	req := httptest.NewRequest(http.MethodPut, "/user/root/password", bytes.NewBufferString(`{"password":"p"}`))
	req.Header.Set("X-Role-Manage-User", "true")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)
}

func Test_GetUser_ErrorFlows(t *testing.T) {
	now := time.Now()
	base := &fakeStore{users: map[string]user.User{"bob": &fakeUser{id: "bob", created: now, extra: map[string]any{"roles": map[string]any{}, "info": map[string]any{}}}}, retryLimit: 3, blockDuration: 5 * time.Minute}
	s := &errableStore{fakeStore: *base}
	h := newHandlerForTest(s)
	r := newRouterWithClaims(h)

	// missing username -> 400 (use explicit empty param path we didn't register; Gin returns 404 or 301 for slash variants, so skip this check)

	// not found -> 404 mapped when store returns fosite.ErrNotFound
	s.findErr = fosite.ErrNotFound
	req := httptest.NewRequest(http.MethodGet, "/user/alice", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
	s.findErr = nil

	// find error -> 500
	s.findErr = errors.New("db down")
	req = httptest.NewRequest(http.MethodGet, "/user/bob", nil)
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func Test_SetAttributes_Flows(t *testing.T) {
	now := time.Now()
	s := &errableStore{fakeStore: fakeStore{users: map[string]user.User{"bob": &fakeUser{id: "bob", created: now, extra: map[string]any{"roles": map[string]any{}, "info": map[string]any{}}}}}}
	h := newHandlerForTest(s)
	r := newRouterWithClaims(h)

	// missing username
	req := httptest.NewRequest(http.MethodPut, "/user//attributes", bytes.NewBufferString(`{}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// cannot manage self
	body, _ := json.Marshal(map[string]any{
		"attributes": map[string]any{
			"roles": map[string]any{"a": true},
		},
	})
	req = httptest.NewRequest(http.MethodPut, "/user/alice/attributes", bytes.NewReader(body))
	req.Header.Set("X-Subject", "alice")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	// insufficient role
	req = httptest.NewRequest(http.MethodPut, "/user/bob/attributes", bytes.NewReader(body))
	req.Header.Set("X-Subject", "alice")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	// trying to set super_admin without being super_admin
	body, _ = json.Marshal(map[string]any{
		"attributes": map[string]any{
			"roles": map[string]any{string(role.RoleSuperAdmin): true},
		},
	})
	req = httptest.NewRequest(http.MethodPut, "/user/bob/attributes", bytes.NewReader(body))
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	// bad body
	req = httptest.NewRequest(http.MethodPut, "/user/bob/attributes", bytes.NewBufferString(`invalid`))
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// attrs nil (null attributes is allowed - creates user without attributes)
	req = httptest.NewRequest(http.MethodPut, "/user/bob/attributes", bytes.NewBufferString(`{"attributes":null}`))
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)

	// store error
	body, _ = json.Marshal(map[string]any{
		"attributes": map[string]any{
			"roles": map[string]any{"a": true},
		},
	})
	s.setAttrErr = errors.New("fail")
	req = httptest.NewRequest(http.MethodPut, "/user/bob/attributes", bytes.NewReader(body))
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	// success
	s.setAttrErr = nil
	req = httptest.NewRequest(http.MethodPut, "/user/bob/attributes", bytes.NewReader(body))
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code)
}

func Test_CreateUser_Flows(t *testing.T) {
	s := &errableStore{fakeStore: fakeStore{users: map[string]user.User{}}}
	h := newHandlerForTest(s)
	r := newRouterWithClaims(h)

	// invalid body
	req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewBufferString("{"))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// create super admin without role - new nested structure
	body, _ := json.Marshal(map[string]any{
		"username": "a",
		"password": "p",
		"attributes": map[string]any{
			"roles": map[string]any{string(role.RoleSuperAdmin): true},
		},
	})
	req = httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(body))
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	// missing fields
	body, _ = json.Marshal(map[string]any{
		"username": "",
		"password": "",
	})
	req = httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(body))
	req.Header.Set("X-Role-Super-Admin", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// password complexity fail
	s.validatePwdErr = errors.New("weak")
	body, _ = json.Marshal(map[string]any{
		"username": "a",
		"password": "p",
	})
	req = httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(body))
	req.Header.Set("X-Role-Super-Admin", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// already exists
	s.validatePwdErr = nil
	s.createErr = user.ErrUserAlreadyExists
	body, _ = json.Marshal(map[string]any{
		"username": "a",
		"password": "p",
	})
	req = httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(body))
	req.Header.Set("X-Role-Super-Admin", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// other error
	s.createErr = errors.New("db")
	body, _ = json.Marshal(map[string]any{
		"username": "a",
		"password": "p",
	})
	req = httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(body))
	req.Header.Set("X-Role-Super-Admin", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	// success
	s.createErr = nil
	body, _ = json.Marshal(map[string]any{
		"username": "a",
		"password": "p",
	})
	req = httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(body))
	req.Header.Set("X-Role-Super-Admin", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code)
}

func Test_Delete_and_Block_Flows(t *testing.T) {
	now := time.Now()
	s := &errableStore{fakeStore: fakeStore{users: map[string]user.User{"alice": &fakeUser{id: "alice", created: now, extra: map[string]any{"roles": map[string]any{}, "info": map[string]any{}}}}}}
	h := newHandlerForTest(s)
	r := newRouterWithClaims(h)

	// delete: missing username -> route not registered; skip explicit 400 check (Gin 404)

	// delete: cannot manage self
	req := httptest.NewRequest(http.MethodDelete, "/user/alice", nil)
	req.Header.Set("X-Role-Manage-User", "true")
	req.Header.Set("X-Subject", "alice")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	// delete: not found after checkSuperUser passes; seed bob and return not found from DeleteUser
	req = httptest.NewRequest(http.MethodDelete, "/user/alice", nil)
	req.Header.Set("X-Role-Manage-User", "true")
	s.deleteErr = fosite.ErrNotFound
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
	s.deleteErr = nil

	// delete: store error -> 500
	s.deleteErr = errors.New("db")
	req = httptest.NewRequest(http.MethodDelete, "/user/alice", nil)
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	s.deleteErr = nil

	// delete: success -> 200
	req = httptest.NewRequest(http.MethodDelete, "/user/alice", nil)
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	// block: missing username
	req = httptest.NewRequest(http.MethodPut, "/user//block", nil)
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// block: store error
	req = httptest.NewRequest(http.MethodPut, "/user/alice/block?block=1", nil)
	req.Header.Set("X-Role-Manage-User", "true")
	s.setBlockErr = errors.New("fail")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)

	// block: success
	s.setBlockErr = nil
	req = httptest.NewRequest(http.MethodPut, "/user/alice/block?block=1", nil)
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func Test_UpdatePasswordExpiration_and_ChangePassword(t *testing.T) {
	now := time.Now()
	s := &errableStore{fakeStore: fakeStore{users: map[string]user.User{"alice": &fakeUser{id: "alice", created: now, extra: map[string]any{"roles": map[string]any{}, "info": map[string]any{}}}}}}
	h := newHandlerForTest(s)
	r := newRouterWithClaims(h)

	// UpdatePasswordExpiration: missing username
	req := httptest.NewRequest(http.MethodPut, "/user//password-expired-at", bytes.NewBufferString(`{"expired_at":null}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// bad body
	req = httptest.NewRequest(http.MethodPut, "/user/alice/password-expired-at", bytes.NewBufferString(`{`))
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// not found path
	s.findErr = fosite.ErrNotFound
	req = httptest.NewRequest(http.MethodPut, "/user/alice/password-expired-at", bytes.NewBufferString(`{"expired_at":null}`))
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	// checkSuperUser calls FindByUsername; ensure it passes with no forbid, then UpdatePasswordExpiration error mapping
	require.Equal(t, http.StatusNotFound, rec.Code)
	s.findErr = nil

	// store error
	s.updatePwdExpErr = errors.New("db")
	req = httptest.NewRequest(http.MethodPut, "/user/alice/password-expired-at", bytes.NewBufferString(`{"expired_at":null}`))
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	// success
	s.updatePwdExpErr = nil
	req = httptest.NewRequest(http.MethodPut, "/user/alice/password-expired-at", bytes.NewBufferString(`{"expired_at":null}`))
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)

	// ChangePassword: forbid temporary user without prepare
	req = httptest.NewRequest(http.MethodPut, "/me/password", bytes.NewBufferString(`{"password":"p"}`))
	req.Header.Set("X-Subject", "alice")
	req.Header.Set("X-Temporary-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusForbidden, rec.Code)

	// ChangePasswordFromAdmin: missing username -> skip because route not registered; focus on success path
	// ChangePasswordFromAdmin: success
	req = httptest.NewRequest(http.MethodPut, "/user/alice/password", bytes.NewBufferString(`{"password":"p"}`))
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func Test_GetMe_Error_Paths(t *testing.T) {
	now := time.Now()
	s := &errableStore{fakeStore: fakeStore{users: map[string]user.User{"alice": &fakeUser{id: "alice", created: now, extra: map[string]any{"roles": map[string]any{}, "info": map[string]any{}}}}}}
	h := newHandlerForTest(s)
	r := newRouterWithClaims(h)
	// not found
	s.findErr = fosite.ErrNotFound
	req := httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("X-Subject", "alice")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
	// internal error
	s.findErr = errors.New("db")
	req = httptest.NewRequest(http.MethodGet, "/me", nil)
	req.Header.Set("X-Subject", "alice")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
}

func Test_passwordChange_ErrorPaths(t *testing.T) {
	now := time.Now()
	alice := &fakeUser{id: "alice", created: now, extra: map[string]any{"roles": map[string]any{}, "info": map[string]any{}}}
	s := &errableStore{fakeStore: fakeStore{users: map[string]user.User{"alice": alice}}}
	h := newHandlerForTest(s)
	r := newRouterWithClaims(h)

	// invalid body
	req := httptest.NewRequest(http.MethodPut, "/user/alice/password", bytes.NewBufferString("{"))
	req.Header.Set("X-Role-Manage-User", "true")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)

	// find error
	req = httptest.NewRequest(http.MethodPut, "/user/alice/password", bytes.NewBufferString(`{"password":"p"}`))
	req.Header.Set("X-Role-Manage-User", "true")
	s.findErr = errors.New("db")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusInternalServerError, rec.Code)
	s.findErr = nil

	// complexity fail
	s.validatePwdErr = errors.New("weak")
	req = httptest.NewRequest(http.MethodPut, "/user/alice/password", bytes.NewBufferString(`{"password":"p"}`))
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	s.validatePwdErr = nil

	// history check fail
	s.checkHistErr = errors.New("hist")
	req = httptest.NewRequest(http.MethodPut, "/user/alice/password", bytes.NewBufferString(`{"password":"p"}`))
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code)
	s.checkHistErr = nil

	// change password not found -> 404 mapping
	s.changePwdErr = fosite.ErrNotFound
	req = httptest.NewRequest(http.MethodPut, "/user/alice/password", bytes.NewBufferString(`{"password":"p"}`))
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNotFound, rec.Code)
	s.changePwdErr = nil

	// success
	req = httptest.NewRequest(http.MethodPut, "/user/alice/password", bytes.NewBufferString(`{"password":"p"}`))
	req.Header.Set("X-Role-Manage-User", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusOK, rec.Code)
}

func Test_SetAttributes_TypeValidation(t *testing.T) {
	now := time.Now()
	s := &errableStore{fakeStore: fakeStore{users: map[string]user.User{"bob": &fakeUser{id: "bob", created: now, extra: map[string]any{"roles": map[string]any{}, "info": map[string]any{}}}}}}
	h := newHandlerForTest(s)
	r := newRouterWithClaims(h)

	// Test: non-boolean value (string) should fail - new nested structure
	body, _ := json.Marshal(map[string]any{
		"attributes": map[string]any{
			"roles": map[string]any{"manage_user_role": "true"}, // string instead of bool
		},
	})
	req := httptest.NewRequest(http.MethodPut, "/user/bob/attributes", bytes.NewReader(body))
	req.Header.Set("X-Role-Manage-User", "true")
	req.Header.Set("X-Subject", "alice")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code, "string value should be rejected")

	// Test: non-boolean value (number) should fail - new nested structure
	body, _ = json.Marshal(map[string]any{
		"attributes": map[string]any{
			"roles": map[string]any{"manage_user_role": 1}, // number instead of bool
		},
	})
	req = httptest.NewRequest(http.MethodPut, "/user/bob/attributes", bytes.NewReader(body))
	req.Header.Set("X-Role-Manage-User", "true")
	req.Header.Set("X-Subject", "alice")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code, "number value should be rejected")

	// Test: non-boolean value (object) should fail - new nested structure
	body, _ = json.Marshal(map[string]any{
		"attributes": map[string]any{
			"roles": map[string]any{"manage_user_role": map[string]any{"enabled": true}}, // object instead of bool
		},
	})
	req = httptest.NewRequest(http.MethodPut, "/user/bob/attributes", bytes.NewReader(body))
	req.Header.Set("X-Role-Manage-User", "true")
	req.Header.Set("X-Subject", "alice")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code, "object value should be rejected")

	// Test: boolean value should succeed - new nested structure
	body, _ = json.Marshal(map[string]any{
		"attributes": map[string]any{
			"roles": map[string]any{"manage_user_role": true}, // correct boolean
		},
	})
	req = httptest.NewRequest(http.MethodPut, "/user/bob/attributes", bytes.NewReader(body))
	req.Header.Set("X-Role-Manage-User", "true")
	req.Header.Set("X-Subject", "alice")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusNoContent, rec.Code, "boolean value should be accepted")
}

func Test_CreateUser_TypeValidation(t *testing.T) {
	s := &errableStore{fakeStore: fakeStore{users: make(map[string]user.User)}}
	h := newHandlerForTest(s)
	r := newRouterWithClaims(h)

	// Test: non-boolean attribute value (string) should fail - new nested structure
	body, _ := json.Marshal(map[string]any{
		"username": "testuser",
		"password": "TestPass123!",
		"attributes": map[string]any{
			"roles": map[string]any{"manage_user_role": "true"}, // string instead of bool
		},
	})
	req := httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(body))
	req.Header.Set("X-Role-Super-Admin", "true")
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code, "string attribute value should be rejected")

	// Test: non-boolean attribute value (number) should fail - new nested structure
	body, _ = json.Marshal(map[string]any{
		"username": "testuser",
		"password": "TestPass123!",
		"attributes": map[string]any{
			"roles": map[string]any{"manage_user_role": 1}, // number instead of bool
		},
	})
	req = httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(body))
	req.Header.Set("X-Role-Super-Admin", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusBadRequest, rec.Code, "number attribute value should be rejected")

	// Test: boolean attribute value should succeed - new nested structure
	body, _ = json.Marshal(map[string]any{
		"username": "testuser",
		"password": "TestPass123!",
		"attributes": map[string]any{
			"roles": map[string]any{"manage_user_role": true}, // correct boolean
		},
	})
	req = httptest.NewRequest(http.MethodPost, "/user", bytes.NewReader(body))
	req.Header.Set("X-Role-Super-Admin", "true")
	rec = httptest.NewRecorder()
	r.ServeHTTP(rec, req)
	require.Equal(t, http.StatusCreated, rec.Code, "boolean attribute value should be accepted")
}
