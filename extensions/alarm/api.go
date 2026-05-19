package alarm

import (
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/pelletier/go-toml/v2"
	"ntels.com/pharos/core/pkg/authhandler"
	"ntels.com/pharos/core/pkg/common"
)

type alertaSection struct {
	URL string `toml:"url"`
}

type alarmTomlConfig struct {
	Alarm struct {
		Alerta alertaSection `toml:"alerta"`
	} `toml:"alarm"`
}

// Api proxies requests from the frontend to the Alerta REST API.
type Api struct {
	alertaURL string
}

func (a *Api) Init(configPath string, _ common.Config) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		slog.Warn("alarm: failed to read config file", "path", configPath, "error", err)
		return
	}
	var cfg alarmTomlConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		slog.Warn("alarm: failed to parse [alarm.alerta] config", "error", err)
		return
	}
	a.alertaURL = cfg.Alarm.Alerta.URL
	if a.alertaURL == "" {
		slog.Warn("alarm: [alarm.alerta] url is not configured")
	} else {
		slog.Info("alarm: alerta URL loaded", "url", a.alertaURL)
	}
}

func (a *Api) Use() bool { return true }

func (a *Api) Load() error { return nil }

func (a *Api) Unload() {}

func (a *Api) GetRelativePath() string { return "/alarm" }

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	auth := authhandler.GetAuthenticationHandler(true)
	routes.GET("/alerts", auth, a.proxyAlerts)
	routes.GET("/alerts/:id", auth, a.proxyAlertDetail)
	routes.GET("/config", auth, a.getConfig)
}

// proxyAlerts forwards GET /alarm/alerts to Alerta's /api/alerts endpoint.
func (a *Api) proxyAlerts(c *gin.Context) {
	if a.alertaURL == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "alerta URL not configured in [alarm.alerta]"})
		return
	}

	target := strings.TrimRight(a.alertaURL, "/") + "/api/alerts"
	if c.Request.URL.RawQuery != "" {
		target += "?" + c.Request.URL.RawQuery
	}

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, target, nil)
	if err != nil {
		slog.Error("alarm: failed to build alerta request", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("alarm: alerta request failed", "url", target, "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	c.Header("Content-Type", "application/json")
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
}

// proxyAlertDetail forwards GET /alarm/alerts/:id to Alerta's /api/alerts/:id endpoint.
func (a *Api) proxyAlertDetail(c *gin.Context) {
	if a.alertaURL == "" {
		c.JSON(http.StatusServiceUnavailable, gin.H{"error": "alerta URL not configured in [alarm.alerta]"})
		return
	}

	id := c.Param("id")
	target := strings.TrimRight(a.alertaURL, "/") + "/api/alert/" + id

	req, err := http.NewRequestWithContext(c.Request.Context(), http.MethodGet, target, nil)
	if err != nil {
		slog.Error("alarm: failed to build alerta detail request", "error", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		slog.Error("alarm: alerta detail request failed", "url", target, "error", err)
		c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	c.Header("Content-Type", "application/json")
	c.Status(resp.StatusCode)
	_, _ = io.Copy(c.Writer, resp.Body)
}

// getConfig exposes the Alerta public URL so the frontend can deep-link into the Alerta web UI.
func (a *Api) getConfig(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"alertaURL": a.alertaURL})
}
