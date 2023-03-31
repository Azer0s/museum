package persistence

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"museum/config"
	"museum/persistence/impl"
	"sync"
	"time"
)

func NewKafkaEmitter(producer *kafka.Writer) Emitter {
	return &impl.KafkaEmitter{
		Writer: producer,
	}
}

func NewRedisStateConnector(config config.Config, redisClient *goredislib.Client) SharedPersistentState {
	rs := &impl.RedisStateConnector{RedisClient: redisClient, Config: config}
	rs.RedisPool = goredis.NewPool(rs.RedisClient)
	rs.RedisSync = redsync.New(rs.RedisPool)
	rs.RedisMu = rs.RedisSync.NewMutex(config.GetRedisBaseKey()+":state:lock", redsync.WithTries(1), redsync.WithExpiry(1*time.Minute))

	return rs
}

func NewKafkaConsumer(config config.Config, consumerGroup *kafka.ConsumerGroup) Consumer {
	return &impl.KafkaConsumer{
		ConsumerGroup: consumerGroup,
		Brokers:       config.GetKafkaBrokers(),
	}
}

func NewSharedPersistentEmittedState(state SharedPersistentState, emitter Emitter, consumer Consumer) SharedPersistentEmittedState {
	sb := &StateBundle{
		SharedPersistentState: state,
		Emitter:               emitter,
		Consumer:              consumer,
	}

	err := state.WithLock(func() error {
		// TODO: start go routine to listen for kafka messages

		currentState, err := state.GetExhibits()
		if err != nil {
			return err
		}

		sb.CurrentState = currentState
		sb.CurrentStateMutex = &sync.RWMutex{}

		return nil
	})

	if err != nil {
		panic(err)
	}

	return sb
}
