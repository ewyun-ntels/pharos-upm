package handler

import (
	"encoding/json"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite/token/jwt"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/internal/casbin"
	"ntels.com/pharos/shared/types/dashboard"
)

const (
	PermissionKindGroup = "group"
	PermissionKindUser  = "user"
)

type Permission struct {
	JwtClaims *jwt.JWTClaims   `json:"-"`
	Object    string           `json:"object"`
	Subject   string           `json:"subject"`
	Action    dashboard.Action `json:"action"`
	Kind      string           `json:"kind"`

	// Internal dashboard config (separate from API schema)
	DashboardConfig *dashboard.DashboardConfig `json:"-"`
}

func (permission *Permission) GetHandler(_ *gin.Context) (int, any) {
	policies, err := enforcer.GetFilteredPolicy(1, permission.Object)
	if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	// Convert internal permissions to API format
	apiResponse := []dashboard.DashboardPermission{}
	for _, policy := range policies {
		if len(policy) != 4 {
			continue
		}

		apiResponse = append(apiResponse, dashboard.DashboardPermission{
			Subject: policy[0],
			Object:  policy[1],
			Action:  dashboard.Action(policy[2]),
			Kind:    policy[3],
		})
	}

	return http.StatusOK, apiResponse
}

func (permission *Permission) PutHandler(c *gin.Context) (int, any) {
	if err := permission.setFromReader(c.Request.Body); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	if err := permission.removePolicies(); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	policy := casbin.Policy{
		permission.Subject,
		permission.Object,
		string(permission.Action),
		permission.Kind,
	}
	if err := enforcer.AddPolicy(policy); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

func (permission *Permission) DeleteHandler(c *gin.Context) (int, any) {
	permission.Subject = c.Param("subject")

	// Code for compatibility with existing frontend when backend is applied first
	// permission.Kind = c.Query("kind")
	permission.Kind = c.DefaultQuery("kind", PermissionKindUser)

	if err := permission.removePolicies(); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

func (permission *Permission) setFromReader(reader io.Reader) error {
	body, err := io.ReadAll(reader)
	if err != nil {
		return err
	}

	// Parse as API format first
	var apiPermission dashboard.DashboardPermission
	if err := json.Unmarshal(body, &apiPermission); err != nil {
		return err
	}

	// Convert from API format to internal format
	if apiPermission.Subject != "" {
		permission.Subject = apiPermission.Subject
	}
	if apiPermission.Object != "" {
		permission.Object = apiPermission.Object
	}
	permission.Action = apiPermission.Action
	permission.Kind = apiPermission.Kind

	if len(permission.Subject) == 0 {
		return external.ErrorEmptySubject
	} else if len(permission.Action) == 0 {
		return external.ErrorEmptyAction
	}

	// Code for compatibility with existing frontend when backend is applied first
	if len(permission.Kind) == 0 {
		permission.Kind = PermissionKindUser
	}

	return nil
}

func (permission *Permission) removePolicies() error {
	policies, err := enforcer.GetFilteredPolicy(0, permission.Subject)
	if err != nil {
		return err
	}

	for _, policy := range policies {
		if len(policy) >= 4 && (policy[1] != permission.Object || policy[3] != permission.Kind) {
			continue
		}

		if err := enforcer.RemovePolicy(policy); err != nil {
			return err
		}
	}

	return nil
}
