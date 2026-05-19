package notification

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/ory/fosite/token/jwt"
	"ntels.com/pharos/core/external"
	"ntels.com/pharos/core/external/orm"
	pkgcom "ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/notification/adapters"
	"ntels.com/pharos/core/pkg/notification/common"
	"ntels.com/pharos/core/pkg/notification/model"
	"ntels.com/pharos/core/pkg/notification/resources"
	"ntels.com/pharos/core/pkg/notification/rule"
	"ntels.com/pharos/shared/types/notification"
	"ntels.com/pharos/shared/types/role"
)

var handlerMutex = sync.Mutex{}

func GetPostRuleHandler(config pkgcom.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		handlerMutex.Lock()
		defer handlerMutex.Unlock()

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			slog.Error("read body failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		var apiRule notification.NotificationRule
		err = json.Unmarshal(body, &apiRule)
		if err != nil {
			slog.Error("json.Unmarshal failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		// Convert API rule to internal rule configuration using adapter
		ruleAdapter := &adapters.RuleAdapter{}
		ruleConfig, err := ruleAdapter.FromAPIRuleCreate(apiRule)
		if err != nil {
			slog.Error("convert API rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		// Create resources.Rule from rule config
		res := &resources.Rule{
			NotificationType: common.Type(apiRule.NotificationType),
			Rule:             ruleConfig,
		}

		if res.Rule == nil {
			slog.Error("rule is nil")
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "rule is nil"})
			return
		}

		newRule, err := rule.GetRuleTemplate(res.NotificationType)
		if err != nil {
			slog.Error("get rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		err = newRule.Load(config, res.Rule)
		if err != nil {
			slog.Error("load rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		if newRule.GetId() != "" {
			slog.Error("rule id already exists", "id", newRule.GetId())
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "rule id already exists"})
			return
		}

		err = newRule.Validate()
		if err != nil {
			slog.Error("validate rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		newRule.SetId(uuid.New().String())

		m := model.Model{
			Config: &config,
		}

		// database error로 판단하기에는 db 종류에 따라 에러 여부를 판단해야 하기 때문에 사전에 확인하는것으로 변경
		getRule, err := m.GetRuleByName(newRule.GetName())
		if err != nil {
			slog.Error("alert get rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		} else if getRule != nil {
			slog.Error("rule name is duplicated")
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: "rule name is duplicated"})
			return
		}

		cur := time.Now().UTC()

		strRule, _ := json.Marshal(newRule)
		insertRule := model.Rule{
			ID:               newRule.GetId(),
			NotificationType: string(res.NotificationType),
			Name:             newRule.GetName(),
			Rule:             string(strRule),
			Timestamp:        orm.Datetime{Time: cur},
			UpdatedAt:        orm.Datetime{Time: cur},
		}
		err = m.InsertRule(&insertRule)
		if err != nil {
			slog.Error("alert rule insert failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		// rule은 insert rule에서 사전체크됨
		err = rule.Register(newRule)
		if err != nil {
			slog.Error("register rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		err = newRule.Run(c.Request.Context())
		if err != nil {
			slog.Error("alert rule run failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		c.JSON(http.StatusOK, map[string]string{"id": newRule.GetId()})
	}
}

func GetRuleListHandler(config pkgcom.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Optional query by name: /rule?name=<rule_name>
		if name := c.Query("name"); name != "" {
			m := model.Model{Config: &config}
			getRule, err := m.GetRuleByName(name)
			if err != nil {
				slog.Error("alert get rule by name failed", "error", err, "name", name)
				c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
				return
			}
			if getRule == nil {
				c.JSON(http.StatusNotFound, external.ErrorResponse{Message: "rule not found"})
				return
			}

			// Use adapter to convert model to API response
			ruleAdapter := adapters.NewRuleAdapter()
			res, err := ruleAdapter.ModelToResource(getRule)
			if err != nil {
				slog.Error("adapter conversion failed", "error", err)
				c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
				return
			}

			// Convert to shared/types response
			apiRes, err := adapters.ToAPIRuleResponseFromInternal(res)
			if err != nil {
				slog.Error("API conversion failed", "error", err)
				c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
				return
			}

			c.JSON(http.StatusOK, apiRes)
			return
		}

		var claims *jwt.JWTClaims
		if value, exists := c.Get(pkgcom.ContextKeyJWTClaims); exists {
			claims = value.(*jwt.JWTClaims)
		} else {
			claims = &jwt.JWTClaims{}
		}

		m := model.Model{
			Config: &config,
		}

		ruleDbs, err := m.GetAllRuleDb()
		if err != nil {
			slog.Error("alert get rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		// Use adapter to convert model list to API response list
		ruleAdapter := adapters.NewRuleAdapter()

		var resList []notification.NotificationRule
		for _, ruleDb := range ruleDbs {
			if pkgcom.GetRole(claims, string(role.RoleSuperAdmin)) ||
				pkgcom.GetRole(claims, string(role.RoleNotificationRead)) {
				// Convert using adapter for full access
				appendRes, err := ruleAdapter.ModelToResource(&ruleDb)
				if err != nil {
					slog.Error("adapter conversion failed", "error", err)
					c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
					return
				}
				if appendRes != nil {
					// Convert to shared/types response
					apiRes, err := adapters.ToAPIRuleResponseFromInternal(appendRes)
					if err != nil {
						slog.Error("API conversion failed", "error", err)
						c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
						return
					}
					if apiRes != nil {
						resList = append(resList, *apiRes)
					}
				}
			} else {
				// Simple rule for limited access - create manually
				simpleRule := map[string]any{"name": ruleDb.Name, "id": ruleDb.ID}
				resourceRule := &resources.Rule{
					NotificationType: common.Type(ruleDb.NotificationType),
					Rule:             simpleRule,
					Timestamp:        ruleDb.Timestamp.Time,
					UpdatedAt:        ruleDb.UpdatedAt.Time,
				}

				// Convert to shared/types response
				apiRes, err := adapters.ToAPIRuleResponseFromInternal(resourceRule)
				if err != nil {
					slog.Error("API conversion failed", "error", err)
					c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
					return
				}
				if apiRes != nil {
					resList = append(resList, *apiRes)
				}
			}
		}

		c.JSON(http.StatusOK, resList)
	}
}

func GetRuleHandler(config pkgcom.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		id := c.Param("id")

		m := model.Model{
			Config: &config,
		}

		existsRule, err := m.GetRule(id)
		if err != nil {
			slog.Error("alert get rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		if existsRule == nil {
			slog.Error("rule not found", "id", id)
			c.JSON(http.StatusNotFound, external.ErrorResponse{Message: "rule not found"})
			return
		}

		// Use adapter to convert model to API response
		ruleAdapter := adapters.NewRuleAdapter()
		res, err := ruleAdapter.ModelToResource(existsRule)
		if err != nil {
			slog.Error("adapter conversion failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		// Convert to shared/types response
		apiRes, err := adapters.ToAPIRuleResponseFromInternal(res)
		if err != nil {
			slog.Error("API conversion failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		c.JSON(http.StatusOK, apiRes)
	}
}

func GetPutRuleHandler(config pkgcom.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		handlerMutex.Lock()
		defer handlerMutex.Unlock()

		id := c.Param("id")

		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			slog.Error("read body failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		var apiRuleUpdate notification.NotificationRule
		err = json.Unmarshal(body, &apiRuleUpdate)
		if err != nil {
			slog.Error("json.Unmarshal failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		// Convert API rule update to internal rule configuration using adapter
		ruleAdapter := &adapters.RuleAdapter{}
		ruleConfig, err := ruleAdapter.FromAPIRuleUpdate(apiRuleUpdate)
		if err != nil {
			slog.Error("convert API rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		// Create resources.Rule from rule config
		newRes := &resources.Rule{
			NotificationType: common.Type(apiRuleUpdate.NotificationType),
			Rule:             ruleConfig,
		}

		m := model.Model{
			Config: &config,
		}

		editeRule, err := m.GetRule(id)
		if err != nil {
			slog.Error("alert get rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		if editeRule == nil {
			slog.Error("rule not found", "id", id)
			c.JSON(http.StatusNotFound, external.ErrorResponse{Message: "rule not found"})
			return
		}

		oldRule, err := rule.GetRuleFromStr(common.Type(editeRule.NotificationType), config, []byte(editeRule.Rule))
		if err != nil {
			slog.Error("get rule from rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		newRule, err := rule.GetRule(newRes.NotificationType, config, newRes.Rule)
		if err != nil {
			slog.Error("alert get rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}
		newRule.SetId(id)

		err = newRule.Validate()
		if err != nil {
			slog.Error("rule validate failed", "error", err)
			c.JSON(http.StatusBadRequest, external.ErrorResponse{Message: err.Error()})
			return
		}

		// 새로운 rule 상태 확인후에 제거
		err = rule.Unregister(oldRule.GetName())
		if err != nil {
			slog.Error("unregister rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
		}

		err = oldRule.Destroy()
		if err != nil {
			slog.Error("alert destroy rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		strRule, _ := json.Marshal(newRule)
		updateRule := model.Rule{
			ID:               newRule.GetId(),
			NotificationType: string(newRes.NotificationType),
			Name:             newRule.GetName(),
			Rule:             string(strRule),
			Timestamp:        orm.Datetime{Time: editeRule.Timestamp.Time},
			UpdatedAt:        orm.Datetime{Time: time.Now()},
		}

		err = m.UpdateRule(&updateRule)
		if err != nil {
			slog.Error("alert rule insert failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		err = newRule.Run(context.Background())
		if err != nil {
			slog.Error("alert rule run failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		err = rule.Register(newRule)
		if err != nil {
			slog.Error("register rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		c.Status(http.StatusOK)
	}
}

func GetDeleteRuleHandler(config pkgcom.Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		// destroy 구간이 존재하기 때문에 modify와 충돌을 대비해 lock 처리 (delete가 얼마 없어서 상관 없을것으로 보임)
		handlerMutex.Lock()
		defer handlerMutex.Unlock()

		id := c.Param("id")

		m := model.Model{
			Config: &config,
		}

		deleteRule, err := m.GetRule(id)
		if err != nil {
			slog.Error("alert get rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		oldRule, err := rule.GetRuleFromStr(common.Type(deleteRule.NotificationType), config, []byte(deleteRule.Rule))
		if err != nil {
			return
		}

		err = oldRule.Destroy()
		if err != nil {
			slog.Error("alert destroy rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}

		err = m.DeleteRule(id)
		if err != nil {
			slog.Error("alert delete rule failed", "error", err)
			c.JSON(http.StatusInternalServerError, external.ErrorResponse{Message: err.Error()})
			return
		}
	}
}
