package alert

import (
	"context"
	"encoding/json"
	"log/slog"

	"github.com/gin-gonic/gin"
	alert_common "ntels.com/pharos/core/pkg/alert/common"
	"ntels.com/pharos/core/pkg/alert/model"
	"ntels.com/pharos/core/pkg/alert/rule"
	alert_rule_common "ntels.com/pharos/core/pkg/alert/rule/common"
	"ntels.com/pharos/core/pkg/authhandler"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/shared/types/role"
)

type Api struct {
	configPath string
	config     common.Config
}

func (a *Api) Init(configPath string, config common.Config) {
	a.configPath = configPath
	a.config = config
}

func (a *Api) Use() bool {
	return true
}

func (a *Api) Load() error {
	if err := alert_rule_common.CronRun(); err != nil {
		slog.Error("cron instance run failed", "error", err)
		return err
	}

	if err := a.loadRules(); err != nil {
		slog.Error("load rules failed", "error", err)
		return err
	}

	return nil
}

func (a *Api) Unload() {
}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	routes.GET("/status", authhandler.GetAuthenticationHandler(true), GetStatusHandler(a.config))
	routes.POST("/rule",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleAlertCreate)), GetPostRuleHandler(a.config))
	routes.GET("/rule",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleAlertRead)), GetRuleListHandler(a.config))
	routes.GET("/rule/:id",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleAlertRead)), GetRuleHandler(a.config))
	routes.PUT("/rule/:id",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleAlertUpdate)), GetPutRuleHandler(a.config))
	routes.DELETE("/rule/:id",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleAlertDelete)), GetDeleteRuleHandler(a.config))
	routes.POST("/query",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleAlertRead)), PostQuery) // Alert query를 테스트하기 위해 사용
	routes.GET("/hist",
		authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleAlertRead)), GetHistQueryHandler(a.config))
	routes.POST("/event/:name", GetPostEventHandler(a.config)) // TODO 외부 연동이슈로 인하여 일단 disable
	routes.PUT("/status/:name/:id/mask", authhandler.GetAuthenticationHandler(true), GetPutStatusMaskHandler(a.config))
	routes.DELETE("/status/:name/:id", authhandler.GetAuthenticationHandler(true), GetPutStatusCleanHandler(a.config))
}

func (a *Api) GetRelativePath() string {
	return "/alert"
}

func (a *Api) loadRules() error {
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

		template, err := rule.GetRuleTemplate(alert_common.Type(r.AlertType))
		if err != nil {
			slog.Error("alert rule template decode failed", "error", err)
			return err
		}

		err = template.Load(a.config, decRule)
		if err != nil {
			slog.Error("alert rule decode failed", "error", err)
			return err
		}
		err = template.Validate()
		if err != nil {
			slog.Error("alert rule validate failed", "error", err)
			return err
		}

		err = template.Run(context.Background())
		if err != nil {
			slog.Error("alert rule run failed", "error", err)
			return err
		}
	}

	return nil
}
