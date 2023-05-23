package impl

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"museum/domain"
	"museum/persistence"
	service "museum/service/interface"
	"strconv"
	"time"
)

type DockerApplicationProvisionerService struct {
	SharedPersistentState        persistence.SharedPersistentState
	SharedPersistentEmittedState persistence.SharedPersistentEmittedState
	LivecheckFactoryService      service.LivecheckFactoryService
	Client                       *docker.Client
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

		create, err := d.Client.ContainerCreate(ctx, containerConfig, nil, nil, nil, name)
		if err != nil {
			//TODO: rollback
			return err
		}

		//TODO: handle exhibit already started, check for docker status

		err = d.Client.ContainerStart(ctx, create.ID, types.ContainerStartOptions{})
		if err != nil {
			//TODO: rollback
			return err
		}

		if o.Name == exhibit.Expose {
			exhibit.RuntimeInfo.Hostname = name
		}

		if o.Livecheck != nil {
			err := d.doLivecheck(*exhibit, o)
			if err != nil {
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

func (d DockerApplicationProvisionerService) StartApplication(ctx context.Context, exhibitId string) error {
	err := d.SharedPersistentState.WithLock(func() error {
		exhibit, err := d.SharedPersistentEmittedState.GetExhibitById(exhibitId)
		if err != nil {
			return err
		}

		// check that exhibit is not already started after lock is acquired
		if exhibit.RuntimeInfo.Status == domain.Running {
			return nil
		}

		if exhibit.RuntimeInfo.Status != domain.Stopped && exhibit.RuntimeInfo.Status != domain.NotCreated {
			return errors.New(string("cannot start application in state " + exhibit.RuntimeInfo.Status))
		}

		exhibit.RuntimeInfo.Status = domain.Starting
		exhibit.RuntimeInfo.RelatedContainers = make([]string, 0)
		exhibit.RuntimeInfo.LastAccessed = strconv.FormatInt(time.Now().UnixNano(), 10)

		return d.SharedPersistentEmittedState.StartingExhibit(ctx, *exhibit)
	})

	if err != nil {
		//TODO: rollback
		return err
	}

	return d.SharedPersistentState.WithLock(func() error {
		exhibit, err := d.SharedPersistentEmittedState.GetExhibitById(exhibitId)

		err = d.startApplicationInsideLock(ctx, exhibit)
		if err != nil {
			return err
		}

		return d.SharedPersistentEmittedState.StartExhibit(ctx, *exhibit)
	})
}

func (d DockerApplicationProvisionerService) StopApplication(ctx context.Context, exhibitId string) error {
	//TODO implement me
	panic("implement me")
}

func (d DockerApplicationProvisionerService) CleanupApplication(ctx context.Context, exhibitId string) error {
	//TODO implement me
	panic("implement me")
}
