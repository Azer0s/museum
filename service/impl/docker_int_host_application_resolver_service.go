package impl

import (
	"context"
	"errors"
	"museum/domain"
	service "museum/service/interface"
)

type DockerIntHostApplicationResolverService struct {
	ExhibitService service.ExhibitService
}

func (d DockerIntHostApplicationResolverService) ResolveApplication(ctx context.Context, exhibitId string) (string, error) {
	exhibit, err := d.ExhibitService.GetExhibitById(ctx, exhibitId)
	if err != nil {
		return "", err
	}

	if exhibit.RuntimeInfo.Status != domain.Running {
		return "", errors.New("exhibit is not running")
	}

	return exhibit.Name + "_" + exhibit.Expose, nil
}

func (d DockerIntHostApplicationResolverService) ResolveExhibitObject(exhibit domain.Exhibit, object domain.Object) (string, error) {
	if exhibit.RuntimeInfo.Status != domain.Running {
		return "", errors.New("exhibit is not running")
	}

	return exhibit.Name + "_" + object.Name, nil
}
