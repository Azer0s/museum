package impl

import (
	"context"
	"errors"
	docker "github.com/docker/docker/client"
	"museum/domain"
	service "museum/service/interface"
	"museum/util/cache"
)

type DockerHostApplicationResolverService struct {
	ExhibitService service.ExhibitService
	IpCache        *cache.LRU[string, string]
	Client         *docker.Client
}

func (d DockerHostApplicationResolverService) ResolveApplication(ctx context.Context, exhibitId string) (string, error) {
	exhibit, err := d.ExhibitService.GetExhibitById(ctx, exhibitId)
	if err != nil {
		return "", err
	}

	if exhibit.RuntimeInfo.Status != domain.Running {
		return "", errors.New("exhibit is not running")
	}

	if ip, ok := d.IpCache.Get(exhibit.RuntimeInfo.Hostname); ok {
		return ip, nil
	}

	var hostContainer *domain.Object
	for _, object := range exhibit.Objects {
		if object.Name == exhibit.Expose {
			hostContainer = &object
		}
	}

	if hostContainer == nil {
		return "", errors.New("exhibit does not have an expose container")
	}

	ipStr, err := d.ResolveExhibitObject(exhibit, *hostContainer)
	if err != nil {
		return "", err
	}

	d.IpCache.Put(exhibit.RuntimeInfo.Hostname, ipStr)
	return ipStr, nil
}

func (d DockerHostApplicationResolverService) ResolveExhibitObject(exhibit domain.Exhibit, object domain.Object) (string, error) {
	if exhibit.RuntimeInfo.Status != domain.Running {
		return "", errors.New("exhibit is not running")
	}

	objectContainerName := exhibit.Name + "_" + object.Name

	inspect, err := d.Client.ContainerInspect(context.Background(), objectContainerName)
	if err != nil {
		return "", err
	}

	return inspect.NetworkSettings.DefaultNetworkSettings.IPAddress, nil
}
