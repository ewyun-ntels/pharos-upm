package authhandler

import (
	"github.com/prometheus/client_golang/prometheus"
	"ntels.com/pharos/core/pkg/metrics"
)

var (
	descActiveSessions = prometheus.NewDesc(
		"pharos_auth_active_sessions",
		"Current number of active refresh token sessions.",
		nil, nil,
	)
	descExpiredTotal = prometheus.NewDesc(
		"pharos_auth_sessions_expired_total",
		"Total number of sessions expired by the GC.",
		nil, nil,
	)
	descRevokedTotal = prometheus.NewDesc(
		"pharos_auth_sessions_revoked_total",
		"Total number of sessions forcibly revoked.",
		nil, nil,
	)
)

type sessionCollector struct{}

func (c *sessionCollector) Describe(ch chan<- *prometheus.Desc) {
	ch <- descActiveSessions
	ch <- descExpiredTotal
	ch <- descRevokedTotal
}

func (c *sessionCollector) Collect(ch chan<- prometheus.Metric) {
	if oauth2Provider == nil {
		return
	}
	storage := oauth2Provider.Storage

	ch <- prometheus.MustNewConstMetric(descActiveSessions, prometheus.GaugeValue,
		float64(storage.GetActiveRefreshTokenCount()))
	ch <- prometheus.MustNewConstMetric(descExpiredTotal, prometheus.CounterValue,
		float64(storage.GetSessionExpiredTotal()))
	ch <- prometheus.MustNewConstMetric(descRevokedTotal, prometheus.CounterValue,
		float64(storage.GetSessionRevokedTotal()))
}

func init() {
	metrics.AddCollector("auth_sessions", &sessionCollector{})
}
