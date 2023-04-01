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
	ConfirmEvents         map[string]chan struct{}
	ConfirmEventsMutex    *sync.RWMutex
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
	s.CurrentStateMutex.Lock()
	defer s.CurrentStateMutex.Unlock()

	return s.SharedPersistentState.WithLock(func() error {
		createEvent, err := domain.NewCreateEvent(app)
		if err != nil {
			return err
		}

		received, err := s.EventReceived(createEvent.ID())
		if err != nil {
			return err
		}

		err = s.Emitter.EmitEvent(createEvent)
		if err != nil {
			return err
		}

		<-received

		err = s.SharedPersistentState.AddExhibit(app)
		if err != nil {
			return err
		}

		s.CurrentState = append(s.CurrentState, app)

		return nil
	})
}

func (s StateBundle) EventReceived(eventId string) (<-chan struct{}, error) {
	s.ConfirmEventsMutex.Lock()
	defer s.ConfirmEventsMutex.Unlock()

	if _, ok := s.ConfirmEvents[eventId]; ok {
		return nil, errors.New("event receiver already exists")
	}

	s.ConfirmEvents[eventId] = make(chan struct{})
	return s.ConfirmEvents[eventId], nil
}
