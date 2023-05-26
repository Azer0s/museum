package impl

import (
	"context"
	"encoding/json"
	"errors"
	etcd "go.etcd.io/etcd/client/v3"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"museum/domain"
	"strings"
)

func (e *EtcdState) CreateExhibit(ctx context.Context, app domain.Exhibit) error {
	e.ExhibitCacheMu.Lock()
	defer e.ExhibitCacheMu.Unlock()

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

	if e.ExhibitCache != nil {
		e.ExhibitCache[app.Id] = app
	}

	return nil
}

func (e *EtcdState) GetExhibitById(ctx context.Context, id string) (domain.Exhibit, error) {
	if e.ExhibitCache != nil {
		e.ExhibitCacheMu.RLock()
		defer e.ExhibitCacheMu.RUnlock()

		if exhibit, ok := e.ExhibitCache[id]; ok {
			return exhibit, nil
		}
	}

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

	if resp.Count == 0 {
		return domain.Exhibit{}, errors.New("exhibit with id " + id + " not found")
	}

	span.AddEvent("found exhibit")
	var exhibit domain.Exhibit
	err = json.Unmarshal(resp.Kvs[0].Value, &exhibit)
	if err != nil {
		return domain.Exhibit{}, err
	}

	return exhibit, nil
}

func (e *EtcdState) GetAllExhibits(ctx context.Context) []domain.Exhibit {
	var exhibits []domain.Exhibit

	if e.ExhibitCache != nil {
		e.ExhibitCacheMu.RLock()
		defer e.ExhibitCacheMu.RUnlock()

		for _, exhibit := range e.ExhibitCache {
			exhibits = append(exhibits, exhibit)
		}

		return exhibits
	}

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

func (e *EtcdState) DeleteExhibitById(ctx context.Context, id string) error {
	e.ExhibitCacheMu.Lock()
	defer e.ExhibitCacheMu.Unlock()

	key := "/" + e.Config.GetEtcdBaseKey() + "/" + id + "/" + "meta"

	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("etcd persistence").
		Start(ctx, "DeleteExhibitById", trace.WithAttributes(attribute.String("key", key), attribute.String("id", id)))
	defer span.End()

	span.AddEvent("deleting exhibit")

	_, err := e.Client.Delete(subCtx, key)
	if err != nil {
		return err
	}

	if e.ExhibitCache != nil {
		delete(e.ExhibitCache, id)
	}

	return nil
}
