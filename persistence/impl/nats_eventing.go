package impl

import (
	"context"
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

	stepStr := strconv.Itoa(*currentStepCount)
	totalStepStr := strconv.Itoa(exhibit.GetTotalSteps())
	n.Log.Debugw("nats eventing dispatching exhibit starting event "+stepStr+"/"+totalStepStr,
		"exhibitId", exhibit.Id, "object", exhibit.Objects[step.Object].Name, "step", step.Step.String())
	span.AddEvent("dispatching exhibit starting event")

	*currentStepCount++

	event := cloudevents.NewEvent()
	event.SetID(uuid.New().String())
	event.SetSource("museum")
	event.SetType("exhibit.starting")
	err := event.SetData(cloudevents.ApplicationJSON, map[string]string{
		"exhibitId": exhibit.Id,
		"object":    exhibit.Objects[step.Object].Name,
		"step":      step.Step.String(),
	})
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

	err = n.Conn.Publish(n.Config.GetNatsBaseKey()+".exhibit.starting", bytes)
	if err != nil {
		n.Log.Errorw("error publishing exhibit starting event", "error", err)
		span.RecordError(err)
		return
	}
}

func (n NatsEventing) GetExhibitMetadataChannel() chan domain.ExhibitMetadata {
	//TODO implement me
	panic("implement me")
}
