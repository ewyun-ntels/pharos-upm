package ui_config

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
	routes.GET("/config", getConfigHandler(a.config))
	routes.POST("/config", getConfigHandler(a.config))
	routes.PUT("/config", getConfigHandler(a.config))
	routes.DELETE("/config", getConfigHandler(a.config))

	routes.GET("/home", getHomeHandler(a.config))
	routes.PUT("/home", getHomeHandler(a.config))
	routes.DELETE("/home", getHomeHandler(a.config))

	routes.GET("/images", imagesHandler)
	routes.GET("/images/:name", imagesHandler)
}

func (a *Api) GetRelativePath() string {
	return "/ui-config"
}
