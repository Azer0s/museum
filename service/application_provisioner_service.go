package service

import (
	docker "github.com/docker/docker/client"
	"go.uber.org/zap"
	"museum/observability"
	"museum/service/impl"
	service "museum/service/interface"
)

type ApplicationProvisionerService service.ApplicationProvisionerService

func NewDockerApplicationProvisionerService(client *docker.Client, exhibitService service.ExhibitService, runtimeInfoService service.RuntimeInfoService, lockService service.LockService, livecheckFactoryService LivecheckFactoryService, log *zap.SugaredLogger, providerFactory *observability.TracerProviderFactory) ApplicationProvisionerService {
	return &impl.DockerApplicationProvisionerService{
		ExhibitService:          exhibitService,
		LivecheckFactoryService: livecheckFactoryService,
		Client:                  client,
		LockService:             lockService,
		RuntimeInfoService:      runtimeInfoService,
		Log:                     log,
		Provider:                providerFactory.Build("docker"),
	}
}
