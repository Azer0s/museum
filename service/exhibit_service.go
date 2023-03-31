package service

import (
	"museum/domain"
	"museum/persistence"
	"museum/service/impl"
)

type ExhibitService interface {
	GetExhibits() []domain.Exhibit
}

func NewExhibitServiceImpl(state persistence.SharedPersistentEmittedState) *impl.ExhibitServiceImpl {
	return &impl.ExhibitServiceImpl{
		State: state,
	}
}
