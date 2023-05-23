package service

import "museum/domain"

type ApplicationResolverService interface {
	ResolveApplication(exhibitId string) (string, error)
	ResolveExhibitObject(exhibit domain.Exhibit, object domain.Object) (string, error)
}
