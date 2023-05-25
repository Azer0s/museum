package impl

import (
	"context"
	"errors"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"museum/domain"
	"museum/persistence"
	service "museum/service/interface"
	"museum/util"
)

type ExhibitServiceImpl struct {
	State       persistence.State
	Provider    trace.TracerProvider
	LockService service.LockService
	Log         *zap.SugaredLogger
}

func (e ExhibitServiceImpl) GetExhibitById(ctx context.Context, id string) (domain.Exhibit, error) {
	globalLock := e.LockService.GetRwLock(ctx, "all", "exhibits")
	err := globalLock.RLock()
	if err != nil {
		e.Log.Errorw("error locking global lock", "error", err)
		return domain.Exhibit{}, err
	}
	defer func(globalLock util.RwErrMutex) {
		err := globalLock.RUnlock()
		if err != nil {
			e.Log.Errorw("error unlocking global lock", "error", err)
		}
	}(globalLock)

	lock := e.LockService.GetRwLock(ctx, id, "exhibit")
	err = lock.RLock()
	if err != nil {
		e.Log.Errorw("error locking exhibit lock", "error", err, "exhibitId", id)
		return domain.Exhibit{}, err
	}
	defer func(lock util.RwErrMutex) {
		err := lock.RUnlock()
		if err != nil {
			e.Log.Errorw("error unlocking exhibit lock", "error", err, "exhibitId", id)
		}
	}(lock)

	exhibit, err := e.State.GetExhibitById(ctx, id)
	if err != nil {
		return domain.Exhibit{}, err
	}

	d, err := e.hydrateExhibit(ctx, id, &exhibit)
	if err != nil {
		return d, err
	}

	return exhibit, nil
}

func (e ExhibitServiceImpl) hydrateExhibit(ctx context.Context, id string, exhibit *domain.Exhibit) (domain.Exhibit, error) {
	//get runtime info
	runtimeInfo, err := e.State.GetRuntimeInfo(ctx, id)
	if err != nil {
		return domain.Exhibit{}, err
	}
	exhibit.RuntimeInfo = &runtimeInfo

	//get last accessed
	lastAccessed, err := e.State.GetLastAccessed(ctx, id)
	if err != nil {
		return domain.Exhibit{}, err
	}
	exhibit.RuntimeInfo.LastAccessed = lastAccessed

	return domain.Exhibit{}, nil
}

func (e ExhibitServiceImpl) GetAllExhibits(ctx context.Context) []domain.Exhibit {
	lock := e.LockService.GetRwLock(ctx, "all", "exhibits")
	err := lock.RLock()
	if err != nil {
		e.Log.Errorw("error locking global lock", "error", err)
		return nil
	}
	defer func(lock util.RwErrMutex) {
		err = lock.RUnlock()
		if err != nil {
			e.Log.Errorw("error unlocking global lock", "error", err)
		}
	}(lock)

	exhibits := e.State.GetAllExhibits(ctx)

	for i, exhibit := range exhibits {
		d, err := e.hydrateExhibit(ctx, exhibit.Id, &exhibit)
		if err != nil {
			continue
		}
		exhibits[i] = d
	}

	return exhibits
}

func (e ExhibitServiceImpl) DeleteExhibitById(ctx context.Context, id string) error {
	lock := e.LockService.GetRwLock(ctx, id, "exhibit")
	err := lock.Lock()
	if err != nil {
		e.Log.Errorw("error locking exhibit lock", "error", err, "exhibitId", id)
		return err
	}
	defer func() {
		err := lock.Unlock()
		if err != nil {
			e.Log.Errorw("error unlocking exhibit lock", "error", err, "exhibitId", id)
		}
		// TODO: delete lock
	}()

	//stop exhibit if running

	return e.State.DeleteExhibitById(ctx, id)
}

func (e ExhibitServiceImpl) CreateExhibit(ctx context.Context, createExhibitRequest domain.CreateExhibit) (string, error) {
	globalLock := e.LockService.GetRwLock(ctx, "all", "exhibits")
	err := globalLock.Lock()
	if err != nil {
		e.Log.Errorw("error locking global lock", "error", err)
		return "", err
	}
	defer func(globalLock util.RwErrMutex) {
		err := globalLock.Unlock()
		if err != nil {
			e.Log.Errorw("error unlocking global lock", "error", err)
		}
	}(globalLock)

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

	err = e.State.SetLastAccessed(subCtx, createExhibitRequest.Exhibit.Id, 0)
	if err != nil {
		return "", err
	}

	err = e.State.SetRuntimeInfo(subCtx, createExhibitRequest.Exhibit.Id, *createExhibitRequest.Exhibit.RuntimeInfo)
	if err != nil {
		err = e.State.DeleteLastAccessed(subCtx, createExhibitRequest.Exhibit.Id)
		return "", err
	}

	err = e.State.CreateExhibit(subCtx, createExhibitRequest.Exhibit)
	if err != nil {
		err = e.State.DeleteLastAccessed(subCtx, createExhibitRequest.Exhibit.Id)
		err = e.State.DeleteRuntimeInfo(subCtx, createExhibitRequest.Exhibit.Id)
		return "", err
	}

	return createExhibitRequest.Exhibit.Id, nil
}
