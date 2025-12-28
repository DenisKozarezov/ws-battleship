package config

import (
	"time"

	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	App  AppConfig
	Game GameConfig
}

type AppConfig struct {
	ServerHost   string `envconfig:"SERVER_HOST" default:"127.0.0.1:8080"`
	IsDebugMode  bool   `envconfig:"DEBUG" default:"true"`
	MouseEnabled bool   `envconfig:"ENABLE_MOUSE" default:"false"`
}

type GameConfig struct {
	TurnTime time.Duration `envconfig:"GAME_TURN_TIME" default:"30s"`
}

func NewConfig() (*Config, error) {
	var cfg Config

	if err := envconfig.Process("", &cfg); err != nil {
		return nil, err
	}

	return &cfg, nil
}
