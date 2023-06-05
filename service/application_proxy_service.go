package service

import (
	"go.uber.org/zap"
	"museum/service/impl"
	service "museum/service/interface"
)

type ApplicationProxyService service.ApplicationProxyService

func NewDockerApplicationProxyService(resolver service.ApplicationResolverService, log *zap.SugaredLogger) ApplicationProxyService {
	return &impl.DockerApplicationProxyService{
		Resolver: resolver,
		Log:      log,
	}
}
