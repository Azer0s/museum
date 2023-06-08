package service

import (
	docker "github.com/docker/docker/client"
	"go.uber.org/zap"
	"museum/config"
	"museum/observability"
	"museum/service/impl"
	service "museum/service/interface"
)

type ApplicationProvisionerService service.ApplicationProvisionerService

func NewDockerApplicationProvisionerService(client *docker.Client, exhibitService service.ExhibitService, environmentTemplateResolver service.EnvironmentTemplateResolverService, runtimeInfoService service.RuntimeInfoService, lastAccessedService service.LastAccessedService, lockService service.LockService, livecheckFactoryService LivecheckFactoryService, log *zap.SugaredLogger, providerFactory *observability.TracerProviderFactory, config config.Config) ApplicationProvisionerService {
	return &impl.DockerApplicationProvisionerService{
		ExhibitService:              exhibitService,
		LivecheckFactoryService:     livecheckFactoryService,
		EnvironmentTemplateResolver: environmentTemplateResolver,
		Client:                      client,
		LockService:                 lockService,
		RuntimeInfoService:          runtimeInfoService,
		LastAccessedService:         lastAccessedService,
		Log:                         log,
		Provider:                    providerFactory.Build("docker"),
		Config:                      config,
	}
}
