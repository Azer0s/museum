package impl

import (
	"go.uber.org/zap"
	"museum/domain"
)

type NoopEventing struct {
	Log *zap.SugaredLogger
}

func (n NoopEventing) DispatchExhibitCreatedEvent(exhibit domain.Exhibit) {
	n.Log.Debugw("noop eventing dispatching exhibit created event", "exhibit", exhibit)
}

func (n NoopEventing) DispatchExhibitStartingEvent(exhibit domain.Exhibit, step domain.ExhibitStartingStep) {
	n.Log.Debugw("noop eventing dispatching exhibit starting event", "exhibit", exhibit, "step", step)
}

func (n NoopEventing) GetExhibitMetadataChannel() chan domain.ExhibitMetadata {
	return make(chan domain.ExhibitMetadata)
}
