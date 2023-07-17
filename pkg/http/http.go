package http

import (
	"fmt"
	"github.com/peacecwz/scaleway-metric-exporter/config"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"net/http"
	"time"
)

func CreateHttpServer(cfg *config.Config, r *prometheus.Registry) *http.Server {
	server := &http.Server{
		Addr:              fmt.Sprintf(":%d", cfg.HttpPort),
		ReadHeaderTimeout: 5 * time.Second,
	}

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("Running..."))
	})
	http.Handle(cfg.MetricEndpoint, promhttp.HandlerFor(r, promhttp.HandlerOpts{}))

	return server
}
