package service

import (
	"go.uber.org/zap"
	"museum/config"
	"museum/service/impl"
	service "museum/service/interface"
)

type RewriteService service.RewriteService

func NewRewriteService(config config.Config, log *zap.SugaredLogger) RewriteService {
	return &impl.RewriteServiceImpl{
		Config: config,
		Log:    log,
	}
}
