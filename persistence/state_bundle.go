package persistence

import (
	"context"
	"errors"
	"go.opentelemetry.io/otel/trace"
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
	Provider              trace.TracerProvider
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

func (s StateBundle) AddExhibit(ctx context.Context, app domain.CreateExhibit) error {
	s.CurrentStateMutex.Lock()
	defer s.CurrentStateMutex.Unlock()

	span := trace.SpanFromContext(ctx)

	return s.SharedPersistentState.WithLock(func() error {
		span.AddEvent("lock acquired")

		// create span for event emission
		eventCtx, eventSpan := s.Provider.
			Tracer("Event emission").
			Start(ctx, "emitCreateEvent")

		spanEnded := false
		defer func() {
			if !spanEnded {
				eventSpan.End()
			}
		}()

		createEvent, err := domain.NewCreateEvent(app.Exhibit)
		if err != nil {
			s.Log.Debugw("failed to create event", "error", err, "requestId", app.RequestID)
			return err
		}

		eventSpan.AddEvent("event created")

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

		eventSpan.AddEvent("event emitted")

		<-received

		eventSpan.AddEvent("event pinged")

		err = s.SharedPersistentState.AddExhibit(eventCtx, app.Exhibit)
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
