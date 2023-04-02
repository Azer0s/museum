package impl

import (
	cloudevents "github.com/cloudevents/sdk-go/v2/event"
	"github.com/nats-io/nats.go"
	"museum/config"
)

type NatsEmitter struct {
	Conn   *nats.Conn
	Config config.Config
}

func (n NatsEmitter) EmitEvent(event cloudevents.Event) error {
	e, err := event.MarshalJSON()
	if err != nil {
		return err
	}

	err = n.Conn.Publish(n.Config.GetNatsSubject(), e)
	if err != nil {
		return err
	}

	return nil
}
