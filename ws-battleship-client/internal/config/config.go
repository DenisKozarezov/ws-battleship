package config

import (
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	App AppConfig
}

type AppConfig struct {
	Port         string `envconfig:"SERVER_PORT" default:"8080"`
	IsDebugMode  bool   `envconfig:"DEBUG" default:"true"`
	MouseEnabled bool   `envconfig:"ENABLE_MOUSE" default:"false"`
}

func NewConfig() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
