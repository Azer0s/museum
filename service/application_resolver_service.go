package service

import (
	docker "github.com/docker/docker/client"
	"museum/service/impl"
	service "museum/service/interface"
	"museum/util/cache"
)

type ApplicationResolverService service.ApplicationResolverService

func NewDockerHostApplicationResolverService(exhibitService service.ExhibitService, client *docker.Client) ApplicationResolverService {
	return &impl.DockerHostApplicationResolverService{
		ExhibitService: exhibitService,
		IpCache:        cache.NewLRU[string, string](1000),
		Client:         client,
	}
}
