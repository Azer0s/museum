package persistence

import (
	"context"
	"museum/domain"
)

type Eventing interface {
	DispatchExhibitCreatedEvent(ctx context.Context, exhibit domain.Exhibit)
	DispatchExhibitStartingEvent(ctx context.Context, exhibit domain.Exhibit, currentStepCount *int, step domain.ExhibitStartingStep)
	DispatchExhibitStoppingEvent(ctx context.Context, exhibit domain.Exhibit)

	GetExhibitStartingChannel(exhibitId string, ctx context.Context) (<-chan domain.ExhibitStartingStepEvent, context.CancelFunc, error)
	GetExhibitStoppingChannel(exhibitId string, parentCtx context.Context) (<-chan domain.ExhibitStoppingEvent, context.CancelFunc, error)
	CanReceive() bool
}
