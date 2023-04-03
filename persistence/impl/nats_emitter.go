package impl

import (
	cloudevents "github.com/cloudevents/sdk-go/v2/event"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"museum/config"
)

type NatsEmitter struct {
	Conn   *nats.Conn
	Config config.Config
	Log    *zap.SugaredLogger
}

func (n NatsEmitter) EmitEvent(event cloudevents.Event) error {
	e, err := event.MarshalJSON()
	if err != nil {
		n.Log.Warnw("error marshalling event", "error", err)
		return err
	}

	err = n.Conn.Publish(n.Config.GetNatsSubject(), e)
	if err != nil {
		n.Log.Warnw("error publishing event", "error", err)
		return err
	}

	n.Log.Debugw("published event", "eventID", event.ID(), "eventType", event.Type())
	return nil
}
