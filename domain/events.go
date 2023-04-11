package domain

import (
	cloudevents "github.com/cloudevents/sdk-go/v2/event"
	"github.com/google/uuid"
)

const (
	CreateEventType       = "museum.exhibit.create"
	DeleteEventType       = "museum.exhibit.delete"
	LeaseRenewedEventType = "museum.exhibit.lease.renewed"
	LeaseExpiredEventType = "museum.exhibit.lease.expired"
	StartEventType        = "museum.exhibit.start"
	StopEventType         = "museum.exhibit.stop"

	source      = "museum"
	contentType = "application/json"
)

func newEvent(eventType string, exhibit Exhibit) (cloudevents.Event, error) {
	event := cloudevents.New()

	event.SetType(eventType)
	event.SetSource(source)
	err := event.SetData(contentType, exhibit)
	if err != nil {
		return cloudevents.Event{}, err
	}

	event.SetID(uuid.New().String())

	return event, nil
}

func NewCreateEvent(exhibit Exhibit) (cloudevents.Event, error) {
	return newEvent(CreateEventType, exhibit)
}

func NewStartEvent(exhibit Exhibit) (cloudevents.Event, error) {
	return newEvent(StartEventType, exhibit)
}

func NewStopEvent(exhibit Exhibit) (cloudevents.Event, error) {
	return newEvent(StopEventType, exhibit)
}

func NewDeleteEvent(exhibit Exhibit) (cloudevents.Event, error) {
	return newEvent(DeleteEventType, exhibit)
}

func NewLeaseRenewedEvent(exhibit Exhibit) (cloudevents.Event, error) {
	return newEvent(LeaseRenewedEventType, exhibit)
}

func NewLeaseExpiredEvent(exhibit Exhibit) (cloudevents.Event, error) {
	return newEvent(LeaseExpiredEventType, exhibit)
}
