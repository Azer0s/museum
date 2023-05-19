package persistence

import (
	"context"
	"errors"
	"github.com/cloudevents/sdk-go/v2/event"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"museum/domain"
	"sync"
)

type StateBundle struct {
	SharedPersistentState SharedPersistentState
	Emitter               Emitter
	Consumer              Consumer
	CurrentState          *[]domain.Exhibit
	CurrentStateMutex     *sync.RWMutex
	ConfirmEvents         map[string]chan struct{}
	ConfirmEventsMutex    *sync.RWMutex
	Log                   *zap.SugaredLogger
	Provider              trace.TracerProvider
}

func (s StateBundle) GetExhibitById(id string) (*domain.Exhibit, error) {
	s.CurrentStateMutex.RLock()
	defer s.CurrentStateMutex.RUnlock()
	for _, exhibit := range *s.CurrentState {
		if exhibit.Id == id {
			return &exhibit, nil
		}
	}
	return nil, errors.New("exhibit not found")
}

func (s StateBundle) updateExhibit(ctx context.Context, app domain.Exhibit, event event.Event) error {
	s.CurrentStateMutex.Lock()
	defer s.CurrentStateMutex.Unlock()

	eventCtx, eventSpan := s.Provider.
		Tracer("Event emission").
		Start(ctx, "emitUpdateEvent")

	spanEnded := false
	defer func() {
		if !spanEnded {
			eventSpan.End()
		}
	}()

	eventSpan.AddEvent("event created")

	received, err := s.EventReceived(event.ID())
	if err != nil {
		s.Log.Debugw("failed create event receiver", "error", err)
	}

	err = s.Emitter.EmitEvent(event)
	if err != nil {
		s.Log.Debugw("failed to emit event", "error", err)
		return err
	}

	eventSpan.AddEvent("event emitted")

	<-received

	eventSpan.AddEvent("event pinged")

	err = s.SharedPersistentState.UpdateExhibit(eventCtx, app)
	if err != nil {
		s.Log.Debugw("failed to update exhibit", "error", err)
		return err
	}

	for i, exhibit := range *s.CurrentState {
		if exhibit.Id == app.Id {
			(*s.CurrentState)[i] = app
		}
	}

	eventSpan.AddEvent("exhibit updated")

	return nil
}

func (s StateBundle) StartingExhibit(ctx context.Context, app domain.Exhibit) error {
	updateEvent, err := domain.NewStartingEvent(app)
	if err != nil {
		s.Log.Debugw("failed to create event", "error", err)
		return err
	}
	return s.updateExhibit(ctx, app, updateEvent)
}

func (s StateBundle) StartExhibit(ctx context.Context, app domain.Exhibit) error {
	updateEvent, err := domain.NewStartEvent(app)
	if err != nil {
		s.Log.Debugw("failed to create event", "error", err)
		return err
	}

	return s.updateExhibit(ctx, app, updateEvent)
}

func (s StateBundle) RenewExhibitLease(ctx context.Context, app domain.Exhibit) error {
	renewEvent, err := domain.NewLeaseRenewedEvent(app)
	if err != nil {
		s.Log.Debugw("failed to create event", "error", err)
		return err
	}

	return s.updateExhibit(ctx, app, renewEvent)
}

func (s StateBundle) ExpireExhibitLease(ctx context.Context, app domain.Exhibit) error {
	expireEvent, err := domain.NewLeaseExpiredEvent(app)
	if err != nil {
		s.Log.Debugw("failed to create event", "error", err)
		return err
	}

	return s.updateExhibit(ctx, app, expireEvent)
}

func (s StateBundle) GetExhibits() []domain.Exhibit {
	// return s.SharedPersistentState.GetExhibits()
	s.CurrentStateMutex.RLock()
	defer s.CurrentStateMutex.RUnlock()
	return *s.CurrentState
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

		*s.CurrentState = append(*s.CurrentState, app.Exhibit)

		eventSpan.AddEvent("exhibit added")

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
