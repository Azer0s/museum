package service

import (
	"museum/persistence"
	"museum/service/impl"
	service "museum/service/interface"
	"museum/util/cache"
)

type ApplicationResolverService service.ApplicationResolverService

func NewDockerHostApplicationResolverService(state persistence.State) ApplicationResolverService {
	return &impl.DockerHostApplicationResolverService{
		State:   state,
		IpCache: cache.NewLRU[string, string](1000),
	}
}
