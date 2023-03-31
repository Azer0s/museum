package tool

import "museum/domain"

type ApiClient interface {
	CreateExhibit(exhibit *domain.Exhibit) error
}

type ApiClientImpl struct {
	BaseUrl string
}

func (a *ApiClientImpl) CreateExhibit(exhibit *domain.Exhibit) error {
	panic("implement me")
}
