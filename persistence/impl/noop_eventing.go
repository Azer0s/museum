package impl

import (
	"context"
	"go.uber.org/zap"
	"museum/domain"
)

type NoopEventing struct {
	Log *zap.SugaredLogger
}

func (n NoopEventing) DispatchExhibitCreatedEvent(ctx context.Context, exhibit domain.Exhibit) {
	n.Log.Debugw("noop eventing dispatching exhibit created event", "exhibit", exhibit)
}

func (n NoopEventing) DispatchExhibitStartingEvent(ctx context.Context, exhibit domain.Exhibit, step domain.ExhibitStartingStep) {
	n.Log.Debugw("noop eventing dispatching exhibit starting event", "exhibit", exhibit, "step", step)
}

func (n NoopEventing) GetExhibitMetadataChannel() chan domain.ExhibitMetadata {
	return make(chan domain.ExhibitMetadata)
}
