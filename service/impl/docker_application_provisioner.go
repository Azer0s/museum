package impl

import (
	"context"
	"errors"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"museum/domain"
	"museum/persistence"
)

type DockerApplicationProvisionerService struct {
	SharedPersistentState        persistence.SharedPersistentState
	SharedPersistentEmittedState persistence.SharedPersistentEmittedState
	Client                       *docker.Client
}

func (d DockerApplicationProvisionerService) startApplicationInsideLock(ctx context.Context, exhibitId string) error {
	exhibit, err := d.SharedPersistentEmittedState.GetExhibitById(exhibitId)
	if err != nil {
		return err
	}

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

		create, err := d.Client.ContainerCreate(ctx, containerConfig, nil, nil, nil, o.Name)
		if err != nil {
			return err
		}

		exhibit.RuntimeInfo.RelatedContainers = append(exhibit.RuntimeInfo.RelatedContainers, create.ID)
		exhibit.RuntimeInfo.Status = domain.Running

		err = d.Client.ContainerStart(ctx, create.ID, types.ContainerStartOptions{})
		if err != nil {
			return err
		}
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

		exhibit.RuntimeInfo.Status = domain.Running
		exhibit.RuntimeInfo.RelatedContainers = make([]string, 0)

		err = d.SharedPersistentEmittedState.StartExhibit(ctx, *exhibit)

		return nil
	})
	if err != nil {
		return err
	}

	return d.SharedPersistentState.WithLock(func() error {
		return d.startApplicationInsideLock(ctx, exhibitId)
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
