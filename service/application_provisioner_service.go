package service

import (
	docker "github.com/docker/docker/client"
	"museum/service/impl"
	service "museum/service/interface"
)

type ApplicationProvisionerService service.ApplicationProvisionerService

func NewDockerApplicationProvisionerService(client *docker.Client, exhibitService service.ExhibitService, livecheckFactoryService LivecheckFactoryService) ApplicationProvisionerService {
	return &impl.DockerApplicationProvisionerService{
		ExhibitService:          exhibitService,
		LivecheckFactoryService: livecheckFactoryService,
		Client:                  client,
	}
}
