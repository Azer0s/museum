package persistence

import (
	"github.com/nats-io/nats.go"
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

func NewNatsClient(config config.Config, log *zap.SugaredLogger) *nats.Conn {
	conn, err := nats.Connect(config.GetNatsHost())
	if err != nil {
		log.Fatalw("failed to connect to NATS", "error", err)
	}

	if !conn.IsConnected() {
		log.Fatalw("failed to connect to NATS")
	}

	log.Debugw("connected to NATS", "host", config.GetNatsHost())

	return conn
}
