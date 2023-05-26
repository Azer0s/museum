package service

import (
	"go.uber.org/zap"
	"museum/observability"
	"museum/persistence"
	"museum/service/impl"
	service "museum/service/interface"
)

type ExhibitService service.ExhibitService

func NewExhibitService(state persistence.State, lockService service.LockService, factory *observability.TracerProviderFactory, log *zap.SugaredLogger) ExhibitService {
	return &impl.ExhibitServiceImpl{
		State:       state,
		Provider:    factory.Build("exhibit-service"),
		LockService: lockService,
		Log:         log,
	}
}
