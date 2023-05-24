package impl

import (
	"context"
	"errors"
	"museum/domain"
	service "museum/service/interface"
	"museum/util/cache"
	"os/exec"
	"strings"
)

type DockerHostApplicationResolverService struct {
	ExhibitService service.ExhibitService
	IpCache        *cache.LRU[string, string]
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

	// TODO: refactor this to use the docker API instead of shelling out
	cmd := "docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' " + objectContainerName
	ip, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", err
	}

	ipStr := strings.ReplaceAll(string(ip), "\n", "")

	return ipStr, nil
}
