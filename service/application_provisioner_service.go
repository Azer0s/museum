package service

import (
	docker "github.com/docker/docker/client"
	"museum/service/impl"
	service "museum/service/interface"
)

type ApplicationProvisionerService service.ApplicationProvisionerService

func NewDockerApplicationProvisionerService(exhibitService service.ExhibitService, client *docker.Client) ApplicationProvisionerService {
	return &impl.DockerApplicationProvisionerService{
		ExhibitService: exhibitService,
		Client:         client,
	}
}
