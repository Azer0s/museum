package impl

import (
	"context"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"museum/domain"
	"museum/persistence"
)

type ExhibitServiceImpl struct {
	State    persistence.SharedPersistentEmittedState
	Provider trace.TracerProvider
}

func (e ExhibitServiceImpl) GetExhibits() []domain.Exhibit {
	return e.State.GetExhibits()
}

func (e ExhibitServiceImpl) GetExhibitById(id string) (*domain.Exhibit, error) {
	return e.State.GetExhibitById(id)
}

func (e ExhibitServiceImpl) CreateExhibit(ctx context.Context, createExhibitRequest domain.CreateExhibit) (error, string) {
	// TODO: validate exhibit

	// give exhibit a unique id
	createExhibitRequest.Exhibit.Id = uuid.New().String()

	// set runtime state
	createExhibitRequest.Exhibit.RuntimeInfo = domain.ExhibitRuntimeInfo{
		Status:            domain.NotCreated,
		RelatedContainers: []string{},
	}

	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("Orchestrate new exhibit").
		Start(ctx, "handleCreateExhibit")
	defer span.End()

	err := e.State.AddExhibit(subCtx, createExhibitRequest)
	if err != nil {
		return err, ""
	}

	return nil, createExhibitRequest.Exhibit.Id
}
