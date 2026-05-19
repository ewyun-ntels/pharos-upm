package userhandler

import (
	"encoding/json"
	"errors"
	"log/slog"
	"math"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite"
	"github.com/ory/fosite/token/jwt"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/internal/user"
	"ntels.com/pharos/core/pkg/common"
	coreRole "ntels.com/pharos/core/pkg/role"
	coreUser "ntels.com/pharos/core/pkg/user"
	"ntels.com/pharos/shared/types/role"
)

type Handler interface {
	GetUserList(c *gin.Context)
	GetMe(c *gin.Context)
	GetUser(c *gin.Context)
	CreateUser(c *gin.Context)
	DeleteUser(c *gin.Context)
	Block(c *gin.Context)
	ChangePassword(c *gin.Context)
	ChangePasswordFromAdmin(c *gin.Context)
	UpdatePasswordExpiration(c *gin.Context)
	SetAttributes(c *gin.Context)
	// User Metadata Config API
	GetUserMetadataConfig(c *gin.Context)
	GetUserMetadataConfigHistory(c *gin.Context)
	SaveUserMetadataConfig(c *gin.Context)
	RollbackUserMetadataConfig(c *gin.Context)

	checkHandler(c *gin.Context) bool
}

// NewHandler creates a new user handler with the given store.
func NewHandler(store user.Store, roleMetadataService *coreRole.MetadataService, userMetadataService *coreUser.MetadataService) Handler {
	h := &handler{
		store:               store,
		roleMetadataService: roleMetadataService,
		userMetadataService: userMetadataService,
	}

	// 범위를 0 ~ math.MaxInt32 로 제한하더라도 gosec에서 검출되기 때문에
	// int32 범위 변환 먼저 시도 후 음수 및 오버플로우 체크
	// G115 (CWE-190): integer overflow conversion int -> int32
	retention := config.User.PasswordHistoryRetention
	if retention >= math.MinInt32 && retention <= math.MaxInt32 {
		h.passwordHistoryRetention = int32(retention)
	}

	// 음수로 설정 될 경우 일부 디비에서 에러로 처리되므로 0으로 강제 설정
	// 0으로 설정 될 경우 패스워드 히스토리 기능을 사용하지 않음
	// 0보다 큰 값으로 설정 될 경우 해당 개수만큼 패스워드 히스토리를 보관함
	if retention < 0 {
		h.passwordHistoryRetention = 0
	} else if retention > math.MaxInt32 {
		h.passwordHistoryRetention = math.MaxInt32
	}

	return h
}

type handler struct {
	store               user.Store
	roleMetadataService *coreRole.MetadataService
	userMetadataService *coreUser.MetadataService

	passwordHistoryRetention int32
}

func (h *handler) checkHandler(c *gin.Context) bool {
	if h == nil || h.store == nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "user handler not initialized"})
		return false
	}
	return true
}

func (h *handler) UpdatePasswordExpiration(c *gin.Context) {
	username := c.Param(common.ParamUsername)
	if username == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	var jwtClaims *jwt.JWTClaims
	if value, exists := c.Get(common.ContextKeyJWTClaims); exists {
		jwtClaims = value.(*jwt.JWTClaims)
	} else {
		jwtClaims = &jwt.JWTClaims{}
	}
	if common.GetRole(jwtClaims, string(role.RoleUserUpdate)) {
		statusCode, err := h.checkSuperUser(c, username, true)
		if err != nil {
			c.AbortWithStatusJSON(statusCode, gin.H{"error": err.Error()})
			return
		}
	}

	var req UpdatePasswordExpirationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	if err := h.store.UpdatePasswordExpiration(username, req.ExpiredAt); err != nil {
		if errors.Is(err, fosite.ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
}

// checkSuperUser는 대상 사용자(username)를 관리할 권한이 있는지 확인합니다.
// 일반 사용자: manage_user 또는 super_admin 권한 필요
// super_admin 사용자: super_admin 권한만 가능 (관리자는 불가)
// allowSelfOperation이 false면 자기 자신은 관리 불가
func (h *handler) checkSuperUser(c *gin.Context, username string, allowSelfOperation bool) (statusCode int, err error) {
	var claims *jwt.JWTClaims
	if value, exists := c.Get(common.ContextKeyJWTClaims); exists {
		claims = value.(*jwt.JWTClaims)
	} else {
		claims = &jwt.JWTClaims{}
	}

	// 자기 자신 관리 가능 여부 확인
	if !allowSelfOperation && claims.Subject == username {
		return http.StatusForbidden, errors.New("cannot manage self")
	}

	// 대상 사용자 조회
	userInfo, err := h.store.FindByUsername(username)
	if errors.Is(err, fosite.ErrNotFound) {
		return http.StatusNotFound, errors.New("user not found")
	} else if err != nil {
		return http.StatusInternalServerError, errors.New("failed to find user")
	}

	// 대상이 super_admin인 경우, 요청자도 super_admin이어야 함
	// (manage_user 권한만으로는 super_admin을 관리할 수 없음)
	targetIsSuperAdmin := common.HasRole(userInfo.GetExtra(), string(role.RoleSuperAdmin))
	if targetIsSuperAdmin && !common.GetRole(claims, string(role.RoleSuperAdmin)) {
		return http.StatusForbidden, errors.New("only super admin can manage super admin users")
	}

	return http.StatusOK, nil
}

func (h *handler) ChangePassword(c *gin.Context) {
	var claims *jwt.JWTClaims
	if value, exists := c.Get(common.ContextKeyJWTClaims); exists {
		claims = value.(*jwt.JWTClaims)
	} else {
		claims = &jwt.JWTClaims{}
	}

	if common.GetRole(claims, string(role.AttrTemporaryUser)) &&
		!common.GetRole(claims, string(role.AttrPasswordChange)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "temporary user cannot change password"})
		return
	}

	var req ChangeMyPasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// Verify current password
	if err := h.store.Authenticate(claims.Subject, req.CurrentPassword); err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid current password"})
		return
	}

	h.internalPasswordChange(c, claims.Subject, req.NewPassword)
}

func (h *handler) ChangePasswordFromAdmin(c *gin.Context) {
	if !h.checkHandler(c) {
		return
	}

	username := c.Param(common.ParamUsername)
	if username == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	statusCode, err := h.checkSuperUser(c, username, false)
	if err != nil {
		c.AbortWithStatusJSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	h.internalPasswordChange(c, username, req.Password)
}

func (h *handler) internalPasswordChange(c *gin.Context, username string, newPassword string) {
	info, err := h.store.FindByUsername(username)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to find user"})
		return
	}

	err = h.store.ValidatePasswordComplexity(newPassword)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	err = h.store.CheckPasswordHistory(username, newPassword)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err = h.store.ChangePassword(username, newPassword); err != nil {
		if errors.Is(err, fosite.ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}

		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	err = h.store.InsertPasswordHistory(username, newPassword)
	if err != nil {
		slog.Error("failed to insert password history", "username", username, "error", err)
	}

	err = h.store.DeletePasswordHistory(username, h.passwordHistoryRetention)
	if err != nil {
		slog.Error("failed to delete password history", "username", username, "error", err)
		return
	}

	prepare := info.GetPrepare()
	if len(prepare) > 0 {
		if prepare[0] == user.PreparePasswordChange {
			prepare = prepare[1:]
		}

		err = h.store.UpdatePrepare(username, prepare)
		if err != nil {
			slog.Error("failed to update prepare", "username", username, "error", err)
			return
		}
	}

	c.Status(http.StatusOK)
}

func (h *handler) Block(c *gin.Context) {
	if !h.checkHandler(c) {
		return
	}

	username := c.Param(common.ParamUsername)
	if username == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	isBlocked := false
	isBlockedStr := c.Query("block")
	if isBlockedStr != "" {
		isBlocked = true
	}

	statusCode, err := h.checkSuperUser(c, username, false)
	if err != nil {
		c.AbortWithStatusJSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	if err := h.store.SetBlock(username, isBlocked); err != nil {
		if errors.Is(err, fosite.ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.Status(http.StatusOK)
}

func (h *handler) GetUserList(c *gin.Context) {
	if !h.checkHandler(c) {
		return
	}

	users, err := h.store.ListUsers()
	if err != nil {
		c.AbortWithStatusJSON(500, external.ErrorResponse{Message: err.Error()})
		return
	}

	var jwtClaims *jwt.JWTClaims
	if value, exists := c.Get(common.ContextKeyJWTClaims); exists {
		jwtClaims = value.(*jwt.JWTClaims)
	} else {
		jwtClaims = &jwt.JWTClaims{}
	}

	res := ListUserResponse{}

	for _, u := range users {
		details := h.getEffectiveUserDetail(u)

		// 자기 자신인지 플래그 설정
		isMe := u.GetID() == jwtClaims.Subject
		details.IsMe = &isMe

		if common.GetRole(jwtClaims, string(role.RoleSuperAdmin)) {
			res.Users = append(res.Users, details)
		} else if common.GetRole(jwtClaims, string(role.RoleUserRead)) {
			// 관리자는 super admin이 아닌 사용자만 조회 가능
			if !common.HasRole(u.GetExtra(), string(role.RoleSuperAdmin)) {
				res.Users = append(res.Users, details)
			}
		}
	}

	c.JSON(http.StatusOK, res)
}

func (h *handler) GetMe(c *gin.Context) {
	if !h.checkHandler(c) {
		return
	}

	var claims *jwt.JWTClaims
	if value, exists := c.Get(common.ContextKeyJWTClaims); exists {
		claims = value.(*jwt.JWTClaims)
	} else {
		claims = &jwt.JWTClaims{}
	}

	username := claims.Subject
	if username == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	u, err := h.store.FindByUsername(username)
	if err != nil {
		if errors.Is(err, fosite.ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to find user"})
		return
	}

	res := h.getEffectiveUserDetail(u)

	c.JSON(http.StatusOK, res)
}

func (h *handler) GetUser(c *gin.Context) {
	if !h.checkHandler(c) {
		return
	}

	username := c.Param(common.ParamUsername)
	if username == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	u, err := h.store.FindByUsername(username)
	if err != nil {
		if errors.Is(err, fosite.ErrNotFound) {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{"error": "user not found"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to find user"})
		return
	}

	var jwtClaims *jwt.JWTClaims
	if value, exists := c.Get(common.ContextKeyJWTClaims); exists {
		jwtClaims = value.(*jwt.JWTClaims)
	} else {
		jwtClaims = &jwt.JWTClaims{}
	}

	statusCode, err := h.checkSuperUser(c, username, false)
	if err != nil {
		c.AbortWithStatusJSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	var res Details
	if common.GetRole(jwtClaims, string(role.RoleSuperAdmin)) ||
		common.GetRole(jwtClaims, string(role.RoleUserRead)) {
		res = h.getEffectiveUserDetail(u)
	} else {
		res = Details{
			Name: u.GetID(),
		}
	}

	c.JSON(http.StatusOK, res)
}

func (h *handler) getEffectiveUserDetail(u user.User) Details {
	details := getUserDetail(u, h.store.GetLoginRetryLimit(), h.store.GetLoginBlockDuration(), h.roleMetadataService)
	if prepare, err := h.store.GetPrepare(u); err == nil {
		details.Prepare = prepare
		labels := make([]string, 0, len(prepare))
		for _, key := range prepare {
			if def, ok := user.GetActionRoleDefinition(key); ok {
				labels = append(labels, def.DisplayLabel(u))
			} else {
				labels = append(labels, key)
			}
		}
		details.PrepareLabels = labels
	} else {
		slog.Warn("failed to resolve effective prepare", "username", u.GetID(), "error", err)
	}
	return details
}

func (h *handler) CreateUser(c *gin.Context) {
	if !h.checkHandler(c) {
		return
	}

	var req CreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			external.ErrorResponse{Message: "invalid request body"})
		return
	}

	// Validate using shared type validation
	if err := ValidateCreateUserRequest(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			external.ErrorResponse{Message: err.Error()})
		return
	}

	var claims *jwt.JWTClaims
	if value, exists := c.Get(common.ContextKeyJWTClaims); exists {
		claims = value.(*jwt.JWTClaims)
	} else {
		claims = &jwt.JWTClaims{}
	}

	if !common.GetRole(claims, string(role.RoleSuperAdmin)) {
		if req.Attributes != nil && req.Attributes.Roles != nil {
			if v, ok := req.Attributes.Roles[string(role.RoleSuperAdmin)]; ok && v {
				c.AbortWithStatusJSON(http.StatusForbidden,
					external.ErrorResponse{Message: "only super admin can create super admin"})
				return
			}
		}
	}

	if req.Username == "" || req.Password == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest,
			external.ErrorResponse{Message: "username and password are required"})
		return
	}

	err := h.store.ValidatePasswordComplexity(req.Password)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, external.ErrorResponse{Message: err.Error()})
		return
	}

	// Convert Attributes to map[string]interface{} using JSON
	var attrs map[string]any
	if req.Attributes != nil {
		data, err := json.Marshal(req.Attributes)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, external.ErrorResponse{Message: "failed to process attributes"})
			return
		}
		if err := json.Unmarshal(data, &attrs); err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, external.ErrorResponse{Message: "failed to process attributes"})
			return
		}
	}

	if err := h.store.CreateUser(req.Username, req.Password, attrs, &config.User.Prepare); err != nil {
		if errors.Is(err, user.ErrUserAlreadyExists) {
			c.AbortWithStatusJSON(http.StatusBadRequest, external.ErrorResponse{Message: "user already exists"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
		return
	}

	err = h.store.InsertPasswordHistory(req.Username, req.Password)
	if err != nil {
		slog.Error("failed to insert password history", "username", req.Username, "error", err)
	}

	c.Status(http.StatusCreated)
}

func (h *handler) DeleteUser(c *gin.Context) {
	if !h.checkHandler(c) {
		return
	}

	username := c.Param(common.ParamUsername)
	if username == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	statusCode, err := h.checkSuperUser(c, username, false)
	if err != nil {
		c.AbortWithStatusJSON(statusCode, gin.H{"error": err.Error()})
		return
	}

	if err := h.store.DeleteUser(username); err != nil {
		if errors.Is(err, fosite.ErrNotFound) {
			c.AbortWithStatusJSON(404, gin.H{"error": "user not found"})
			return
		}
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to delete user"})
		return
	}

	err = h.store.DeletePasswordHistory(username, 0)
	if err != nil {
		slog.Error("failed to delete password history", "username", username, "error", err)
	}

	c.Status(http.StatusOK)
}

func (h *handler) SetAttributes(c *gin.Context) {
	if !h.checkHandler(c) {
		return
	}

	username := c.Param(common.ParamUsername)
	if username == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "username is required"})
		return
	}

	var claims *jwt.JWTClaims
	if value, exists := c.Get(common.ContextKeyJWTClaims); exists {
		claims = value.(*jwt.JWTClaims)
	} else {
		claims = &jwt.JWTClaims{}
	}

	if claims.Subject == username {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "cannot manage self"})
		return
	}

	var req SetAttributesRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid request body"})
		return
	}

	// check attribute
	if !common.GetRole(claims, string(role.RoleUserUpdate)) &&
		!common.GetRole(claims, string(role.RoleSuperAdmin)) {
		c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "only user update permission can set attributes"})
		return
	}

	if req.Attributes.Roles != nil {
		if _, exists := req.Attributes.Roles[string(role.RoleSuperAdmin)]; exists &&
			!common.GetRole(claims, string(role.RoleSuperAdmin)) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "only super admin can set super admin"})
			return
		}
	}

	// Convert Attributes to map[string]interface{} using JSON
	data, err := json.Marshal(&req.Attributes)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to process attributes"})
		return
	}

	var attrs map[string]any
	if err := json.Unmarshal(data, &attrs); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to process attributes"})
		return
	}

	if err := h.store.SetAttributes(username, attrs); err != nil {
		slog.Error("failed to set attributes", "username", username, "error", err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to set user attributes"})
		return
	}

	c.Status(http.StatusNoContent)
}
