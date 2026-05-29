package homeupm

import (
	"log/slog"
	"net/http"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
	"ntels.com/pharos/core/pkg/common"
)

type homeUPMSection struct {
	Datasource string `toml:"datasource"`
}

type homeUPMTomlConfig struct {
	HomeUPM homeUPMSection `toml:"home_upm"`
}

type ConfigResponse struct {
	Datasource string `json:"datasource"`
}

type Api struct {
	datasource string
}

func (a *Api) Init(configPath string, _ common.Config) {
	if configPath == "" {
		slog.Warn("home-upm: config path is empty")
		return
	}

	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Warn("home-upm: failed to read config file", "path", configPath, "error", err)
		return
	}

	var cfg homeUPMTomlConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		slog.Warn("home-upm: failed to parse [home_upm] config", "error", err)
		return
	}

	a.datasource = cfg.HomeUPM.Datasource
	if a.datasource == "" {
		slog.Warn("home-upm: [home_upm] datasource is not configured")
	} else {
		slog.Info("home-upm: datasource loaded", "datasource", a.datasource)
	}
}

func (a *Api) Use() bool { return true }

func (a *Api) Load() error { return nil }

func (a *Api) Unload() {}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	routes.GET("", a.getConfig)
}

func (a *Api) GetRelativePath() string {
	return "/home-upm"
}

func (a *Api) getConfig(c *gin.Context) {
	c.JSON(http.StatusOK, ConfigResponse{
		Datasource: a.datasource,
	})
}
