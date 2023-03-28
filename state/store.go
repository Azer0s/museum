package state

import (
	"museum/domain"
	"museum/persistence"
	"sync"
)

type Store struct {
	persistentState persistence.SharedPersistentState
	emitter         persistence.Emitter
	consumer        persistence.Consumer
	state           map[string]domain.Application
	stateMu         sync.RWMutex
}

func (s *Store) handleEvents() {
	consumerChan, err := s.consumer.GetEvents()
	if err != nil {
		panic(err)
	}
	for event := range consumerChan {
		switch event.Type() {
		case domain.CreateEventType:
			var app domain.Application
			err := event.DataAs(&app)
			if err != nil {
				// this should never happen
				panic(err)
			}

			s.stateMu.Lock()
			s.state[app.Id] = app
			s.stateMu.Unlock()
		}
	}
}

func NewStore(persistentState persistence.SharedPersistentState, emitter persistence.Emitter, consumer persistence.Consumer) (*Store, error) {
	store := &Store{
		persistentState: persistentState,
		emitter:         emitter,
		consumer:        consumer,
		state:           make(map[string]domain.Application),
		stateMu:         sync.RWMutex{},
	}

	go store.handleEvents()

	apps, err := persistentState.GetApplications()
	if err != nil {
		return nil, err
	}

	store.stateMu.Lock()
	defer store.stateMu.Unlock()

	for _, app := range apps {
		store.state[app.Id] = app
	}

	return store, nil
}
