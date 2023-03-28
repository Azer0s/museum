package persistence

import cloudevents "github.com/cloudevents/sdk-go/v2/event"

type Consumer interface {
	GetEvents() (<-chan cloudevents.Event, error)
}
