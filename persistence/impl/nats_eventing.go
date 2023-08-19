package impl

import (
	"context"
	"github.com/nats-io/nats.go"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"museum/config"
	"museum/domain"
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

	err := n.Conn.Publish(n.Config.GetNatsBaseKey()+".exhibit.created", []byte(exhibit.Id))
	if err != nil {
		n.Log.Errorw("error publishing exhibit created event", "error", err)
		span.RecordError(err)
		return
	}
}

func (n NatsEventing) DispatchExhibitStartingEvent(ctx context.Context, exhibit domain.Exhibit, step domain.ExhibitStartingStep) {
	//TODO implement me
	panic("implement me")
}

func (n NatsEventing) GetExhibitMetadataChannel() chan domain.ExhibitMetadata {
	//TODO implement me
	panic("implement me")
}
