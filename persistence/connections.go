package persistence

import (
	"context"
	"github.com/google/uuid"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"museum/config"
)

func NewRedisClient(config config.Config) *goredislib.Client {
	redisClient := goredislib.NewClient(&goredislib.Options{
		Addr: config.GetRedisHost(),
	})

	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		panic(err)
		return nil
	}

	return redisClient
}

func NewKafkaWriter(config config.Config) *kafka.Writer {
	return &kafka.Writer{
		Addr:  kafka.TCP(config.GetKafkaBrokers()[0]),
		Topic: config.GetKafkaTopic(),
	}
}

func NewKafkaConsumerGroup(config config.Config) *kafka.ConsumerGroup {
	group, err := kafka.NewConsumerGroup(kafka.ConsumerGroupConfig{
		ID:      uuid.New().String(),
		Brokers: config.GetKafkaBrokers(),
		Topics:  []string{config.GetKafkaTopic()},
	})

	if err != nil {
		panic(err)
	}

	return group
}
