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
	"museum/domain"
	"museum/util"
	"sync"
)

type EtcdState struct {
	Client   *etcd.Client
	Config   config.Config
	Provider trace.TracerProvider
	Session  *concurrency.Session
	Log      *zap.SugaredLogger

	ExhibitCache   map[string]domain.Exhibit
	ExhibitCacheMu *sync.RWMutex

	RuntimeInfoCache   map[string]domain.ExhibitRuntimeInfo
	RuntimeInfoCacheMu *sync.RWMutex
}

func (e *EtcdState) Init() {
	e.Log.Infow("initializing etcd persistence")

	// create etcd session
	session, err := concurrency.NewSession(e.Client)
	if err != nil {
		e.Log.Fatalw("error creating etcd session", "error", err)
	}
	e.Log.Debugw("etcd session created")
	e.Session = session

	e.Log.Infow("retrieving all exhibits")
	exhibits := e.GetAllExhibits(context.Background())
	e.Log.Debugw("retrieved all exhibits")

	e.ExhibitCache = make(map[string]domain.Exhibit)
	e.ExhibitCacheMu = &sync.RWMutex{}

	for _, exhibit := range exhibits {
		e.ExhibitCache[exhibit.Id] = exhibit
		e.watchExhibit(exhibit.Id)
	}

	e.Log.Infow("retrieving all exhibit runtime info")
	runtimeInfo := make(map[string]domain.ExhibitRuntimeInfo)
	for _, exhibit := range exhibits {
		info, err := e.GetRuntimeInfo(context.Background(), exhibit.Id)
		if err != nil {
			e.Log.Errorw("error retrieving exhibit runtime info", "error", err)
			continue
		}
		runtimeInfo[exhibit.Id] = info
	}
	e.Log.Debugw("retrieved all exhibit runtime info")

	e.RuntimeInfoCache = runtimeInfo
	e.RuntimeInfoCacheMu = &sync.RWMutex{}

	for _, exhibit := range exhibits {
		e.watchRuntimeInfo(exhibit.Id)
	}

	createChan := e.Client.Watch(context.Background(), "/"+e.Config.GetEtcdBaseKey()+"/")
	go func(w etcd.WatchChan) {
		for {
			e.handleCreateEvent(w)
		}
	}(createChan)
}

func (e *EtcdState) GetRwLock(ctx context.Context, id string, lockName string) util.RwErrMutex {
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
