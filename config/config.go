package config

import (
	"encoding/json"
	arg "github.com/alexflint/go-arg"
	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
	log "github.com/sirupsen/logrus"
	"gopkg.in/yaml.v3"
	"os"
	"path/filepath"
)

type Config struct {
	ScalewayAccessKey  string `json:"scalewayAccessKey" yaml:"scalewayAccessKey" env:"SCALEWAY_ACCESS_KEY" arg:"--scaleway-access-key"`
	ScalewaySecretKey  string `json:"scalewaySecretKey" yaml:"scalewaySecretKey" env:"SCALEWAY_SECRET_KEY" arg:"--scaleway-secret-key"`
	ScalewayRegion     string `json:"scalewayRegion" yaml:"scalewayRegion" env:"SCALEWAY_REGION" arg:"--scaleway-region"`
	ScalewayZone       string `json:"scalewayZone" yaml:"scalewayZone" env:"SCALEWAY_ZONE" arg:"--scaleway-zone"`
	ScalewayProjectID  string `json:"scalewayProjectID" yaml:"scalewayProjectID" env:"SCALEWAY_PROJECT_ID" arg:"--scaleway-project-id"`
	HttpPort           int    `json:"httpPort" yaml:"httpPort" env:"HTTP_PORT" arg:"--http-port"`
	MetricEndpoint     string `json:"metricEndpoint" yaml:"metricEndpoint" env:"METRIC_ENDPOINT" arg:"--metric-endpoint"`
	TimeoutInMs        int    `json:"timeoutInMs" yaml:"timeoutInMs" env:"TIMEOUT_IN_MS" arg:"--timeout-in-ms"`
	ExternalConfigFile string `json:"externalConfigFile" yaml:"externalConfigFile" env:"EXTERNAL_CONFIG_FILE" arg:"--external-config-file"`
}

func InitConfig() *Config {
	defaultCfg := Config{
		TimeoutInMs:        5 * 1000,
		HttpPort:           9706,
		MetricEndpoint:     "/metrics",
		ExternalConfigFile: "./config.yaml",
	}

	// Load config from env
	err := godotenv.Load()
	if err != nil {
		log.Warnf("failed to read .env file, %v", err)
	}

	if err := env.Parse(&defaultCfg); err != nil {
		log.Errorf("failed to parse env to struct. details: %+v", err)
	}

	// Load config from file
	configData, err := os.ReadFile(defaultCfg.ExternalConfigFile)
	if err != nil {
		log.Warnf("cannot read %s file. details: %+v", defaultCfg.ExternalConfigFile, err)
	}

	ext := filepath.Ext(defaultCfg.ExternalConfigFile)
	switch ext {
	case ".yaml", ".yml":
		err = yaml.Unmarshal(configData, &defaultCfg)
		if err != nil {
			log.Errorf("failed to unmarshall config. details: %+v", err)
		}
	case ".json":
		err = json.Unmarshal(configData, &defaultCfg)
		if err != nil {
			log.Errorf("failed to unmarshall config. details: %+v", err)
		}
	default:
		log.Errorf("unsupported config '%s' file format. details: %+v", ext, err)
	}

	// Load config from args
	err = arg.Parse(&defaultCfg)
	if err != nil {
		log.Errorf("failed to parse args to struct. details: %+v", err)
	}

	return &defaultCfg
}
