package impl

import (
	"museum/domain"
	"museum/persistence"
)

type ExhibitServiceImpl struct {
	State persistence.SharedPersistentEmittedState
}

func (e ExhibitServiceImpl) GetExhibits() []domain.Exhibit {
	return e.State.GetExhibits()
}
