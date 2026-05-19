package folders

import (
	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/authhandler"
	"ntels.com/pharos/core/pkg/common"
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
	return nil
}

func (a *Api) Unload() {}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	routes.Use(authhandler.GetAuthenticationHandler(false))
	RegisterRoutes(a.config, routes)
}

func (a *Api) GetRelativePath() string {
	return "/folders"
}
