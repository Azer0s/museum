package persistence

import (
	"errors"
	"museum/domain"
	"sync"
)

type StateBundle struct {
	SharedPersistentState SharedPersistentState
	Emitter               Emitter
	Consumer              Consumer
	CurrentState          []domain.Exhibit
	CurrentStateMutex     *sync.RWMutex
}

func (s StateBundle) GetExhibitById(id string) (*domain.Exhibit, error) {
	s.CurrentStateMutex.RLock()
	defer s.CurrentStateMutex.RUnlock()
	for _, exhibit := range s.CurrentState {
		if exhibit.Id == id {
			return &exhibit, nil
		}
	}
	return nil, errors.New("exhibit not found")
}

func (s StateBundle) RenewExhibitLeaseById(id string) error {
	//TODO implement me
	panic("implement me")
}

func (s StateBundle) ExpireExhibitLease(id string) error {
	//TODO implement me
	panic("implement me")
}

func (s StateBundle) GetExhibits() []domain.Exhibit {
	// return s.SharedPersistentState.GetExhibits()
	s.CurrentStateMutex.RLock()
	defer s.CurrentStateMutex.RUnlock()
	return s.CurrentState
}

func (s StateBundle) AddExhibit(app domain.Exhibit) error {
	//TODO implement me
	panic("implement me")
}
