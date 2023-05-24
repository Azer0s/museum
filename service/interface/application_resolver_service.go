package service

import (
	"context"
	"museum/domain"
)

type ApplicationResolverService interface {
	ResolveApplication(ctx context.Context, exhibitId string) (string, error)
	ResolveExhibitObject(exhibit domain.Exhibit, object domain.Object) (string, error)
}
