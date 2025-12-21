package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	App AppConfig
}

type AppConfig struct {
	Port                  string        `envconfig:"SERVER_PORT" default:"8080"`
	IsDebugMode           bool          `envconfig:"DEBUG" default:"true"`
	ClientsConnectionsMax int32         `envconfig:"CLIENTS_CONN_MAX" default:"10"`
	KeepAlivePeriod       time.Duration `envconfig:"KEEP_ALIVE_PERIOD" default:"5s"`
}

func NewConfig() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
