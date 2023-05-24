package persistence

import (
	"go.etcd.io/etcd/client/v3"
	"go.uber.org/zap"
	"museum/config"
	"time"
)

func NewEtcdClient(config config.Config, log *zap.SugaredLogger) *clientv3.Client {
	client, err := clientv3.New(clientv3.Config{
		Endpoints:   []string{config.GetEtcdHost()},
		DialTimeout: 5 * time.Second,
	})
	if err != nil {
		log.Panicw("error connecting to etcd", "error", err)
	}

	log.Debugw("connected to etcd", "host", config.GetEtcdHost())

	return client
}
