package service

import (
	"go.uber.org/zap"
	"museum/config"
	"museum/service/impl"
	service "museum/service/interface"
)

type ApplicationProxyService service.ApplicationProxyService

func NewDockerApplicationProxyService(resolver service.ApplicationResolverService, rewriteService service.RewriteService, log *zap.SugaredLogger, config config.Config) ApplicationProxyService {
	return &impl.DockerApplicationProxyService{
		Resolver:       resolver,
		RewriteService: rewriteService,
		Log:            log,
		Config:         config,
	}
}
