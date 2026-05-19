package handler

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite/token/jwt"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/internal/casbin"
	"ntels.com/pharos/core/pkg/authhandler"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/shared/types/dashboard"
	"ntels.com/pharos/shared/types/role"
)

var config common.Config
var enforcer *casbin.Enforcer
var svc *DashboardService

func getGroupName(jwtClaims *jwt.JWTClaims) string {
	groups := common.GetGroups(jwtClaims)
	if len(groups) > 0 {
		return groups[0]
	}
	return ""
}

func dashboardHandler(c *gin.Context) {
	switch c.Request.Method {
	case http.MethodGet:
		c.JSON(svc.GetHandler(c))
	case http.MethodPost:
		c.JSON(svc.PostHandler(c))
	case http.MethodPut:
		c.JSON(svc.PutHandler(c))
	case http.MethodDelete:
		c.JSON(svc.DeleteHandler(c))
	default:
		c.JSON(http.StatusNotImplemented, nil)
	}
}

func permissionHandler(c *gin.Context) {
	permission := Permission{DashboardConfig: &dashboard.DashboardConfig{}, JwtClaims: getJWTClaims(c)}
	permission.Object = c.Param("id")

	if !common.GetRole(permission.JwtClaims, string(role.RoleSuperAdmin)) {
		groupName := getGroupName(permission.JwtClaims)

		policies := []casbin.Policy{
			{permission.JwtClaims.Subject, permission.Object, casbin.ActionOwner, PermissionKindUser},
			{permission.JwtClaims.Subject, permission.Object, casbin.ActionEditor, PermissionKindUser},
			{permission.JwtClaims.Subject, permission.Object, casbin.ActionViewer, PermissionKindUser},

			{groupName, permission.Object, casbin.ActionOwner, PermissionKindGroup},
			{groupName, permission.Object, casbin.ActionEditor, PermissionKindGroup},
			{groupName, permission.Object, casbin.ActionViewer, PermissionKindGroup},

			{casbin.SubjectPublic, permission.Object, casbin.ActionOwner, PermissionKindUser},
			{casbin.SubjectPublic, permission.Object, casbin.ActionEditor, PermissionKindUser},
			{casbin.SubjectPublic, permission.Object, casbin.ActionViewer, PermissionKindUser},
		}
		if policy, err := enforcer.Enforce(policies...); errors.Is(err, external.ErrorNoSuchPolicy) {
			c.JSON(http.StatusForbidden, nil)
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		} else if policy[2] != casbin.ActionOwner {
			c.JSON(http.StatusForbidden, nil)
			return
		}
	}

	switch c.Request.Method {
	case http.MethodGet:
		c.JSON(permission.GetHandler(c))
	case http.MethodPut:
		c.JSON(permission.PutHandler(c))
	case http.MethodDelete:
		c.JSON(permission.DeleteHandler(c))
	default:
		c.JSON(http.StatusNotImplemented, nil)
	}
}

func queryHandler(c *gin.Context) {
	query := Query{}

	switch c.Request.Method {
	case http.MethodPost:
		c.JSON(query.PostHandler(c))
	default:
		c.JSON(http.StatusNotImplemented, nil)
	}
}

func idsHandler(c *gin.Context) {
	var responseBody []map[string]string

	dashboards, err := svc.gets()
	if err != nil {
		c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
		return
	}

	jwtClaims := getJWTClaims(c)

	for _, temp := range dashboards {
		if !common.GetRole(jwtClaims, string(role.RoleSuperAdmin)) {
			groupName := getGroupName(jwtClaims)

			policies := []casbin.Policy{
				{jwtClaims.Subject, temp.ID, casbin.ActionOwner, PermissionKindUser},
				{jwtClaims.Subject, temp.ID, casbin.ActionEditor, PermissionKindUser},
				{jwtClaims.Subject, temp.ID, casbin.ActionViewer, PermissionKindUser},

				{groupName, temp.ID, casbin.ActionOwner, PermissionKindGroup},
				{groupName, temp.ID, casbin.ActionEditor, PermissionKindGroup},
				{groupName, temp.ID, casbin.ActionViewer, PermissionKindGroup},

				{casbin.SubjectPublic, temp.ID, casbin.ActionOwner, PermissionKindUser},
				{casbin.SubjectPublic, temp.ID, casbin.ActionEditor, PermissionKindUser},
				{casbin.SubjectPublic, temp.ID, casbin.ActionViewer, PermissionKindUser},
			}
			if _, err := enforcer.Enforce(policies...); errors.Is(err, external.ErrorNoSuchPolicy) {
				continue
			} else if err != nil {
				c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
				return
			}
		}

		responseBody = append(responseBody, map[string]string{"id": temp.ID})
	}

	c.JSON(http.StatusOK, responseBody)
}

func getJWTClaims(c *gin.Context) *jwt.JWTClaims {
	if value, exists := c.Get(common.ContextKeyJWTClaims); exists {
		return value.(*jwt.JWTClaims)
	} else {
		return &jwt.JWTClaims{}
	}
}

func RegisterRoutes(cfg common.Config, efc *casbin.Enforcer, routes gin.IRoutes) {
	config = cfg
	enforcer = efc
	svc = NewDashboardService(cfg, efc)

	routes.GET("", authhandler.GetAuthenticationHandler(false), dashboardHandler)
	routes.GET("/:id", authhandler.GetAuthenticationHandler(false), dashboardHandler)
	routes.POST("",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleDashboardCreate)),
		dashboardHandler)
	routes.PUT("/:id", authhandler.GetAuthenticationHandler(false), dashboardHandler)
	routes.DELETE("/:id", authhandler.GetAuthenticationHandler(false), dashboardHandler)

	routes.GET("/:id/permission", authhandler.GetAuthenticationHandler(false), permissionHandler)
	routes.PUT("/:id/permission", authhandler.GetAuthenticationHandler(false), permissionHandler)
	routes.DELETE("/:id/permission/:subject", authhandler.GetAuthenticationHandler(false), permissionHandler)
	routes.POST("/:id/query", authhandler.GetAuthenticationHandler(false), queryHandler)

	routes.GET("/ids", authhandler.GetAuthenticationHandler(false), idsHandler)
	routes.GET("/favorites", authhandler.GetAuthenticationHandler(false), getFavoritesHandler)
	routes.PATCH("/:id/favorite", authhandler.GetAuthenticationHandler(false), patchFavoriteHandler)
	routes.PATCH("/:id/folder", authhandler.GetAuthenticationHandler(false), func(c *gin.Context) {
		c.JSON(svc.PatchFolderHandler(c))
	})
	routes.GET("/history", authhandler.GetAuthenticationHandler(false), func(c *gin.Context) {
		c.JSON(svc.HistoryHandler(c))
	})
	routes.PUT("/:id/annotations", authhandler.GetAuthenticationHandler(false), func(c *gin.Context) {
		c.JSON(svc.PutAnnotationsHandler(c))
	})
}
