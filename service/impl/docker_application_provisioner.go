package impl

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"museum/domain"
	service "museum/service/interface"
	"museum/util"
	"strconv"
	"time"
)

type DockerApplicationProvisionerService struct {
	ExhibitService          service.ExhibitService
	LivecheckFactoryService service.LivecheckFactoryService
	Client                  *docker.Client
	LockService             service.LockService
	RuntimeInfoService      service.RuntimeInfoService
	Log                     *zap.SugaredLogger
	Provider                trace.TracerProvider
}

func (d DockerApplicationProvisionerService) startApplicationInsideLock(ctx context.Context, exhibit *domain.Exhibit) error {
	sortedObjects := exhibit.Objects
	if exhibit.Order != nil {
		sortedObjects = make([]domain.Object, 0)
		for _, s := range exhibit.Order {
			for _, o := range exhibit.Objects {
				if s == o.Name {
					sortedObjects = append(sortedObjects, o)
				}
			}
		}
	}

	// create a container on the swarm for each object
	for _, o := range sortedObjects {
		containerConfig := &container.Config{
			Image: o.Image,
			Env:   make([]string, 0),
		}

		if o.Environment != nil {
			for k, v := range o.Environment {
				containerConfig.Env = append(containerConfig.Env, k+"="+v)
			}
		}

		name := exhibit.Name + "_" + o.Name

		ctx, span := d.Provider.
			Tracer("docker provisioner").
			Start(ctx, "startApplicationInsideLock", trace.WithAttributes(attribute.String("container", name), attribute.String("exhibitId", exhibit.Id)))
		defer span.End()

		span.AddEvent("inspecting container")

		//check if container already exists
		inspect, err := d.Client.ContainerInspect(ctx, name)
		if err == nil && inspect.ID != "" {
			d.Log.Warnw("container already exists", "container", name, "exhibitId", exhibit.Id)

			// container already exists, check if it's running
			if inspect.State.Running {
				span.AddEvent("stopping container")

				d.Log.Warnw("container is running, stopping", "container", name, "exhibitId", exhibit.Id)

				// stop container
				err = d.Client.ContainerStop(ctx, inspect.ID, container.StopOptions{})
				if err != nil {
					return err
				}
			}

			span.AddEvent("removing container")

			d.Log.Debugw("removing container", "container", name, "exhibitId", exhibit.Id)

			// remove container
			err = d.Client.ContainerRemove(ctx, inspect.ID, types.ContainerRemoveOptions{})
			if err != nil {
				return err
			}
		}

		span.AddEvent("creating container")

		create, err := d.Client.ContainerCreate(ctx, containerConfig, nil, nil, nil, name)
		if err != nil {
			//TODO: rollback
			return err
		}

		span.AddEvent("starting container")

		d.Log.Debugw("starting container", "container", name, "exhibitId", exhibit.Id)
		err = d.Client.ContainerStart(ctx, create.ID, types.ContainerStartOptions{})
		if err != nil {
			//TODO: rollback
			d.Log.Errorw("error starting container", "container", name, "exhibitId", exhibit.Id, "error", err)
			return err
		}

		if o.Name == exhibit.Expose {
			exhibit.RuntimeInfo.Hostname = name
		}

		if o.Livecheck != nil {
			span.AddEvent("doing livecheck")

			err := d.doLivecheck(*exhibit, o)
			if err != nil {
				d.Log.Errorw("error doing livecheck", "exhibitId", exhibit.Id, "error", err)
				return err
			}
		}

		exhibit.RuntimeInfo.RelatedContainers = append(exhibit.RuntimeInfo.RelatedContainers, create.ID)
	}

	exhibit.RuntimeInfo.Status = domain.Running

	return nil
}

func (d DockerApplicationProvisionerService) doLivecheck(exhibit domain.Exhibit, object domain.Object) error {
	var err error = nil
	runtimeInfoCopy := *exhibit.RuntimeInfo
	exhibit.RuntimeInfo = &runtimeInfoCopy

	livecheck := d.LivecheckFactoryService.GetLivecheckService(object.Livecheck.Type)
	if livecheck == nil {
		return errors.New("livecheck type not found")
	}

	maxRetries := 10
	if r, ok := object.Livecheck.Config["maxRetries"]; ok {
		maxRetries, err = strconv.Atoi(r)
		if err != nil {
			return err
		}
	}

	interval := 1 * time.Second
	if i, ok := object.Livecheck.Config["interval"]; ok {
		interval, err = time.ParseDuration(i)
		if err != nil {
			return err
		}
	}

	retry := true
	for counter := 0; (retry && err == nil) && counter < maxRetries; counter++ {
		retry, err = livecheck.Check(exhibit, object)
		time.Sleep(interval)
	}
	return nil
}

func (d DockerApplicationProvisionerService) applicationStartingStep(ctx context.Context, exhibitId string) (err error) {
	subCtx, span := d.Provider.
		Tracer("docker provisioner").
		Start(ctx, "applicationStartingStep", trace.WithAttributes(attribute.String("exhibitId", exhibitId)))
	defer span.End()

	exhibit, err := d.ExhibitService.GetExhibitById(subCtx, exhibitId)
	if err != nil {
		return err
	}

	lock := d.LockService.GetRwLock(subCtx, exhibitId, "runtime_info")
	err = lock.Lock()
	if err != nil {
		return err
	}

	defer func(lock util.RwErrMutex) {
		err = lock.Unlock()
	}(lock)

	span.AddEvent("checking exhibit status")

	// check that exhibit is not already started after lock is acquired
	if exhibit.RuntimeInfo.Status == domain.Running {
		return nil
	}

	if exhibit.RuntimeInfo.Status != domain.Stopped && exhibit.RuntimeInfo.Status != domain.NotCreated {
		return errors.New(string("cannot start application in state " + exhibit.RuntimeInfo.Status))
	}

	span.AddEvent("setting exhibit status to starting")

	exhibit.RuntimeInfo.Status = domain.Starting
	exhibit.RuntimeInfo.RelatedContainers = make([]string, 0)

	err = d.RuntimeInfoService.SetRuntimeInfo(subCtx, exhibitId, *exhibit.RuntimeInfo)
	if err != nil {
		return err
	}

	return nil
}

func (d DockerApplicationProvisionerService) applicationRunningStep(ctx context.Context, exhibitId string) (err error) {
	subCtx, span := d.Provider.
		Tracer("docker provisioner").
		Start(ctx, "applicationRunningStep", trace.WithAttributes(attribute.String("exhibitId", exhibitId)))
	defer span.End()

	exhibitRlock := d.LockService.GetRwLock(subCtx, exhibitId, "exhibit")
	err = exhibitRlock.RLock()
	if err != nil {
		d.Log.Errorw("error locking exhibit", "exhibitId", exhibitId, "error", err)
		return err
	}
	defer func(lock util.RwErrMutex) {
		err = lock.RUnlock()
	}(exhibitRlock)

	exhibit, err := d.ExhibitService.GetExhibitById(subCtx, exhibitId)
	if err != nil {
		return err
	}

	lock := d.LockService.GetRwLock(subCtx, exhibitId, "runtime_info")
	err = lock.Lock()
	if err != nil {
		return err
	}

	defer func(lock util.RwErrMutex) {
		err = lock.Unlock()
	}(lock)

	err = d.startApplicationInsideLock(subCtx, &exhibit)
	if err != nil {
		span.AddEvent("error starting application, reverting status to stopped")

		exhibit.RuntimeInfo.Status = domain.Stopped
		exhibit.RuntimeInfo.RelatedContainers = make([]string, 0)
		_ = d.RuntimeInfoService.SetRuntimeInfo(subCtx, exhibitId, *exhibit.RuntimeInfo)
		return err
	}

	return d.RuntimeInfoService.SetRuntimeInfo(subCtx, exhibitId, *exhibit.RuntimeInfo)
}

func (d DockerApplicationProvisionerService) StartApplication(ctx context.Context, exhibitId string) error {
	subCtx, span := d.Provider.
		Tracer("docker provisioner").
		Start(ctx, "StartApplication", trace.WithAttributes(attribute.String("exhibitId", exhibitId)))
	defer span.End()

	err := d.applicationStartingStep(subCtx, exhibitId)
	if err != nil {
		d.Log.Errorw("error starting application", "exhibitId", exhibitId, "error", err)
		return err
	}

	err = d.applicationRunningStep(subCtx, exhibitId)
	if err != nil {
		d.Log.Errorw("error starting application", "exhibitId", exhibitId, "error", err)
		return err
	}

	return nil
}

func (d DockerApplicationProvisionerService) StopApplication(ctx context.Context, exhibitId string) error {
	//TODO implement me
	panic("implement me")
}

func (d DockerApplicationProvisionerService) CleanupApplication(ctx context.Context, exhibitId string) error {
	//TODO implement me
	panic("implement me")
}
