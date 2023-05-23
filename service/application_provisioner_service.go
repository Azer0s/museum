package service

import (
	docker "github.com/docker/docker/client"
	"museum/persistence"
	"museum/service/impl"
	service "museum/service/interface"
)

type ApplicationProvisionerService service.ApplicationProvisionerService

func NewDockerApplicationProvisionerService(client *docker.Client, sharedPersistentState persistence.SharedPersistentState, sharedPersistentEmittedState persistence.SharedPersistentEmittedState, livecheckFactoryService LivecheckFactoryService) ApplicationProvisionerService {
	return &impl.DockerApplicationProvisionerService{
		SharedPersistentState:        sharedPersistentState,
		SharedPersistentEmittedState: sharedPersistentEmittedState,
		LivecheckFactoryService:      livecheckFactoryService,
		Client:                       client,
	}
}
