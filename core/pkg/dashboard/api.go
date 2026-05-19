package dashboard

import (
	"sync"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/internal/casbin"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/dashboard/handler"
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
		enforcer, err = casbin.NewEnforcer(casbin.EnforcerTypeDashboard, a.config.Database)
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
	handler.RegisterRoutes(a.config, enforcer, routes)
}

func (a *Api) GetRelativePath() string {
	return "/dashboard"
}
