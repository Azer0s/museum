package impl

import (
	"errors"
	"museum/domain"
	service "museum/service/interface"
	"museum/util/cache"
	"os/exec"
)

type DockerHostApplicationResolverService struct {
	ExhibitService service.ExhibitService
	IpCache        *cache.LRU[string, string]
}

func (d DockerHostApplicationResolverService) ResolveApplication(exhibitId string) (string, error) {
	exhibit, err := d.ExhibitService.GetExhibitById(exhibitId)
	if err != nil {
		return "", err
	}

	if exhibit.RuntimeInfo.Status != domain.Running {
		return "", errors.New("exhibit is not running")
	}

	if ip, ok := d.IpCache.Get(exhibit.RuntimeInfo.Hostname); ok {
		return ip, nil
	}

	// TODO: refactor this to use the docker API instead of shelling out
	cmd := "docker inspect -f '{{range .NetworkSettings.Networks}}{{.IPAddress}}{{end}}' " + exhibit.RuntimeInfo.Hostname
	ip, err := exec.Command("bash", "-c", cmd).Output()
	if err != nil {
		return "", err
	}

	d.IpCache.Put(exhibit.RuntimeInfo.Hostname, string(ip))

	return string(ip), nil
}
