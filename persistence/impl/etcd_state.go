package impl

import (
	"context"
	"encoding/json"
	"errors"
	etcd "go.etcd.io/etcd/client/v3"
	"go.etcd.io/etcd/client/v3/concurrency"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"museum/config"
	"museum/domain"
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

func (e EtcdState) WithLock(ctx context.Context, id string, f func() error) (err error) {
	key := "/" + e.Config.GetEtcdBaseKey() + "/" + id + "/" + "lock"

	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("etcd persistence").
		Start(ctx, "WithLock", trace.WithAttributes(attribute.String("key", key), attribute.String("id", id)))
	defer span.End()

	span.AddEvent("acquiring lock")

	// create lock
	lock := concurrency.NewMutex(e.Session, key)
	err = lock.Lock(subCtx)
	if err != nil {
		return
	}

	// defer unlock
	defer func() {
		err = lock.Unlock(subCtx)
		if err != nil {
			e.Log.Errorw("error unlocking etcd lock", "error", err)
		}
		span.AddEvent("lock released")
	}()

	span.AddEvent("lock acquired")

	// execute function
	err = f()
	if err != nil {
		return
	}

	return
}

func (e EtcdState) CreateExhibit(ctx context.Context, app domain.Exhibit) error {
	key := "/" + e.Config.GetEtcdBaseKey() + "/" + app.Id + "/" + "object"

	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("etcd persistence").
		Start(ctx, "CreateExhibit", trace.WithAttributes(attribute.String("key", key), attribute.String("id", app.Id)))
	defer span.End()

	span.AddEvent("checked if exhibit already exists")

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
	key := "/" + e.Config.GetEtcdBaseKey() + "/" + id + "/" + "object"

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
		if !strings.HasSuffix(string(kv.Key), "/object") {
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

func (e EtcdState) UpdateExhibit(ctx context.Context, app domain.Exhibit) error {
	key := "/" + e.Config.GetEtcdBaseKey() + "/" + app.Id + "/" + "object"

	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("etcd persistence").
		Start(ctx, "UpdateExhibit", trace.WithAttributes(attribute.String("key", key), attribute.String("id", app.Id)))
	defer span.End()

	b, err := json.Marshal(app)
	if err != nil {
		return err
	}

	_, err = e.Client.Put(subCtx, key, string(b))
	if err != nil {
		return err
	}

	span.AddEvent("updated exhibit in etcd")

	return nil
}

func (e EtcdState) DeleteExhibitById(ctx context.Context, id string) error {
	//TODO implement me
	panic("implement me")
}
