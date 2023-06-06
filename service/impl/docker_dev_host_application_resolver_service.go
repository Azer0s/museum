package impl

import (
	"context"
	"errors"
	"museum/domain"
	service "museum/service/interface"
)

type DockerDevHostApplicationResolverService struct {
	ExhibitService service.ExhibitService
}

func (d DockerDevHostApplicationResolverService) ResolveApplication(ctx context.Context, exhibitId string) (string, error) {
	exhibit, err := d.ExhibitService.GetExhibitById(ctx, exhibitId)
	if err != nil {
		return "", err
	}

	if exhibit.RuntimeInfo.Status != domain.Running {
		return "", errors.New("exhibit is not running")
	}

	var hostContainer *domain.Object
	for _, object := range exhibit.Objects {
		if object.Name == exhibit.Expose {
			hostContainer = &object
		}
	}

	return exhibit.Name + "_" + exhibit.Expose + "/" + *hostContainer.Port, nil
}

func (d DockerDevHostApplicationResolverService) ResolveExhibitObject(exhibit domain.Exhibit, object domain.Object) (string, error) {
	if exhibit.RuntimeInfo.Status != domain.Running {
		return "", errors.New("exhibit is not running")
	}

	return exhibit.Name + "_" + object.Name + "/" + *object.Port, nil
}
