package persistence

import (
	"github.com/go-redsync/redsync/v4"
	"github.com/go-redsync/redsync/v4/redis/goredis/v9"
	"github.com/nats-io/nats.go"
	goredislib "github.com/redis/go-redis/v9"
	"go.uber.org/zap"
	"museum/config"
	"museum/observability"
	"museum/persistence/impl"
	"sync"
	"time"
)

func NewNatsEmitter(config config.Config, conn *nats.Conn, log *zap.SugaredLogger) Emitter {
	return &impl.NatsEmitter{
		Conn:   conn,
		Config: config,
		Log:    log,
	}
}

func NewRedisStateConnector(config config.Config, redisClient *goredislib.Client, providerFactory *observability.TracerProviderFactory) SharedPersistentState {
	rs := &impl.RedisStateConnector{
		RedisClient: redisClient,
		Config:      config,
		Provider:    providerFactory.Build("redis"),
	}
	rs.RedisPool = goredis.NewPool(rs.RedisClient)
	rs.RedisSync = redsync.New(rs.RedisPool)
	rs.RedisMu = rs.RedisSync.NewMutex(config.GetRedisBaseKey()+":state:lock", redsync.WithTries(1), redsync.WithExpiry(1*time.Minute))

	return rs
}

func NewNatsConsumer(config config.Config, conn *nats.Conn, log *zap.SugaredLogger) Consumer {
	return &impl.NatsConsumer{
		Conn:   conn,
		Config: config,
		Log:    log,
	}
}

func NewSharedPersistentEmittedState(state SharedPersistentState, emitter Emitter, consumer Consumer, log *zap.SugaredLogger, providerFactory *observability.TracerProviderFactory) SharedPersistentEmittedState {
	sb := &StateBundle{
		SharedPersistentState: state,
		Emitter:               emitter,
		Consumer:              consumer,
		ConfirmEvents:         make(map[string]chan struct{}),
		ConfirmEventsMutex:    &sync.RWMutex{},
		Log:                   log,
		Provider:              providerFactory.Build("event-hub"),
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
		log.Panicw("failed to initialize shared state", "error", err)
	}

	return sb
}
