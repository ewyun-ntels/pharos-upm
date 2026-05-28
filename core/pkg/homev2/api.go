package homev2

import (
	"github.com/gin-gonic/gin"
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
	routes.GET("", GetConfigHandler(a.config))
}

func (a *Api) GetRelativePath() string {
	return "/home-v2"
}
