package service

import (
	docker "github.com/docker/docker/client"
	"museum/persistence"
	"museum/service/impl"
	service "museum/service/interface"
)

type ApplicationProvisionerService service.ApplicationProvisionerService

func NewDockerApplicationProvisionerService(exhibitService service.ExhibitService, client *docker.Client, sharedPersistentState persistence.SharedPersistentState) ApplicationProvisionerService {
	return &impl.DockerApplicationProvisionerService{
		SharedPersistentState: sharedPersistentState,
		ExhibitService:        exhibitService,
		Client:                client,
	}
}
