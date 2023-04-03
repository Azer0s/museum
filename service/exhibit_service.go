package service

import (
	"context"
	"go.opentelemetry.io/otel/trace"
	"museum/domain"
	"museum/persistence"
	"museum/service/impl"
)

type ExhibitService interface {
	GetExhibits() []domain.Exhibit
	GetExhibitById(id string) (*domain.Exhibit, error)
	CreateExhibit(ctx context.Context, createExhibit domain.CreateExhibit) (error, string)
}

func NewExhibitServiceImpl(state persistence.SharedPersistentEmittedState, provider trace.TracerProvider) *impl.ExhibitServiceImpl {
	return &impl.ExhibitServiceImpl{
		State:    state,
		Provider: provider,
	}
}
