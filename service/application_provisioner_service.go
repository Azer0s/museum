package service

import (
	docker "github.com/docker/docker/client"
	"museum/persistence"
	"museum/service/impl"
	service "museum/service/interface"
)

type ApplicationProvisionerService service.ApplicationProvisionerService

func NewDockerApplicationProvisionerService(client *docker.Client, state persistence.State, livecheckFactoryService LivecheckFactoryService) ApplicationProvisionerService {
	return &impl.DockerApplicationProvisionerService{
		State:                   state,
		LivecheckFactoryService: livecheckFactoryService,
		Client:                  client,
	}
}
