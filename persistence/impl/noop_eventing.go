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
	n.Log.Debugw("noop eventing dispatching exhibit created event", "exhibitId", exhibit.Id)
}

func (n NoopEventing) DispatchExhibitStartingEvent(ctx context.Context, exhibit domain.Exhibit, currentStepCount *int, step domain.ExhibitStartingStep) {
	n.Log.Debugw("noop eventing dispatching exhibit starting event", "exhibitId", exhibit.Id, "step", step)
}

func (n NoopEventing) GetExhibitMetadataChannel() <-chan domain.ExhibitMetadata {
	return make(chan domain.ExhibitMetadata)
}

func (n NoopEventing) GetExhibitStartingChannel(string, context.Context) (<-chan domain.ExhibitStartingStepEvent, context.CancelFunc, error) {
	return make(chan domain.ExhibitStartingStepEvent), func() {}, nil
}

func (n NoopEventing) CanReceive() bool {
	return false
}
