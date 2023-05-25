package impl

import (
	"context"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"strconv"
)

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

func (e EtcdState) DeleteLastAccessed(ctx context.Context, id string) error {
	key := "/" + e.Config.GetEtcdBaseKey() + "/" + id + "/" + "last_accessed"

	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("etcd persistence").
		Start(ctx, "DeleteLastAccessed", trace.WithAttributes(attribute.String("key", key), attribute.String("id", id)))
	defer span.End()

	_, err := e.Client.Delete(subCtx, key)
	if err != nil {
		return err
	}

	span.AddEvent("deleted last_accessed time for exhibit")

	return nil
}
