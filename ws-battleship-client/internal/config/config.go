package config

import "github.com/kelseyhightower/envconfig"

type Config struct {
	App AppConfig
}

type AppConfig struct {
	ServerHost  string `envconfig:"SERVER_HOST" default:"127.0.0.1:8080"`
	IsDebugMode bool   `envconfig:"DEBUG" default:"true"`
}

func NewConfig() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
