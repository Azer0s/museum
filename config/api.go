package config

import (
	"github.com/caarlos0/env/v7"
	"go.uber.org/zap"
	"museum/config/impl"
)

func NewEnvConfig(log *zap.SugaredLogger) *impl.EnvConfig {
	cfg := &impl.EnvConfig{}
	err := env.Parse(cfg)
	if err != nil {
		log.Panic(err)
	}
	log.Debugw("config loaded", "config", cfg)
	return cfg
}
