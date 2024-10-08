package impl

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types/image"
	docker "github.com/docker/docker/client"
	"github.com/google/uuid"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"io"
	"museum/domain"
	"museum/persistence"
	service "museum/service/interface"
	"museum/util"
	"strconv"
	"time"
)

type ExhibitServiceImpl struct {
	State                    persistence.State
	Eventing                 persistence.Eventing
	RuntimeInfoService       service.RuntimeInfoService
	Provider                 trace.TracerProvider
	LockService              service.LockService
	Log                      *zap.SugaredLogger
	DockerClient             *docker.Client
	VolumeProvisionerFactory service.VolumeProvisionerFactoryService
}

func (e ExhibitServiceImpl) GetExhibitById(ctx context.Context, id string) (domain.Exhibit, error) {
	globalLock := e.LockService.GetRwLock(ctx, "all", "exhibits")
	err := globalLock.RLock()
	if err != nil {
		e.Log.Errorw("error locking global lock", "error", err)
		return domain.Exhibit{}, err
	}

	//TODO: span

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

	err = e.hydrateExhibit(ctx, id, &exhibit)
	if err != nil {
		return domain.Exhibit{}, err
	}

	return exhibit, nil
}

func (e ExhibitServiceImpl) hydrateExhibit(ctx context.Context, id string, exhibit *domain.Exhibit) error {
	subCtx, span := e.Provider.
		Tracer("exhibit-service").
		Start(ctx, "hydrateExhibit("+id+")", trace.WithAttributes(attribute.String("exhibitId", id)))
	defer span.End()

	span.AddEvent("getting runtime info")
	//get runtime info
	runtimeInfo, err := e.RuntimeInfoService.GetRuntimeInfo(subCtx, id)
	if err != nil {
		return err
	}
	exhibit.RuntimeInfo = &runtimeInfo

	span.AddEvent("getting last accessed")
	//get last accessed
	lastAccessed, err := e.State.GetLastAccessed(subCtx, id)
	if err != nil {
		return err
	}
	exhibit.RuntimeInfo.LastAccessed = lastAccessed

	return nil
}

func (e ExhibitServiceImpl) GetAllExhibits(ctx context.Context) []domain.Exhibit {
	lock := e.LockService.GetRwLock(ctx, "all", "exhibits")
	err := lock.RLock()
	if err != nil {
		e.Log.Errorw("error locking global lock", "error", err)
		return nil
	}

	//TODO: span

	defer func(lock util.RwErrMutex) {
		err = lock.RUnlock()
		if err != nil {
			e.Log.Errorw("error unlocking global lock", "error", err)
		}
	}(lock)

	exhibits := e.State.GetAllExhibits(ctx)

	for i, exhibit := range exhibits {
		err = e.hydrateExhibit(ctx, exhibit.Id, &exhibit)
		if err != nil {
			continue
		}
		exhibits[i] = exhibit
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

	//TODO: span

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
	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("exhibit-service").
		Start(ctx, "CreateExhibit")
	defer span.End()

	e.Log.Infow("creating new exhibit", "exhibitId", createExhibitRequest.Exhibit.Id)

	globalLock := e.LockService.GetRwLock(subCtx, "all", "exhibits")
	err := globalLock.Lock()
	if err != nil {
		e.Log.Errorw("error locking global lock", "error", err)
		return "", err
	}

	//TODO: check container address replacement in ENV

	defer func(globalLock util.RwErrMutex) {
		err := globalLock.Unlock()
		if err != nil {
			e.Log.Errorw("error unlocking global lock", "error", err)
		}
	}(globalLock)

	// check that exhibit name is unique
	exhibits := e.State.GetAllExhibits(subCtx)
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
	{
		found := false
		for i, c := range createExhibitRequest.Exhibit.Objects {
			if c.Name == createExhibitRequest.Exhibit.Expose {
				if c.Port == nil || *c.Port == "" {
					createExhibitRequest.Exhibit.Objects[i].Port = new(string)
					*createExhibitRequest.Exhibit.Objects[i].Port = "80"
				}
				found = true
			}
		}

		if !found {
			return "", errors.New("exhibit must expose a container that is part of the exhibit")
		}
	}

	//---------------------------------------------------

	// validate mount paths
	mounts := make([]string, 0)
	for _, c := range createExhibitRequest.Exhibit.Objects {
		if len(c.Mounts) != 0 {
			for mount := range c.Mounts {
				mounts = append(mounts, mount)
			}
		}
	}

	// check that volumes have required mounts
	for _, m := range mounts {
		found := false
		for _, v := range createExhibitRequest.Exhibit.Volumes {
			if v.Name == m {
				found = true
			}
		}

		if !found {
			return "", errors.New("mount " + m + " does not have a corresponding volume")
		}
	}

	for _, v := range createExhibitRequest.Exhibit.Volumes {
		vp, err := e.VolumeProvisionerFactory.GetForDriverType(v.Driver.Type)
		if err != nil {
			return "", err
		}

		err = vp.CheckValidity(v.Driver.Config)
		if err != nil {
			return "", err
		}
	}

	//---------------------------------------------------

	// validate lease time
	_, err = time.ParseDuration(createExhibitRequest.Exhibit.Lease)
	if err != nil {
		return "", errors.New("lease time must be a valid duration")
	}

	// validate livechecks
	for _, object := range createExhibitRequest.Exhibit.Objects {
		l := object.Livecheck
		if l == nil {
			continue
		}

		// validate livecheck type
		if l.Type != domain.LivecheckTypeHttp && l.Type != domain.LivecheckTypeExec {
			return "", errors.New("livecheck type must be one of: " + domain.LivecheckTypeHttp + ", " + domain.LivecheckTypeExec + " (in object " + object.Name + ")")
		}

		// check http livecheck
		if l.Type == domain.LivecheckTypeHttp {
			// check valid http method
			method, ok := l.Config["method"]
			if ok {
				if method != "GET" && method != "POST" && method != "PUT" && method != "DELETE" {
					return "", errors.New("http livecheck method must be one of: GET, POST, PUT, DELETE (in object " + object.Name + ")")
				}
			}

			// check valid http status
			status, ok := l.Config["status"]
			if ok {
				_, err := strconv.Atoi(status)
				if err != nil {
					return "", errors.New("http livecheck status must be a valid integer (in object " + object.Name + ")")
				}
			}

			// check valid port
			port, ok := l.Config["port"]
			if ok {
				_, err := strconv.Atoi(port)
				if err != nil {
					return "", errors.New("http livecheck port must be a valid integer (in object " + object.Name + ")")
				}
			}
		}
	}

	//---------------------------------------------------

	// give exhibit a unique id
	createExhibitRequest.Exhibit.Id = uuid.New().String()

	// set runtime state
	createExhibitRequest.Exhibit.RuntimeInfo = &domain.ExhibitRuntimeInfo{
		Status:            domain.NotCreated,
		RelatedContainers: []string{},
	}

	//---------------------------------------------------

	e.Log.Infow("pulling images", "exhibitId", createExhibitRequest.Exhibit.Id)
	for _, object := range createExhibitRequest.Exhibit.Objects {
		e.Log.Debugw("pulling image", "image", object.Image+":"+object.Label, "exhibitId", createExhibitRequest.Exhibit.Id)

		inspect, _, err := e.DockerClient.ImageInspectWithRaw(subCtx, object.Image+":"+object.Label)
		if err != nil && !docker.IsErrNotFound(err) {
			e.Log.Errorw("error inspecting image", "image", object.Image+":"+object.Label, "exhibitId", createExhibitRequest.Exhibit.Id, "error", err)
			return "", err
		}

		if inspect.ID != "" {
			e.Log.Debugw("image already pulled", "image", object.Image+":"+object.Label, "exhibitId", createExhibitRequest.Exhibit.Id)
			continue
		}

		containerImage := object.Image + ":" + object.Label
		pull, err := e.DockerClient.ImagePull(ctx, containerImage, image.PullOptions{})
		if err != nil {
			e.Log.Errorw("error pulling image", "image", containerImage, "exhibitId", createExhibitRequest.Exhibit.Id, "error", err)
			return "", err
		}

		_, err = io.ReadAll(pull)
		if err != nil {
			e.Log.Errorw("error reading pull response", "image", containerImage, "exhibitId", createExhibitRequest.Exhibit.Id, "error", err)
			return "", err
		}

		err = pull.Close()
		if err != nil {
			e.Log.Errorw("error closing pull response", "image", containerImage, "exhibitId", createExhibitRequest.Exhibit.Id, "error", err)
			return "", err
		}
	}

	err = e.State.SetLastAccessed(subCtx, createExhibitRequest.Exhibit.Id, time.Now().Unix())
	if err != nil {
		return "", err
	}

	err = e.State.SetRuntimeInfo(subCtx, createExhibitRequest.Exhibit.Id, *createExhibitRequest.Exhibit.RuntimeInfo)
	if err != nil {
		e.Log.Errorw("error setting runtime info, reverting", "error", err, "exhibitId", createExhibitRequest.Exhibit.Id)
		err = e.State.DeleteLastAccessed(subCtx, createExhibitRequest.Exhibit.Id)
		return "", err
	}

	err = e.State.CreateExhibit(subCtx, createExhibitRequest.Exhibit)
	if err != nil {
		e.Log.Errorw("error creating exhibit, reverting", "error", err, "exhibitId", createExhibitRequest.Exhibit.Id)
		err = e.State.DeleteLastAccessed(subCtx, createExhibitRequest.Exhibit.Id)
		err = e.State.DeleteRuntimeInfo(subCtx, createExhibitRequest.Exhibit.Id)
		return "", err
	}

	e.Eventing.DispatchExhibitCreatedEvent(subCtx, createExhibitRequest.Exhibit)
	e.Log.Debugw("created new exhibit", "exhibitId", createExhibitRequest.Exhibit.Id)

	return createExhibitRequest.Exhibit.Id, nil
}

func (e ExhibitServiceImpl) Count() int {
	return len(e.State.GetAllExhibits(context.Background()))
}
