package persistence

import (
	"github.com/nats-io/nats.go"
	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"museum/config"
	"museum/observability"
	"museum/persistence/impl"
	"time"
)

func NewEtcdState(config config.Config, etcdClient *etcd.Client, providerFactory *observability.TracerProviderFactory, log *zap.SugaredLogger) State {
	etcdState := &impl.EtcdState{
		Client:   etcdClient,
		Config:   config,
		Provider: providerFactory.Build("etcd"),
		Log:      log,
	}

	done := make(chan bool)
	go func() {
		etcdState.Init()
		done <- true
	}()

	select {
	case <-done:
		break
	case <-time.After(10 * time.Second):
		log.Fatalw("etcd persistence initialization timed out")
	}

	log.Debugw("etcd persistence initialized")

	return etcdState
}

func NewNatsEventing(config config.Config, log *zap.SugaredLogger, conn *nats.Conn, providerFactory *observability.TracerProviderFactory) Eventing {
	log.Debugw("using nats eventing")
	return &impl.NatsEventing{
		Config:   config,
		Log:      log,
		Provider: providerFactory.Build("nats"),
		Conn:     conn,
	}
}

func NewNoopEventing(log *zap.SugaredLogger) Eventing {
	log.Debugw("using noop eventing")
	return &impl.NoopEventing{
		Log: log,
	}
}
