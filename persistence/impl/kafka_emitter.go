package impl

import (
	"context"
	cloudevents "github.com/cloudevents/sdk-go/v2/event"
	"github.com/segmentio/kafka-go"
)

type KafkaEmitter struct {
	Writer *kafka.Writer
}

func (k KafkaEmitter) EmitEvent(event cloudevents.Event) error {
	e, err := event.MarshalJSON()
	if err != nil {
		return err
	}

	err = k.Writer.WriteMessages(context.Background(), kafka.Message{
		Key:   []byte(event.ID()),
		Value: e,
	})
	if err != nil {
		return err
	}

	return nil
}
