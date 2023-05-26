package persistence

import (
	etcd "go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"museum/config"
	"museum/observability"
	"museum/persistence/impl"
)

func NewEtcdState(config config.Config, etcdClient *etcd.Client, providerFactory *observability.TracerProviderFactory, log *zap.SugaredLogger) State {
	etcdState := &impl.EtcdState{
		Client:   etcdClient,
		Config:   config,
		Provider: providerFactory.Build("etcd"),
		Log:      log,
	}

	etcdState.Init()
	return etcdState
}
