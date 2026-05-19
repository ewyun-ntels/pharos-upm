package api

import (
	"sync"

	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/alert"
	"ntels.com/pharos/core/pkg/authhandler"
	"ntels.com/pharos/core/pkg/badges"
	"ntels.com/pharos/core/pkg/common"
	"ntels.com/pharos/core/pkg/dashboard"
	"ntels.com/pharos/core/pkg/folders"
	"ntels.com/pharos/core/pkg/groups"
	"ntels.com/pharos/core/pkg/metrics"
	"ntels.com/pharos/core/pkg/notification"
	"ntels.com/pharos/core/pkg/plugins"
	"ntels.com/pharos/core/pkg/rolehandler"
	"ntels.com/pharos/core/pkg/ui_config"
	"ntels.com/pharos/core/pkg/userhandler"
	"ntels.com/pharos/core/pkg/version"
	"ntels.com/pharos/core/pkg/websocket"
	"ntels.com/pharos/core/pkg/workflow"
)

var addApi []Api
var apiMutex sync.RWMutex

type Api interface {
	Init(string, common.Config)

	Use() bool

	Load() error
	Unload()

	RegisterRoutes(gin.IRoutes)

	GetRelativePath() string
}

func AddApi(api Api) {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	addApi = append(addApi, api)
}

func GetApis(configPath string, config common.Config) []Api {
	apiMutex.Lock()
	defer apiMutex.Unlock()

	var allApi = []Api{
		&version.Api{},
		&badges.Api{},
		&authhandler.Api{},
		&alert.Api{},
		&notification.Api{},
		&workflow.Api{},
		&websocket.Api{},
		&dashboard.Api{},
		&ui_config.Api{},
		&plugins.Api{},
		&folders.Api{},
		&groups.Api{},
		&userhandler.Api{},
		&rolehandler.Api{},
		&metrics.Api{},
	}
	allApi = append(allApi, addApi...)

	var apis []Api

	for index := range allApi {
		allApi[index].Init(configPath, config)

		if allApi[index].Use() {
			apis = append(apis, allApi[index])
		}
	}

	return apis
}
