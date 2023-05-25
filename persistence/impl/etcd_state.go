package impl

import (
	"context"
	"encoding/json"
	"errors"
	etcd "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	recipe "go.etcd.io/etcd/client/v3/experimental/recipes"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"museum/config"
	"museum/domain"
	"museum/util"
	"strconv"
	"strings"
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

func (e EtcdState) SetRuntimeInfo(ctx context.Context, id string, runtimeInfo domain.ExhibitRuntimeInfo) error {
	key := "/" + e.Config.GetEtcdBaseKey() + "/" + id + "/" + "runtime_info"

	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("etcd persistence").
		Start(ctx, "SetRuntimeInfo", trace.WithAttributes(attribute.String("key", key), attribute.String("id", id)))
	defer span.End()

	b, err := json.Marshal(runtimeInfo)
	if err != nil {
		return err
	}

	_, err = e.Client.Put(subCtx, key, string(b))
	if err != nil {
		return err
	}

	span.AddEvent("set runtime info for exhibit")

	return nil
}

func (e EtcdState) GetRuntimeInfo(ctx context.Context, id string) (domain.ExhibitRuntimeInfo, error) {
	key := "/" + e.Config.GetEtcdBaseKey() + "/" + id + "/" + "runtime_info"

	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("etcd persistence").
		Start(ctx, "GetRuntimeInfo", trace.WithAttributes(attribute.String("key", key), attribute.String("id", id)))
	defer span.End()

	span.AddEvent("searching for runtime info for exhibit")

	resp, err := e.Client.Get(subCtx, key)
	if err != nil {
		return domain.ExhibitRuntimeInfo{}, err
	}

	span.AddEvent("found runtime info for exhibit")

	var runtimeInfo domain.ExhibitRuntimeInfo
	err = json.Unmarshal(resp.Kvs[0].Value, &runtimeInfo)
	if err != nil {
		return domain.ExhibitRuntimeInfo{}, err
	}

	return runtimeInfo, nil
}

func (e EtcdState) GetLastAccessed(ctx context.Context, id string) (int64, error) {
	key := "/" + e.Config.GetEtcdBaseKey() + "/" + id + "/" + "last_accessed"

	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("etcd persistence").
		Start(ctx, "GetLastAccessed", trace.WithAttributes(attribute.String("key", key), attribute.String("id", id)))
	defer span.End()

	span.AddEvent("searching for last_accessed time for exhibit")

	resp, err := e.Client.Get(subCtx, key)
	if err != nil {
		return -1, err
	}

	span.AddEvent("found last_accessed time for exhibit")

	i, err := strconv.ParseInt(string(resp.Kvs[0].Value), 10, 64)
	if err != nil {
		return -1, err
	}

	return i, nil
}

func (e EtcdState) SetLastAccessed(ctx context.Context, id string, lastAccessed int64) error {
	key := "/" + e.Config.GetEtcdBaseKey() + "/" + id + "/" + "last_accessed"

	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("etcd persistence").
		Start(ctx, "SetLastAccessed", trace.WithAttributes(attribute.String("key", key), attribute.String("id", id)))
	defer span.End()

	_, err := e.Client.Put(subCtx, key, strconv.FormatInt(lastAccessed, 10))
	if err != nil {
		return err
	}

	span.AddEvent("set last_accessed time for exhibit")

	return nil
}

func (e EtcdState) CreateExhibit(ctx context.Context, app domain.Exhibit) error {
	key := "/" + e.Config.GetEtcdBaseKey() + "/" + app.Id + "/" + "meta"

	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("etcd persistence").
		Start(ctx, "CreateExhibit", trace.WithAttributes(attribute.String("key", key), attribute.String("id", app.Id)))
	defer span.End()

	span.AddEvent("checking if exhibit already exists")

	// check if app already exists, if so, return error
	res, err := e.Client.Get(subCtx, key)
	if err != nil {
		return err
	}

	if res.Count > 0 {
		return errors.New("exhibit with id " + app.Id + " already exists")
	}

	b, err := json.Marshal(app)
	if err != nil {
		return err
	}

	_, err = e.Client.Put(subCtx, key, string(b))
	if err != nil {
		return err
	}

	span.AddEvent("added exhibit to etcd")

	return nil
}

func (e EtcdState) GetExhibitById(ctx context.Context, id string) (domain.Exhibit, error) {
	key := "/" + e.Config.GetEtcdBaseKey() + "/" + id + "/" + "meta"

	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("etcd persistence").
		Start(ctx, "GetExhibitById", trace.WithAttributes(attribute.String("key", key), attribute.String("id", id)))
	defer span.End()

	span.AddEvent("searching for exhibit")

	resp, err := e.Client.Get(subCtx, key)
	if err != nil {
		return domain.Exhibit{}, err
	}

	span.AddEvent("found exhibit")
	var exhibit domain.Exhibit
	err = json.Unmarshal(resp.Kvs[0].Value, &exhibit)
	if err != nil {
		return domain.Exhibit{}, err
	}

	return exhibit, nil
}

func (e EtcdState) GetAllExhibits(ctx context.Context) []domain.Exhibit {
	searchKey := "/" + e.Config.GetEtcdBaseKey() + "/"

	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("etcd persistence").
		Start(ctx, "GetAllExhibits", trace.WithAttributes(attribute.String("key", searchKey)))
	defer span.End()

	span.AddEvent("searching for exhibits")

	resp, err := e.Client.Get(subCtx, searchKey, etcd.WithPrefix())
	if err != nil {
		return []domain.Exhibit{}
	}

	span.AddEvent("found exhibits")
	var exhibits []domain.Exhibit
	for _, kv := range resp.Kvs {
		//check that kv ends with /object
		if !strings.HasSuffix(string(kv.Key), "/meta") {
			continue
		}

		var exhibit domain.Exhibit
		err := json.Unmarshal(kv.Value, &exhibit)
		if err != nil {
			continue
		}
		exhibits = append(exhibits, exhibit)
	}

	return exhibits
}

func (e EtcdState) DeleteExhibitById(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}
