package impl

import (
	"context"
	docker "github.com/docker/docker/client"
	service "museum/service/interface"
)

type DockerApplicationProvisionerService struct {
	ExhibitService service.ExhibitService
	Client         *docker.Client
}

func (d DockerApplicationProvisionerService) StartApplication(ctx context.Context, exhibitId string) error {
	//TODO implement me
	panic("implement me")
}

func (d DockerApplicationProvisionerService) StopApplication(ctx context.Context, exhibitId string) error {
	//TODO implement me
	panic("implement me")
}

func (d DockerApplicationProvisionerService) CleanupApplication(ctx context.Context, exhibitId string) error {
	//TODO implement me
	panic("implement me")
}
