package persistence

import "museum/domain"

type Eventing interface {
	DispatchExhibitCreatedEvent(exhibit domain.Exhibit)
	DispatchExhibitStartingEvent(exhibit domain.Exhibit, step domain.ExhibitStartingStep)

	GetExhibitMetadataChannel() chan domain.ExhibitMetadata
}
