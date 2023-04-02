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
	eventChan := make(chan cloudevents.Event, 10)

	go func() {
		_, err := n.Conn.Subscribe(n.Config.GetNatsSubject(), func(msg *nats.Msg) {
			var event cloudevents.Event
			err := event.UnmarshalJSON(msg.Data)
			if err != nil {
				return
			}

			eventChan <- event
		})

		if err != nil {
			n.Log.Panicw("error subscribing to nats topic", "error", err)
		}
	}()

	return eventChan, nil
}
