package impl

import (
	"context"
	"encoding/json"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"museum/domain"
)

func (e *EtcdState) SetRuntimeInfo(ctx context.Context, id string, runtimeInfo domain.ExhibitRuntimeInfo) error {
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

	e.RuntimeInfoCacheMu.Lock()
	defer e.RuntimeInfoCacheMu.Unlock()

	if e.RuntimeInfoCache != nil {
		e.RuntimeInfoCache[id] = runtimeInfo
	}

	return nil
}

func (e *EtcdState) GetRuntimeInfo(ctx context.Context, id string) (domain.ExhibitRuntimeInfo, error) {
	if e.RuntimeInfoCache != nil {
		e.RuntimeInfoCacheMu.RLock()
		defer e.RuntimeInfoCacheMu.RUnlock()

		if runtimeInfo, ok := e.RuntimeInfoCache[id]; ok {
			return runtimeInfo, nil
		}
	}

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

func (e *EtcdState) DeleteRuntimeInfo(ctx context.Context, id string) error {
	key := "/" + e.Config.GetEtcdBaseKey() + "/" + id + "/" + "runtime_info"

	// create new trace span for event service
	subCtx, span := e.Provider.
		Tracer("etcd persistence").
		Start(ctx, "DeleteRuntimeInfo", trace.WithAttributes(attribute.String("key", key), attribute.String("id", id)))
	defer span.End()

	span.AddEvent("searching for runtime info for exhibit")

	_, err := e.Client.Delete(subCtx, key)
	if err != nil {
		return err
	}

	return nil
}
