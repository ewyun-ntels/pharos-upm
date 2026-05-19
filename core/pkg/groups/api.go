package groups

import (
	"sync"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/internal/casbin"
	"ntels.com/pharos/core/pkg/authhandler"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/shared/types/role"
)

var once sync.Once
var enforcer *casbin.Enforcer

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
	var err error
	once.Do(func() {
		enforcer, err = casbin.NewEnforcer(casbin.EnforcerTypeGroup, a.config.Database)
		if err != nil {
			return
		}
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *Api) Unload() {}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	routes.GET("", authhandler.GetAuthenticationHandler(true), groupsHandler)
	routes.GET("/:group-name", authhandler.GetAuthenticationHandler(true), groupsHandler)
	routes.POST("", authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), string(role.RoleGroupAdd)), groupsHandler)
	routes.DELETE("/:group-name", authhandler.GetAuthenticationHandler(true), groupsHandler)

	routes.GET("/:group-name/members", authhandler.GetAuthenticationHandler(true), membersHandler)
	routes.PUT("/:group-name/members", authhandler.GetAuthenticationHandler(true), membersHandler)
	routes.DELETE("/:group-name/members/:user-name", authhandler.GetAuthenticationHandler(true), membersHandler)
}

func (a *Api) GetRelativePath() string {
	return "/groups"
}
