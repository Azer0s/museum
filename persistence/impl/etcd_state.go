package impl

import (
	"context"
	etcd "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	recipe "go.etcd.io/etcd/client/v3/experimental/recipes"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"museum/config"
	"museum/util"
)

type EtcdState struct {
	Client   *etcd.Client
	Config   config.Config
	Provider trace.TracerProvider
	Session  *concurrency.Session
	Log      *zap.SugaredLogger

	//TODO: reintroduce cache
	//etcd supports notifiying on changes, so we can use that to invalidate the cache
	//we should get all keys on startup and then cache them
	//for this we'll need a mutex to lock the entire etcd state
	//this mutex will only be used for museum startup, so it should be fine
}

func (e EtcdState) GetRwLock(ctx context.Context, id string, lockName string) util.RwErrMutex {
	key := "/" + e.Config.GetEtcdBaseKey() + "/" + id + "/" + "locks" + "/" + lockName

	// create new trace span for event service
	_, span := e.Provider.
		Tracer("etcd persistence").
		Start(ctx, "WithRwLock", trace.WithAttributes(attribute.String("key", key), attribute.String("id", id), attribute.String("lockName", lockName)))
	defer span.End()

	span.AddEvent("retrieving lock")

	// create lock
	lock := recipe.NewRWMutex(e.Session, key)

	span.AddEvent("lock retrieved")

	return lock
}
