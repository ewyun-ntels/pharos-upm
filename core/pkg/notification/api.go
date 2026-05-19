package notification

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/authhandler"
	pkgcom "ntels.com/pharos/core/pkg/common"
	notification_common "ntels.com/pharos/core/pkg/notification/common"
	"ntels.com/pharos/core/pkg/notification/model"
	"ntels.com/pharos/core/pkg/notification/rule"
	"ntels.com/pharos/shared/types/role"
)

type Api struct {
	configPath string
	config     pkgcom.Config
}

func (a *Api) Init(configPath string, config pkgcom.Config) {
	a.configPath = configPath
	a.config = config
}

func (a *Api) Use() bool {
	return true
}

func (a *Api) Load() error {
	if err := a.loadRules(); err != nil {
		slog.Error("load rules failed", "error", err)
		return err
	}

	return nil
}

func (a *Api) Unload() {}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	routes.POST("/rule",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleNotificationCreate)),
		GetPostRuleHandler(a.config))
	routes.GET("/rule",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleNotificationRead)),
		GetRuleListHandler(a.config))
	routes.GET("/rule/:id",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleNotificationRead)),
		GetRuleHandler(a.config))
	routes.PUT("/rule/:id",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleNotificationUpdate)),
		GetPutRuleHandler(a.config))
	routes.DELETE("/rule/:id",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleNotificationDelete)),
		GetDeleteRuleHandler(a.config))
}

func (a *Api) GetRelativePath() string {
	return "/notification"
}

func (a *Api) loadRules() error {
	if isLoading {
		return nil
	}
	isLoading = true

	m := model.Model{
		Config: &a.config,
	}

	rules, err := m.GetAllRuleDb()
	if err != nil {
		slog.Error("load rules failed", "err", err)
		return err
	}

	for _, r := range rules {
		var decRule map[string]any
		err = json.Unmarshal([]byte(r.Rule), &decRule)
		if err != nil {
			slog.Error("alert rule decode failed", "error", err)
			return err
		}

		template, err := rule.GetRuleTemplate(notification_common.Type(r.NotificationType))
		if err != nil {
			slog.Error("alert rule template decode failed", "error", err)
			return err
		}

		err = template.Load(a.config, decRule)
		if err != nil {
			slog.Error("alert rule decode failed", "error", err)
			return err
		}

		err = template.Run(context.Background())
		if err != nil {
			slog.Error("alert rule run failed", "error", err)
			return err
		}

		err = rule.Register(template)
		if err != nil {
			return err
		}
	}

	return nil
}
