package persistence

import (
	"context"
	"github.com/nats-io/nats.go"
	goredislib "github.com/redis/go-redis/v9"
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

	log.Debugw("connected to redis", "host", config.GetRedisHost())

	return redisClient
}

func NewNatsConn(config config.Config, log *zap.SugaredLogger) *nats.Conn {
	nc, err := nats.Connect(config.GetNatsHost())
	if err != nil {
		log.Panicw("error connecting to nats", "error", err)
	}

	log.Debugw("connected to nats", "host", config.GetNatsHost())

	return nc
}
