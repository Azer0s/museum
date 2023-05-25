package service

import (
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"museum/persistence"
	"museum/service/impl"
	service "museum/service/interface"
)

type ExhibitService service.ExhibitService

func NewExhibitService(state persistence.State, lockService service.LockService, provider trace.TracerProvider, log *zap.SugaredLogger) ExhibitService {
	return &impl.ExhibitServiceImpl{
		State:       state,
		Provider:    provider,
		LockService: lockService,
		Log:         log,
	}
}
