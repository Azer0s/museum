package persistence

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	goredislib "github.com/redis/go-redis/v9"
	"github.com/segmentio/kafka-go"
	"go.uber.org/zap"
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
		Config:        config,
	}
}

func NewSharedPersistentEmittedState(state SharedPersistentState, emitter Emitter, consumer Consumer, log *zap.SugaredLogger) SharedPersistentEmittedState {
	sb := &StateBundle{
		SharedPersistentState: state,
		Emitter:               emitter,
		Consumer:              consumer,
		ConfirmEvents:         make(map[string]chan struct{}),
		ConfirmEventsMutex:    &sync.RWMutex{},
		Log:                   log,
	}
	events, err := consumer.GetEvents()

	err = state.WithLock(func() error {
		go func() {
			for {
				m := <-events

				sb.ConfirmEventsMutex.RLock()
				if ch, ok := sb.ConfirmEvents[m.ID()]; ok {
					ch <- struct{}{}
					delete(sb.ConfirmEvents, m.ID())
					sb.ConfirmEventsMutex.RUnlock()
					continue
				}
				sb.ConfirmEventsMutex.RUnlock()

				// TODO: handle other events
			}
		}()

		currentState, err := state.GetExhibits()
		if err != nil {
			log.Errorw("failed to get current state", "error", err)
			return err
		}

		sb.CurrentState = currentState
		sb.CurrentStateMutex = &sync.RWMutex{}

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	return sb
}
