package service

import (
	"go.opentelemetry.io/otel/trace"
	"museum/persistence"
	"museum/service/impl"
	service "museum/service/interface"
)

type ExhibitService service.ExhibitService

func NewExhibitServiceImpl(state persistence.SharedPersistentEmittedState, provider trace.TracerProvider) ExhibitService {
	return &impl.ExhibitServiceImpl{
		State:    state,
		Provider: provider,
	}
}
