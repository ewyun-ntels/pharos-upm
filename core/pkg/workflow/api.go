package workflow

import (
	"sync"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/external/goflow"
	"ntels.com/pharos/core/pkg/common"
)

var once sync.Once

type Api struct {
	configPath string
	config     common.Config
}

func (a *Api) Init(configPath string, config common.Config) {
	a.configPath = configPath
	a.config = config
}

func (a *Api) Use() bool {
	return a.config.Workflow.Use
}

func (a *Api) Load() error {
	var err error
	once.Do(func() {
		if _, exist := a.config.Servers[common.ServerTypeNats]; !exist {
			err = goflow.GlobalGoflow.Start(a.configPath, a.config)
			if err != nil {
				return
			}
		}
	})
	if err != nil {
		return err
	}

	return nil
}

func (a *Api) Unload() {
	goflow.GlobalGoflow.Stop()
}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	routes.GET("/task/:name", getTaskHandler)
	routes.GET("/task", getsTaskHandler)

	routes.GET("/job/:name", getJobHandler(a.configPath, a.config))
	routes.GET("/job", getJobHandler(a.configPath, a.config))
	routes.POST("/job", getJobHandler(a.configPath, a.config))
	routes.PUT("/job/:name", getJobHandler(a.configPath, a.config))
	routes.DELETE("/job/:name", getJobHandler(a.configPath, a.config))

	routes.GET("/execution/:job", getExecutionHandler(a.config))
	routes.GET("/execution", getExecutionsHandler(a.config))
}

func (a *Api) GetRelativePath() string {
	return "/workflow"
}
