package persistence

import (
	cloudevents "github.com/cloudevents/sdk-go/v2/event"
)

type Emitter interface {
	EmitEvent(event cloudevents.Event) error
}
