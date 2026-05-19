package version

import (
	"log/slog"

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
	m := Model{Config: &a.config}
	if err := m.UpsertVersion(); err != nil {
		slog.Error("failed to record version into database", "error", err)
		// Do not fail server startup because version recording is non-critical
	}

	return nil
}

func (a *Api) Unload() {}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	routes.GET("", GetHandler(a.config))
}

func (a *Api) GetRelativePath() string {
	return "/version"
}
