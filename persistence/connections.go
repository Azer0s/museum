package persistence

import (
	"context"
	"github.com/google/uuid"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
	"museum/config"
)

func NewRedisClient(config config.Config, log *zap.SugaredLogger) *goredislib.Client {
	redisClient := goredislib.NewClient(&goredislib.Options{
		Addr: config.GetRedisHost(),
	})

	if _, err := redisClient.Ping(context.Background()).Result(); err != nil {
		log.Panic(err)
	}

	return redisClient
}

func NewKafkaWriter(config config.Config) *kafka.Writer {
	return &kafka.Writer{
		Addr:  kafka.TCP(config.GetKafkaBrokers()[0]),
		Topic: config.GetKafkaTopic(),
	}
}

func NewKafkaConsumerGroup(config config.Config, log *zap.SugaredLogger) *kafka.ConsumerGroup {
	client := kafka.Client{
		Addr: kafka.TCP(config.GetKafkaBrokers()[0]),
	}
	_, err := client.CreateTopics(context.Background(), &kafka.CreateTopicsRequest{
		Topics: []kafka.TopicConfig{
			{Topic: config.GetKafkaTopic()},
		},
	})
	if err != nil {
		log.Panic(err)
	}

	group, err := kafka.NewConsumerGroup(kafka.ConsumerGroupConfig{
		ID:      uuid.New().String(),
		Brokers: config.GetKafkaBrokers(),
		Topics:  []string{config.GetKafkaTopic()},
	})

	if err != nil {
		log.Panic(err)
	}

	return group
}
