package domain

import cloudevents "github.com/cloudevents/sdk-go/v2/event"

const (
	CreateEventType       = "museum.application.create"
	DeleteEventType       = "museum.application.delete"
	LeaseRenewedEventType = "museum.application.lease.renewed"
	LeaseExpiredEventType = "museum.application.lease.expired"

	source      = "museum"
	contentType = "application/json"
)

func newEvent(eventType string, application Application) (cloudevents.Event, error) {
	event := cloudevents.New()

	event.SetType(eventType)
	event.SetSource(source)
	err := event.SetData(contentType, application)
	if err != nil {
		return cloudevents.Event{}, err
	}

	return event, nil
}

func NewCreateEvent(application Application) (cloudevents.Event, error) {
	return newEvent(CreateEventType, application)
}

func NewDeleteEvent(application Application) (cloudevents.Event, error) {
	return newEvent(DeleteEventType, application)
}

func NewLeaseRenewedEvent(application Application) (cloudevents.Event, error) {
	return newEvent(LeaseRenewedEventType, application)
}

func NewLeaseExpiredEvent(application Application) (cloudevents.Event, error) {
	return newEvent(LeaseExpiredEventType, application)
}
