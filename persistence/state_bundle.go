package persistence

import (
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

func (s StateBundle) RenewExhibitLease(app domain.Exhibit) error {
	//TODO implement me
	panic("implement me")
}

func (s StateBundle) ExpireExhibitLease(app domain.Exhibit) error {
	//TODO implement me
	panic("implement me")
}
