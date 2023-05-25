package persistence

import (
	etcd "go.etcd.io/etcd/client/v3"
	etcdc "go.etcd.io/etcd/client/v3/concurrency"
	"go.uber.org/zap"
	"museum/config"
	"museum/observability"
	"museum/persistence/impl"
)

func NewEtcdState(config config.Config, etcdClient *etcd.Client, providerFactory *observability.TracerProviderFactory, log *zap.SugaredLogger) State {
	session, err := etcdc.NewSession(etcdClient)
	if err != nil {
		log.Fatalw("error creating etcd session", "error", err)
	}

	return &impl.EtcdState{
		Client:   etcdClient,
		Config:   config,
		Provider: providerFactory.Build("etcd"),
		Session:  session,
		Log:      log,
	}
}
