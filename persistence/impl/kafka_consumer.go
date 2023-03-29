package impl

import (
	"context"
	"errors"
	cloudevents "github.com/cloudevents/sdk-go/v2/event"
	"github.com/segmentio/kafka-go"
	"museum/config"
)

type KafkaConsumer struct {
	ConsumerGroup *kafka.ConsumerGroup
	Brokers       []string
	Config        config.Config
}

type generationStartInfo struct {
	partition int
	offset    int64
	gen       *kafka.Generation
	eventChan chan<- cloudevents.Event
}

func readAllMessages(config config.Config, reader *kafka.Reader, startInfo *generationStartInfo) {
	partition, offset, gen := startInfo.partition, startInfo.offset, startInfo.gen
	for {
		msg, err := reader.ReadMessage(context.Background())
		if err != nil {
			if errors.Is(err, kafka.ErrGenerationEnded) {
				err := gen.CommitOffsets(map[string]map[int]int64{config.GetKafkaTopic(): {partition: offset + 1}})
				if err != nil {
					return
				}
				return
			}

			//fmt.Printf("error reading message: %+v\n", err)
			return
		}

		event := cloudevents.New()
		err = event.UnmarshalJSON(msg.Value)
		if err != nil {
			// this should never happen
			panic(err)
		}

		startInfo.eventChan <- event

		offset = msg.Offset
	}
}

func (k KafkaConsumer) handleGenerationStart(startInfo *generationStartInfo) func(ctx context.Context) {
	partition, offset := startInfo.partition, startInfo.offset
	return func(ctx context.Context) {
		// create reader for this partition.
		reader := kafka.NewReader(kafka.ReaderConfig{
			Brokers:   k.Brokers,
			Topic:     k.Config.GetKafkaTopic(),
			Partition: partition,
		})
		defer func(reader *kafka.Reader) {
			err := reader.Close()
			if err != nil {
				panic(err)
			}
		}(reader)

		// seek to the last committed offset for this partition.
		err := reader.SetOffset(offset)
		if err != nil {
			return
		}

		readAllMessages(k.Config, reader, startInfo)
	}
}

func (k KafkaConsumer) GetEvents() (<-chan cloudevents.Event, error) {
	eventChan := make(chan cloudevents.Event, 10)

	go func() {
		defer close(eventChan)

		for {
			gen, err := k.ConsumerGroup.Next(context.Background())
			if err != nil {
				panic(err)
			}

			assignments := gen.Assignments[k.Config.GetKafkaTopic()]
			for _, assignment := range assignments {
				gen.Start(k.handleGenerationStart(&generationStartInfo{
					partition: assignment.ID,
					offset:    assignment.Offset,
					gen:       gen,
					eventChan: eventChan,
				}))
			}
		}
	}()

	return eventChan, nil
}
