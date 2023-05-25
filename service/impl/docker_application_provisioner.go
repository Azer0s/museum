package impl

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
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

		//check if container already exists
		inspect, err := d.Client.ContainerInspect(ctx, name)
		if err == nil && inspect.ID != "" {
			d.Log.Warnw("container already exists", "container", name, "exhibitId", exhibit.Id)

			// container already exists, check if it's running
			if inspect.State.Running {
				d.Log.Warnw("container is running, stopping", "container", name, "exhibitId", exhibit.Id)

				// stop container
				err = d.Client.ContainerStop(ctx, inspect.ID, container.StopOptions{})
				if err != nil {
					return err
				}
			}

			d.Log.Debugw("removing container", "container", name, "exhibitId", exhibit.Id)

			// remove container
			err = d.Client.ContainerRemove(ctx, inspect.ID, types.ContainerRemoveOptions{})
			if err != nil {
				return err
			}
		}

		create, err := d.Client.ContainerCreate(ctx, containerConfig, nil, nil, nil, name)
		if err != nil {
			//TODO: rollback
			return err
		}

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
	exhibit, err := d.ExhibitService.GetExhibitById(ctx, exhibitId)
	if err != nil {
		return err
	}

	lock := d.LockService.GetRwLock(ctx, exhibitId, "runtime_info")
	err = lock.Lock()
	if err != nil {
		return err
	}

	defer func(lock util.RwErrMutex) {
		err = lock.Unlock()
	}(lock)

	// check that exhibit is not already started after lock is acquired
	if exhibit.RuntimeInfo.Status == domain.Running {
		return nil
	}

	if exhibit.RuntimeInfo.Status != domain.Stopped && exhibit.RuntimeInfo.Status != domain.NotCreated {
		return errors.New(string("cannot start application in state " + exhibit.RuntimeInfo.Status))
	}

	exhibit.RuntimeInfo.Status = domain.Starting
	exhibit.RuntimeInfo.RelatedContainers = make([]string, 0)

	err = d.RuntimeInfoService.SetRuntimeInfo(ctx, exhibitId, *exhibit.RuntimeInfo)
	if err != nil {
		return err
	}

	return nil
}

func (d DockerApplicationProvisionerService) applicationRunningStep(ctx context.Context, exhibitId string) (err error) {
	exhibitRlock := d.LockService.GetRwLock(ctx, exhibitId, "exhibit")
	err = exhibitRlock.RLock()
	if err != nil {
		//TODO: log
	}
	defer func(lock util.RwErrMutex) {
		err = lock.RUnlock()
	}(exhibitRlock)

	exhibit, err := d.ExhibitService.GetExhibitById(ctx, exhibitId)
	if err != nil {
		return err
	}

	lock := d.LockService.GetRwLock(ctx, exhibitId, "runtime_info")
	err = lock.Lock()
	if err != nil {
		return err
	}

	defer func(lock util.RwErrMutex) {
		err = lock.Unlock()
	}(lock)

	err = d.startApplicationInsideLock(ctx, &exhibit)
	if err != nil {
		exhibit.RuntimeInfo.Status = domain.Stopped
		exhibit.RuntimeInfo.RelatedContainers = make([]string, 0)
		_ = d.RuntimeInfoService.SetRuntimeInfo(ctx, exhibitId, *exhibit.RuntimeInfo)
		return err
	}

	return d.RuntimeInfoService.SetRuntimeInfo(ctx, exhibitId, *exhibit.RuntimeInfo)
}

func (d DockerApplicationProvisionerService) StartApplication(ctx context.Context, exhibitId string) error {
	//TODO: log and span
	err := d.applicationStartingStep(ctx, exhibitId)
	if err != nil {
		return err
	}

	return d.applicationRunningStep(ctx, exhibitId)
}

func (d DockerApplicationProvisionerService) StopApplication(ctx context.Context, exhibitId string) error {
	//TODO implement me
	panic("implement me")
}

func (d DockerApplicationProvisionerService) CleanupApplication(ctx context.Context, exhibitId string) error {
	//TODO implement me
	panic("implement me")
}
