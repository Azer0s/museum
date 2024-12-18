package service

import (
	"go.uber.org/zap"
	"museum/config"
	"museum/observability"
	"museum/service/impl"
	service "museum/service/interface"
)

type ExhibitCleanupService service.ExhibitCleanupService

func NewExhibitCleanupService(exhibitService service.ExhibitService, lockService service.LockService, provisionerService service.ApplicationProvisionerService, factory *observability.TracerProviderFactory, log *zap.SugaredLogger, config config.Config) ExhibitCleanupService {
	return &impl.ExhibitCleanupServiceImpl{
		ExhibitService:                exhibitService,
		LockService:                   lockService,
		ApplicationProvisionerService: provisionerService,
		Provider:                      factory.Build("cleanup-service"),
		Log:                           log,
		Config:                        config,
	}
}
