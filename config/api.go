package config

import (
	"github.com/caarlos0/env/v7"
	"museum/config/impl"
)

func NewEnvConfig() *impl.EnvConfig {
	cfg := &impl.EnvConfig{}
	err := env.Parse(cfg)
	if err != nil {
		panic(err)
	}
	return cfg
}
