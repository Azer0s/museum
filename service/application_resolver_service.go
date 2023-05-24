package service

import (
	"museum/service/impl"
	service "museum/service/interface"
	"museum/util/cache"
)

type ApplicationResolverService service.ApplicationResolverService

func NewDockerHostApplicationResolverService(exhibitService service.ExhibitService) ApplicationResolverService {
	return &impl.DockerHostApplicationResolverService{
		ExhibitService: exhibitService,
		IpCache:        cache.NewLRU[string, string](1000),
	}
}
