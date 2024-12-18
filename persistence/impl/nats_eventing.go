package impl

import (
	"context"
	"encoding/json"
	"errors"
	cloudevents "github.com/cloudevents/sdk-go/v2"
	"github.com/google/uuid"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"museum/config"
	"museum/domain"
	"strconv"
)

type NatsEventing struct {
	Config   config.Config
	Log      *zap.SugaredLogger
	Provider trace.TracerProvider
	Conn     *nats.Conn
}

func (n NatsEventing) DispatchExhibitCreatedEvent(ctx context.Context, exhibit domain.Exhibit) {
	_, span := n.Provider.
		Tracer("nats eventing").
		Start(ctx, "DispatchExhibitCreatedEvent", trace.WithAttributes(attribute.String("exhibitId", exhibit.Id)))
	defer span.End()

	n.Log.Debugw("nats eventing dispatching exhibit created event", "exhibitId", exhibit.Id)
	span.AddEvent("dispatching exhibit created event")

	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())
	event.SetSource("museum")
	event.SetType("exhibit.created")
	err := event.SetData(cloudevents.ApplicationJSON, map[string]string{"exhibitId": exhibit.Id})
	if err != nil {
		n.Log.Errorw("error setting event data", "error", err)
		span.RecordError(err)
		return
	}

	bytes, err := event.MarshalJSON()
	if err != nil {
		n.Log.Errorw("error marshalling event", "error", err)
		span.RecordError(err)
		return
	}

	err = n.Conn.Publish(n.Config.GetNatsBaseKey()+".exhibit.created", bytes)
	if err != nil {
		n.Log.Errorw("error publishing exhibit created event", "error", err)
		span.RecordError(err)
		return
	}
}

func (n NatsEventing) DispatchExhibitStartingEvent(ctx context.Context, exhibit domain.Exhibit, currentStepCount *int, step domain.ExhibitStartingStep) {
	_, span := n.Provider.
		Tracer("nats eventing").
		Start(ctx, "DispatchExhibitStartingEvent", trace.WithAttributes(
			attribute.String("exhibitId", exhibit.Id),
			attribute.String("object", exhibit.Objects[step.Object].Name),
			attribute.String("step", step.Step.String()),
		))
	defer span.End()

	if step.Error == nil {
		stepStr := strconv.Itoa(*currentStepCount)
		totalStepStr := strconv.Itoa(exhibit.GetTotalSteps())
		n.Log.Debugw("nats eventing dispatching exhibit starting event "+stepStr+"/"+totalStepStr,
			"exhibitId", exhibit.Id, "object", exhibit.Objects[step.Object].Name, "step", step.Step.String())
		span.AddEvent("dispatching exhibit starting event")
	} else {
		n.Log.Debugw("nats eventing dispatching exhibit starting event with error",
			"exhibitId", exhibit.Id, "object", exhibit.Objects[step.Object].Name, "step", step.Step.String(), "error", step.Error)
		span.AddEvent("dispatching exhibit starting event with error")
	}

	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())
	event.SetSource("museum")
	event.SetType("exhibit.starting")

	errStr := ""
	if step.Error != nil {
		errStr = step.Error.Error()
	}

	err := event.SetData(cloudevents.ApplicationJSON, domain.ExhibitStartingStepEvent{
		ExhibitId:        exhibit.Id,
		Object:           exhibit.Objects[step.Object].Name,
		Step:             step.Step.String(),
		CurrentStepCount: *currentStepCount,
		TotalStepCount:   exhibit.GetTotalSteps(),
		Error:            errStr,
	})

	*currentStepCount++

	if err != nil {
		n.Log.Errorw("error setting event data", "error", err)
		span.RecordError(err)
		return
	}

	bytes, err := event.MarshalJSON()
	if err != nil {
		n.Log.Errorw("error marshalling event", "error", err)
		span.RecordError(err)
		return
	}

	err = n.Conn.Publish(n.Config.GetNatsBaseKey()+".exhibit."+exhibit.Id+".starting", bytes)
	if err != nil {
		n.Log.Errorw("error publishing exhibit starting event", "error", err)
		span.RecordError(err)
		return
	}
}

func (n NatsEventing) DispatchExhibitStoppingEvent(ctx context.Context, exhibit domain.Exhibit) {
	_, span := n.Provider.
		Tracer("nats eventing").
		Start(ctx, "DispatchExhibitStoppingEvent", trace.WithAttributes(attribute.String("exhibitId", exhibit.Id)))
	defer span.End()

	n.Log.Debugw("nats eventing dispatching exhibit stopping event", "exhibitId", exhibit.Id)
	span.AddEvent("dispatching exhibit stopping event")

	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())
	event.SetSource("museum")
	event.SetType("exhibit.stopping")
	err := event.SetData(cloudevents.ApplicationJSON, map[string]string{"exhibitId": exhibit.Id})
	if err != nil {
		n.Log.Errorw("error setting event data", "error", err)
		span.RecordError(err)
		return
	}

	bytes, err := event.MarshalJSON()
	if err != nil {
		n.Log.Errorw("error marshalling event", "error", err)
		span.RecordError(err)
		return
	}

	err = n.Conn.Publish(n.Config.GetNatsBaseKey()+".exhibit."+exhibit.Id+".stopping", bytes)
	if err != nil {
		n.Log.Errorw("error publishing exhibit stopping event", "error", err)
		span.RecordError(err)
		return
	}
}

func (n NatsEventing) GetExhibitStartingChannel(exhibitId string, parentCtx context.Context) (<-chan domain.ExhibitStartingStepEvent, context.CancelFunc, error) {
	subChan := make(chan domain.ExhibitStartingStepEvent)

	sync, err := n.Conn.SubscribeSync(n.Config.GetNatsBaseKey() + ".exhibit." + exhibitId + ".starting")
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithCancel(parentCtx)

	go func() {
		for {
			msg, err := sync.NextMsgWithContext(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					n.Log.Debugw("context cancelled, stopping exhibit starting channel", "exhibitId", exhibitId)
					err := sync.Unsubscribe()
					if err != nil {
						n.Log.Warnw("error unsubscribing from exhibit starting channel", "error", err, "exhibitId", exhibitId)
					}
					return
				}

				n.Log.Errorw("error getting next message", "error", err)
				continue
			}

			event := cloudevents.Event{}
			err = json.Unmarshal(msg.Data, &event)
			if err != nil {
				n.Log.Errorw("error unmarshalling event", "error", err)
				continue
			}

			eventData := domain.ExhibitStartingStepEvent{}
			err = json.Unmarshal(event.Data(), &eventData)

			subChan <- eventData
		}
	}()

	return subChan, cancel, nil
}

func (n NatsEventing) GetExhibitStoppingChannel(exhibitId string, parentCtx context.Context) (<-chan domain.ExhibitStoppingEvent, context.CancelFunc, error) {
	subChan := make(chan domain.ExhibitStoppingEvent)

	sync, err := n.Conn.SubscribeSync(n.Config.GetNatsBaseKey() + ".exhibit." + exhibitId + ".stopping")
	if err != nil {
		return nil, nil, err
	}

	ctx, cancel := context.WithCancel(parentCtx)

	go func() {
		for {
			msg, err := sync.NextMsgWithContext(ctx)
			if err != nil {
				if errors.Is(err, context.Canceled) {
					n.Log.Debugw("context cancelled, stopping exhibit stopping channel", "exhibitId", exhibitId)
					err := sync.Unsubscribe()
					if err != nil {
						n.Log.Warnw("error unsubscribing from exhibit stopping channel", "error", err, "exhibitId", exhibitId)
					}
					return
				}

				n.Log.Errorw("error getting next message", "error", err)
				continue
			}

			event := cloudevents.Event{}
			err = json.Unmarshal(msg.Data, &event)
			if err != nil {
				n.Log.Errorw("error unmarshalling event", "error", err)
				continue
			}

			eventData := domain.ExhibitStoppingEvent{}
			err = json.Unmarshal(event.Data(), &eventData)

			subChan <- eventData
		}
	}()

	return subChan, cancel, nil
}

func (n NatsEventing) CanReceive() bool {
	return true
}
