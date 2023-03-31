package impl

import (
	"github.com/google/uuid"
	"museum/domain"
	"museum/persistence"
)

type ExhibitServiceImpl struct {
	State persistence.SharedPersistentEmittedState
}

func (e ExhibitServiceImpl) GetExhibits() []domain.Exhibit {
	return e.State.GetExhibits()
}

func (e ExhibitServiceImpl) GetExhibitById(id string) (*domain.Exhibit, error) {
	return e.State.GetExhibitById(id)
}

func (e ExhibitServiceImpl) CreateExhibit(exhibit domain.Exhibit) error {
	// TODO: validate exhibit

	// give exhibit a unique id
	exhibit.Id = uuid.New().String()

	// set runtime state
	exhibit.RuntimeInfo = domain.ExhibitRuntimeInfo{
		Status:            domain.NotCreated,
		RelatedContainers: []string{},
	}

	return e.State.AddExhibit(exhibit)
}
