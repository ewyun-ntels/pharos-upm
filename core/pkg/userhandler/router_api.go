package userhandler

import (
	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/internal/user"
	"ntels.com/pharos/core/pkg/authhandler"
	"ntels.com/pharos/core/pkg/common"
	coreRole "ntels.com/pharos/core/pkg/role"
	coreUser "ntels.com/pharos/core/pkg/user"
	"ntels.com/pharos/shared/types/role"
)

var config common.Config

// RegisterRoutes registers user-related routes (/users and /me) into the given router groups.
// It relies on external middlewares provided by the auth package via function arguments to avoid package dependency cycles.
func RegisterRoutes(router gin.IRouter, store user.Store, roleMetadataService *coreRole.MetadataService, userMetadataService *coreUser.MetadataService) {
	userHandler := NewHandler(store, roleMetadataService, userMetadataService)

	// /me
	me := router.Group("/me")
	{
		me.GET("", authhandler.GetDecodeAuthenticationHandler(),
			userHandler.GetMe)
		me.PUT("/password", authhandler.GetDecodeAuthenticationHandler(),
			userHandler.ChangePassword, authhandler.ClearOwnToken)
	}

	// /user
	userGroup := router.Group("/user")
	{
		userGroup.GET("",
			authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleUserRead)),
			userHandler.GetUserList)
		userGroup.POST("",
			authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleUserCreate)),
			userHandler.CreateUser)
		userGroup.GET("/:username",
			authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleUserRead)),
			userHandler.GetUser)
		userGroup.DELETE("/:username",
			authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleUserDelete)),
			userHandler.DeleteUser, authhandler.ClearParamToken)
		userGroup.PUT("/:username/block",
			authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleUserUpdate)),
			userHandler.Block, authhandler.ClearParamToken)
		userGroup.PUT("/:username/password",
			authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleUserUpdate)),
			userHandler.ChangePasswordFromAdmin, authhandler.ClearParamToken)
		userGroup.PUT("/:username/password-expired-at",
			authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleUserUpdate)),
			userHandler.UpdatePasswordExpiration)
		userGroup.PUT("/:username/attributes",
			authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleUserUpdate)),
			userHandler.SetAttributes)
	}

	// /users/metadata - user additional field configuration
	usersMetadata := router.Group("/users/metadata")
	{
		usersMetadata.GET("/config",
			authhandler.GetAuthenticationHandler(true),
			userHandler.GetUserMetadataConfig)
		usersMetadata.GET("/config/history",
			authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin)),
			userHandler.GetUserMetadataConfigHistory)
		usersMetadata.PUT("/config",
			authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin)),
			userHandler.SaveUserMetadataConfig)
		usersMetadata.POST("/config/rollback",
			authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin)),
			userHandler.RollbackUserMetadataConfig)
	}
}
