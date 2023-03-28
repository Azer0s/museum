package persistence

import (
	"context"
	"github.com/google/uuid"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"museum/persistence/impl"
)

func NewRedisClient() (*goredislib.Client, func()) {
	redisClient := goredislib.NewClient(&goredislib.Options{
		Addr: "localhost:6379",
	})

	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		panic(err)
		return nil, nil
	}

	return redisClient, func() {
		err := redisClient.Close()
		if err != nil {
			panic(err)
		}
	}
}

func NewKafkaWriter() *kafka.Writer {
	return &kafka.Writer{
		Addr:  kafka.TCP("localhost:9092"),
		Topic: impl.KafkaTopic,
	}
}

func NewKafkaConsumerGroup() (*kafka.ConsumerGroup, error) {
	return kafka.NewConsumerGroup(kafka.ConsumerGroupConfig{
		ID:      uuid.New().String(),
		Brokers: []string{"localhost:9092"},
		Dialer:  nil,
		Topics:  []string{impl.KafkaTopic},
	})
}
