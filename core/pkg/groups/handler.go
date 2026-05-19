package groups

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/ory/fosite/token/jwt"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/internal/casbin"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/shared/types/role"
)

type Group struct {
	jwtClaims *jwt.JWTClaims

	Name string `json:"name" db:"name"`
}

func (group *Group) GetHandler(c *gin.Context) (int, any) {
	if len(c.Param("group-name")) != 0 {
		return group.getHandler(c)
	} else {
		return group.getsHandler(c)
	}
}

func (group *Group) PostHandler(c *gin.Context) (int, any) {
	if len(group.jwtClaims.Subject) == 0 {
		return http.StatusBadRequest, external.ErrorResponse{Message: "empty subject"}
	}

	if body, err := io.ReadAll(c.Request.Body); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	} else if err := json.Unmarshal(body, group); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	if len(group.Name) == 0 {
		return http.StatusBadRequest, external.ErrorResponse{Message: "invalid name"}
	}

	if policies, err := enforcer.GetFilteredPolicy(0, group.jwtClaims.Subject); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	} else if len(policies) != 0 {
		return http.StatusBadRequest, external.ErrorResponse{Message: "user already in the group"}
	}

	if policies, err := enforcer.GetFilteredPolicy(1, group.Name); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	} else if len(policies) != 0 {
		return http.StatusBadRequest, external.ErrorResponse{Message: "name that already exists"}
	}

	policy := casbin.Policy{group.jwtClaims.Subject, group.Name, casbin.ActionAdmin}
	if err := enforcer.AddPolicy(policy); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

func (group *Group) DeleteHandler(c *gin.Context) (int, any) {
	group.Name = c.Param("group-name")

	if !common.GetRole(group.jwtClaims, string(role.RoleSuperAdmin)) {
		policies := []casbin.Policy{
			{group.jwtClaims.Subject, group.Name, casbin.ActionAdmin},
		}
		if _, err := enforcer.Enforce(policies...); errors.Is(err, external.ErrorNoSuchPolicy) {
			return http.StatusForbidden, nil
		} else if err != nil {
			return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
		}
	}

	if err := enforcer.RemovePolicyFromField(1, group.Name); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

func (group *Group) getHandler(c *gin.Context) (int, any) {
	group.Name = c.Param("group-name")

	if response, err := group.makeGetResponse(); errors.Is(err, external.ErrorNoSuchPolicy) {
		return http.StatusForbidden, nil
	} else if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	} else {
		return http.StatusOK, response
	}
}

func (group *Group) getsHandler(_ *gin.Context) (int, any) {
	policies, err := enforcer.GetFilteredPolicy(0, group.jwtClaims.Subject)
	if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	response := map[string]any{}

	for _, policy := range policies {
		if len(policy) != 3 {
			continue
		}

		g := Group{jwtClaims: group.jwtClaims, Name: policy[1]}

		if getResponse, err := g.makeGetResponse(); errors.Is(err, external.ErrorNoSuchPolicy) {
			continue
		} else if err != nil {
			return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
		} else {
			response[g.Name] = getResponse
		}
	}

	return http.StatusOK, response
}

func (group *Group) makeGetResponse() (map[string]any, error) {
	action := casbin.ActionAdmin
	if !common.GetRole(group.jwtClaims, string(role.RoleSuperAdmin)) {
		policies := []casbin.Policy{
			{group.jwtClaims.Subject, group.Name, casbin.ActionAdmin},
			{group.jwtClaims.Subject, group.Name, casbin.ActionMember},
		}
		if policy, err := enforcer.Enforce(policies...); err != nil {
			return nil, err
		} else {
			action = policy[2]
		}
	}

	if policies, err := enforcer.GetFilteredPolicy(1, group.Name); err != nil {
		return nil, err
	} else {
		return map[string]any{"permission": action, "members": len(policies)}, nil
	}
}

type Member struct {
	jwtClaims *jwt.JWTClaims
	action    string

	Subject string `json:"subject"`
	Object  string `json:"object"`
	Action  string `json:"action"`
}

func (member *Member) GetHandler(_ *gin.Context) (int, any) {
	policies, err := enforcer.GetFilteredPolicy(1, member.Object)
	if err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	response := []Member{}
	for _, policy := range policies {
		if len(policy) != 3 {
			continue
		}

		response = append(response, Member{
			Subject: policy[0],
			Object:  policy[1],
			Action:  policy[2],
		})
	}

	return http.StatusOK, response
}

func (member *Member) PutHandler(c *gin.Context) (int, any) {
	if member.action != casbin.ActionAdmin {
		return http.StatusForbidden, nil
	}

	if err := member.setFromReader(c.Request.Body); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	if policies, err := enforcer.GetFilteredPolicy(0, member.Subject); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	} else if len(policies) != 0 && policies[0][1] != member.Object {
		return http.StatusBadRequest, external.ErrorResponse{Message: "user already in the group"}
	}

	if err := member.removePolicies(member.Subject, member.Object); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	policy := casbin.Policy{member.Subject, member.Object, member.Action}
	if err := enforcer.AddPolicy(policy); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

func (member *Member) DeleteHandler(c *gin.Context) (int, any) {
	if member.action != casbin.ActionAdmin {
		return http.StatusForbidden, nil
	}

	if err := member.removePolicies(c.Param("user-name"), member.Object); err != nil {
		return http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()}
	}

	return http.StatusOK, nil
}

func (member *Member) setFromReader(reader io.Reader) error {
	if body, err := io.ReadAll(reader); err != nil {
		return err
	} else if err := json.Unmarshal(body, member); err != nil {
		return err
	}

	if len(member.Subject) == 0 {
		return external.ErrorEmptySubject
	} else if len(member.Action) == 0 {
		return external.ErrorEmptyAction
	}

	return nil
}

func (member *Member) removePolicies(subject, object string) error {
	policies, err := enforcer.GetFilteredPolicy(0, subject)
	if err != nil {
		return err
	}

	for _, policy := range policies {
		if len(policy) > 2 && policy[1] != object {
			continue
		}

		if err := enforcer.RemovePolicy(policy); err != nil {
			return err
		}
	}

	return nil
}

func groupsHandler(c *gin.Context) {
	group := Group{}

	if value, exists := c.Get(common.ContextKeyJWTClaims); exists {
		group.jwtClaims = value.(*jwt.JWTClaims)
	} else {
		group.jwtClaims = &jwt.JWTClaims{}
	}

	switch c.Request.Method {
	case http.MethodGet:
		c.JSON(group.GetHandler(c))
	case http.MethodPost:
		c.JSON(group.PostHandler(c))
	case http.MethodDelete:
		c.JSON(group.DeleteHandler(c))
	default:
		c.JSON(http.StatusNotImplemented, nil)
	}
}

func membersHandler(c *gin.Context) {
	member := Member{}

	if value, exists := c.Get(common.ContextKeyJWTClaims); exists {
		member.jwtClaims = value.(*jwt.JWTClaims)
	} else {
		member.jwtClaims = &jwt.JWTClaims{}
	}

	member.Object = c.Param("group-name")
	if len(member.Object) == 0 {
		c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: "group-name is empty"})
		return
	}

	member.action = casbin.ActionAdmin
	if !common.GetRole(member.jwtClaims, string(role.RoleSuperAdmin)) {
		policies := []casbin.Policy{
			{member.jwtClaims.Subject, member.Object, casbin.ActionAdmin},
			{member.jwtClaims.Subject, member.Object, casbin.ActionMember},
		}
		if policy, err := enforcer.Enforce(policies...); errors.Is(err, external.ErrorNoSuchPolicy) {
			c.JSON(http.StatusForbidden, nil)
			return
		} else if err != nil {
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		} else {
			member.action = policy[2]
		}
	}

	switch c.Request.Method {
	case http.MethodGet:
		c.JSON(member.GetHandler(c))
	case http.MethodPut:
		c.JSON(member.PutHandler(c))
	case http.MethodDelete:
		c.JSON(member.DeleteHandler(c))
	default:
		c.JSON(http.StatusNotImplemented, nil)
	}
}
