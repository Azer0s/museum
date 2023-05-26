package service

import (
	"go.uber.org/zap"
	"museum/observability"
	"museum/service/impl"
	service "museum/service/interface"
)

type ExhibitCleanupService service.ExhibitCleanupService

func NewExhibitCleanupService(exhibitService service.ExhibitService, lockService service.LockService, factory *observability.TracerProviderFactory, log *zap.SugaredLogger) ExhibitCleanupService {
	return &impl.ExhibitCleanupServiceImpl{
		ExhibitService: exhibitService,
		LockService:    lockService,
		Provider:       factory.Build("cleanup-service"),
		Log:            log,
	}
}
