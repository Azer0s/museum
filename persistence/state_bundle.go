package persistence

import (
	"errors"
	"go.uber.org/zap"
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
	Log                   *zap.SugaredLogger
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

func (s StateBundle) AddExhibit(app domain.CreateExhibit) error {
	s.CurrentStateMutex.Lock()
	defer s.CurrentStateMutex.Unlock()

	return s.SharedPersistentState.WithLock(func() error {
		createEvent, err := domain.NewCreateEvent(app.Exhibit)
		if err != nil {
			s.Log.Debugw("failed to create event", "error", err, "requestId", app.RequestID)
			return err
		}

		received, err := s.EventReceived(createEvent.ID())
		if err != nil {
			s.Log.Debugw("failed to create event receiver", "error", err, "requestId", app.RequestID)
			return err
		}

		err = s.Emitter.EmitEvent(createEvent)
		if err != nil {
			s.Log.Debugw("failed to emit event", "error", err, "requestId", app.RequestID)
			return err
		}

		<-received

		err = s.SharedPersistentState.AddExhibit(app.Exhibit)
		if err != nil {
			s.Log.Debugw("failed to add exhibit to persistent state", "error", err, "requestId", app.RequestID)
			return err
		}

		s.CurrentState = append(s.CurrentState, app.Exhibit)

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
