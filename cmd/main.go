package main

import (
	"github.com/peacecwz/scaleway-metric-exporter/collectors"
	"github.com/peacecwz/scaleway-metric-exporter/config"
	"github.com/peacecwz/scaleway-metric-exporter/pkg/http"
	"github.com/peacecwz/scaleway-metric-exporter/pkg/scaleway"
	log "github.com/sirupsen/logrus"
)

func main() {
	cfg := config.InitConfig()
	client := scaleway.EnsureCreateClient(cfg)
	registeredCollectors := collectors.RegisterCollectors(cfg, client)

	server := http.CreateHttpServer(cfg, registeredCollectors)

	log.Infof("starting http server on port %d", cfg.HttpPort)
	err := server.ListenAndServe()

	if err != nil {
		log.Panicf("failed to start http server %v", err)
	}
}
