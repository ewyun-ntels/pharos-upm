package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"ntels.com/pharos/core/internal"
)

var collectors = internal.NewMap[prometheus.Collector]()

func AddCollector(name string, collector prometheus.Collector) {
	collectors.Set(name, collector)
}
