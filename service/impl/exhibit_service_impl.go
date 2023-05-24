package impl

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"museum/domain"
	"museum/persistence"
)

type ExhibitServiceImpl struct {
	State    persistence.State
	Provider trace.TracerProvider
}

func (e ExhibitServiceImpl) WithLock(ctx context.Context, id string, f func() error) (err error) {
	return e.State.WithLock(ctx, id, f)
}

func (e ExhibitServiceImpl) GetExhibitById(ctx context.Context, id string) (domain.Exhibit, error) {
	exhibit, err := e.State.GetExhibitById(ctx, id)
	if err != nil {
		return domain.Exhibit{}, err
	}

	//get last accessed
	lastAccessed, err := e.State.GetLastAccessed(ctx, id)
	if err != nil {
		return domain.Exhibit{}, err
	}

	exhibit.RuntimeInfo.LastAccessed = lastAccessed

	return exhibit, nil
}

func (e ExhibitServiceImpl) GetAllExhibits(ctx context.Context) []domain.Exhibit {
	exhibits := e.State.GetAllExhibits(ctx)

	//get last accessed
	for i, exhibit := range exhibits {
		lastAccessed, err := e.State.GetLastAccessed(ctx, exhibit.Id)
		if err != nil {
			continue
		}
		exhibits[i].RuntimeInfo.LastAccessed = lastAccessed
	}

	return exhibits
}

func (e ExhibitServiceImpl) UpdateExhibit(ctx context.Context, app domain.Exhibit) error {
	return e.State.UpdateExhibit(ctx, app)
}

func (e ExhibitServiceImpl) DeleteExhibitById(ctx context.Context, id string) error {
	return e.State.DeleteExhibitById(ctx, id)
}

func (e ExhibitServiceImpl) GetLastAccessed(ctx context.Context, id string) (int64, error) {
	return e.State.GetLastAccessed(ctx, id)
}

func (e ExhibitServiceImpl) SetLastAccessed(ctx context.Context, id string, lastAccessed int64) error {
	return e.State.SetLastAccessed(ctx, id, lastAccessed)
}

func (e ExhibitServiceImpl) CreateExhibit(ctx context.Context, createExhibitRequest domain.CreateExhibit) (string, error) {
	// check that exhibit name is unique
	exhibits := e.State.GetAllExhibits(ctx)
	for _, e := range exhibits {
		if e.Name == createExhibitRequest.Exhibit.Name {
			return "", errors.New("exhibit with name " + createExhibitRequest.Exhibit.Name + " already exists")
		}
	}

	// check that a container is exposed
	if createExhibitRequest.Exhibit.Expose == "" {
		return "", errors.New("exhibit must expose a container")
	}

	// check that exposed container has an exposed port
	found := false
	for _, c := range createExhibitRequest.Exhibit.Objects {
		if c.Name == createExhibitRequest.Exhibit.Expose {
			if c.Port != nil && *c.Port == "" {
				return "", errors.New("exhibit must expose a container with an exposed port")
			}
			found = true
		}
	}

	if !found {
		return "", errors.New("exhibit must expose a container that is part of the exhibit")
	}

	// TODO: validate mount paths

	// give exhibit a unique id
	createExhibitRequest.Exhibit.Id = uuid.New().String()

	// set runtime state
	createExhibitRequest.Exhibit.RuntimeInfo = &domain.ExhibitRuntimeInfo{
		Status:            domain.NotCreated,
		RelatedContainers: []string{},
	}

	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("Orchestrate new exhibit").
		Start(ctx, "handleCreateExhibit")
	defer span.End()

	err := e.State.SetLastAccessed(subCtx, createExhibitRequest.Exhibit.Id, 0)
	if err != nil {
		return "", err
	}

	err = e.State.CreateExhibit(subCtx, createExhibitRequest.Exhibit)
	if err != nil {
		return "", err
	}

	return createExhibitRequest.Exhibit.Id, nil
}
