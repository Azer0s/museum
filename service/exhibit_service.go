package service

import (
	docker "github.com/docker/docker/client"
	"go.uber.org/zap"
	"museum/observability"
	"museum/persistence"
	"museum/service/impl"
	service "museum/service/interface"
)

type ExhibitService service.ExhibitService

func NewExhibitService(state persistence.State, eventing persistence.Eventing, infoService service.RuntimeInfoService, lockService service.LockService, factory *observability.TracerProviderFactory, log *zap.SugaredLogger, dockerClient *docker.Client) ExhibitService {
	return &impl.ExhibitServiceImpl{
		State:              state,
		Eventing:           eventing,
		RuntimeInfoService: infoService,
		Provider:           factory.Build("exhibit-service"),
		LockService:        lockService,
		Log:                log,
		DockerClient:       dockerClient,
	}
}
