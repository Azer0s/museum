package impl

import (
	"context"
	"fmt"
	docker "github.com/docker/docker/client"
	"museum/domain"
	"museum/persistence"
	service "museum/service/interface"
)

type DockerApplicationProvisionerService struct {
	SharedPersistentState persistence.SharedPersistentState
	ExhibitService        service.ExhibitService
	Client                *docker.Client
}

func (d DockerApplicationProvisionerService) startApplicationInsideLock(ctx context.Context, exhibitId string) error {
	exhibit, err := d.ExhibitService.GetExhibitById(exhibitId)
	if err != nil {
		return err
	}

	sortedObjects := make([]domain.Object, 0)
	for _, s := range exhibit.Order {
		for _, o := range exhibit.Objects {
			if s == o.Name {
				sortedObjects = append(sortedObjects, o)
			}
		}
	}

	fmt.Println(sortedObjects)

	return nil
}

func (d DockerApplicationProvisionerService) StartApplication(ctx context.Context, exhibitId string) error {
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
