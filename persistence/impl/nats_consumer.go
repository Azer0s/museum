package impl

import (
	cloudevents "github.com/cloudevents/sdk-go/v2/event"
	"github.com/nats-io/nats.go"
	"go.uber.org/zap"
	"museum/config"
)

type NatsConsumer struct {
	Conn   *nats.Conn
	Config config.Config
	Log    *zap.SugaredLogger
}

func (n NatsConsumer) GetEvents() (<-chan cloudevents.Event, error) {
	eventChan := make(chan cloudevents.Event, 1000)

	go func() {
		_, err := n.Conn.Subscribe(n.Config.GetNatsSubject(), func(msg *nats.Msg) {
			var event cloudevents.Event
			err := event.UnmarshalJSON(msg.Data)
			if err != nil {
				n.Log.Errorw("error unmarshalling event", "error", err)
				return
			}

			n.Log.Debugw("received event", "eventID", event.ID(), "eventType", event.Type())
			eventChan <- event
		})

		if err != nil {
			n.Log.Panicw("error subscribing to nats topic", "error", err)
		}
	}()

	n.Log.Debugw("subscribed to subject", "subject", n.Config.GetNatsSubject())

	return eventChan, nil
}
