package service

import (
	"museum/config"
	"museum/service/impl"
	service "museum/service/interface"
)

type EnvironmentTemplateResolverService service.EnvironmentTemplateResolverService

func NewEnvironmentTemplateResolverService(config config.Config) EnvironmentTemplateResolverService {
	return &impl.EnvironmentTemplateResolverServiceImpl{
		Config: config,
	}
}
