package impl

import (
	docker "github.com/docker/docker/client"
	service "museum/service/interface"
)

type DockerApplicationProvisionerService struct {
	ExhibitService service.ExhibitService
	Client         *docker.Client
}

func (d DockerApplicationProvisionerService) StartApplication(exhibitId string) error {
	//TODO implement me
	panic("implement me")
}

func (d DockerApplicationProvisionerService) StopApplication(exhibitId string) error {
	//TODO implement me
	panic("implement me")
}

func (d DockerApplicationProvisionerService) CleanupApplication(exhibitId string) error {
	//TODO implement me
	panic("implement me")
}
