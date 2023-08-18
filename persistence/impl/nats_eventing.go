package impl

import (
	"go.uber.org/zap"
	"museum/config"
	"museum/domain"
)

type NatsEventing struct {
	Config config.Config
	Log    *zap.SugaredLogger
}

func (n NatsEventing) DispatchExhibitCreatedEvent(exhibit domain.Exhibit) {
	//TODO implement me
	panic("implement me")
}

func (n NatsEventing) DispatchExhibitStartingEvent(exhibit domain.Exhibit, step domain.ExhibitStartingStep) {
	//TODO implement me
	panic("implement me")
}

func (n NatsEventing) GetExhibitMetadataChannel() chan domain.ExhibitMetadata {
	//TODO implement me
	panic("implement me")
}
