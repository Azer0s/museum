package persistence

import (
	"context"
	"museum/domain"
)

type Eventing interface {
	DispatchExhibitCreatedEvent(ctx context.Context, exhibit domain.Exhibit)
	DispatchExhibitStartingEvent(ctx context.Context, exhibit domain.Exhibit, currentStepCount *int, step domain.ExhibitStartingStep)

	GetExhibitMetadataChannel() chan domain.ExhibitMetadata
}
