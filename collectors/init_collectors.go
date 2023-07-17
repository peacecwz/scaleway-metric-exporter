package collectors

import (
	"github.com/peacecwz/scaleway-metric-exporter/config"
	"github.com/peacecwz/scaleway-metric-exporter/pkg/scaleway"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

func RegisterCollectors(cfg *config.Config, client *scaleway.ScalewayClient) *prometheus.Registry {
	allCollectors := []prometheus.Collector{
		NewRDBCollector(cfg, client),
	}

	r := prometheus.NewRegistry()
	r.MustRegister(collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}))
	for _, c := range allCollectors {
		r.MustRegister(c)
	}

	return r
}
