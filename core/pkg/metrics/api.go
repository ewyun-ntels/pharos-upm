package metrics

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	prometheus_collectors "github.com/prometheus/client_golang/prometheus/collectors"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"ntels.com/pharos/core/pkg/common"
)

type Api struct {
	configPath string
	config     common.Config

	registry *prometheus.Registry
}

func (a *Api) Init(configPath string, config common.Config) {
	a.configPath = configPath
	a.config = config

	a.registry = prometheus.NewRegistry()
}

func (a *Api) Use() bool {
	return a.config.Metrics.Use
}

func (a *Api) Load() error {
	a.registry = prometheus.NewRegistry()

	// Standard Go collectors
	a.registry.MustRegister(prometheus_collectors.NewGoCollector())
	a.registry.MustRegister(prometheus_collectors.NewProcessCollector(prometheus_collectors.ProcessCollectorOpts{}))

	for _, collector := range collectors.GetAll() {
		a.registry.MustRegister(collector)
	}

	return nil
}

func (a *Api) Unload() {
	for _, collector := range collectors.GetAll() {
		a.registry.Unregister(collector)
	}
}

func (a *Api) RegisterRoutes(routes gin.IRoutes) {
	routes.GET("/", gin.WrapH(promhttp.HandlerFor(a.registry, promhttp.HandlerOpts{})))
}

func (a *Api) GetRelativePath() string {
	return "/metrics"
}
