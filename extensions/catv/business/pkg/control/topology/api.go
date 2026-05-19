package topology

import (
	"github.com/gin-gonic/gin"
	"ntels.com/pharos/core/pkg/authhandler"
	"ntels.com/pharos/core/pkg/common"
	control_common "ntels.com/pharos/extensions/catv/business/pkg/control/common"
	"ntels.com/pharos/extensions/catv/business/pkg/permissions"
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
	return nil
}

func (a *Api) Unload() {
}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	routes.GET("/topology/sos", authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), permissions.Read), getSosHandler(a.config))
	routes.GET("/topology/sos/:so_id/l3s", authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), permissions.Read), getL3sHandler(a.config))
	routes.GET("/topology/sos/:so_id/l3s/:l3_id/cells", authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), permissions.Read), getCellsHandler(a.config))
	routes.GET("/topology/sos/:so_id/l3s/:l3_id/cells/:cell_id/settopboxes", authhandler.GetAuthenticationHandler(true, string(role.RoleSuperAdmin), permissions.Read), getSettopboxesHandler(a.config))
}

func (a *Api) GetRelativePath() string {
	return control_common.HttpRelativePath
}
