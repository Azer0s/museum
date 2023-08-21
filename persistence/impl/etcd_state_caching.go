package impl

import (
	"context"
	"encoding/json"
	etcd "go.etcd.io/etcd/client/v3"
	"museum/domain"
	"strings"
)

func (e *EtcdState) handleExhibitEvent(exhibitId string, w etcd.WatchChan) (deleted bool) {
	events := <-w

	e.ExhibitCacheMu.Lock()
	defer e.ExhibitCacheMu.Unlock()

	for _, event := range events.Events {
		if strings.Contains(string(event.Kv.Key), "lock") {
			continue
		}

		e.Log.Infow("received event for exhibit", "exhibitId", exhibitId, "event", event.Type.String())

		if event.Type == etcd.EventTypeDelete {
			e.Log.Debugw("exhibit deleted", "exhibitId", exhibitId)
			delete(e.ExhibitCache, exhibitId)
			return true
		}

		updatedExhibit := domain.Exhibit{}
		err := json.Unmarshal(event.Kv.Value, &updatedExhibit)
		if err != nil {
			e.Log.Errorw("error unmarshalling exhibit", "error", err)
			continue
		}

		e.Log.Debugw("exhibit updated", "exhibitId", exhibitId)
		e.ExhibitCache[exhibitId] = updatedExhibit
	}

	return false
}

func (e *EtcdState) handleRuntimeInfoEvent(w etcd.WatchChan) bool {
	events := <-w

	e.RuntimeInfoCacheMu.Lock()
	defer e.RuntimeInfoCacheMu.Unlock()

	for _, event := range events.Events {
		exhibitId := strings.Split(string(event.Kv.Key), "/")[1]

		e.Log.Infow("received runtime_info event for exhibit", "exhibitId", exhibitId, "event", event.Type.String())
		if event.Type == etcd.EventTypeDelete {
			e.Log.Debugw("runtime_info deleted", "exhibitId", exhibitId)
			delete(e.RuntimeInfoCache, exhibitId)
			return true
		}

		updatedRuntimeInfo := domain.ExhibitRuntimeInfo{}
		err := json.Unmarshal(event.Kv.Value, &updatedRuntimeInfo)
		if err != nil {
			e.Log.Errorw("error unmarshalling exhibit runtime info", "error", err)
			continue
		}

		e.Log.Debugw("runtime_info updated", "exhibitId", exhibitId)
		e.RuntimeInfoCache[exhibitId] = updatedRuntimeInfo
	}

	return false
}

func (e *EtcdState) handleCreateEvent(w etcd.WatchChan) {
	events := <-w

	for _, event := range events.Events {
		if !event.IsCreate() {
			continue
		}

		if strings.HasSuffix(string(event.Kv.Key), "meta") {
			e.handleNewExhibit(event)
		}

		if strings.HasSuffix(string(event.Kv.Key), "runtime_info") {
			e.handleNewRuntimeInfo(event)
		}
	}
}

func (e *EtcdState) handleNewRuntimeInfo(event *etcd.Event) {
	e.RuntimeInfoCacheMu.Lock()
	defer e.RuntimeInfoCacheMu.Unlock()

	newRuntimeInfo := domain.ExhibitRuntimeInfo{}
	err := json.Unmarshal(event.Kv.Value, &newRuntimeInfo)
	if err != nil {
		e.Log.Errorw("error unmarshalling exhibit runtime info", "error", err)
		return
	}

	exhibitId := strings.Split(string(event.Kv.Key), "/")[1]

	e.Log.Debugw("new exhibit runtime info created", "exhibitId", exhibitId)
	e.RuntimeInfoCache[exhibitId] = newRuntimeInfo
	e.watchRuntimeInfo(exhibitId)
	return
}

func (e *EtcdState) handleNewExhibit(event *etcd.Event) {
	e.ExhibitCacheMu.Lock()
	defer e.ExhibitCacheMu.Unlock()

	newExhibit := domain.Exhibit{}
	err := json.Unmarshal(event.Kv.Value, &newExhibit)
	if err != nil {
		e.Log.Errorw("error unmarshalling exhibit", "error", err)
		return
	}

	e.Log.Debugw("new exhibit created", "exhibitId", newExhibit.Id)
	e.ExhibitCache[newExhibit.Id] = newExhibit
	e.watchRuntimeInfo(newExhibit.Id)
	e.watchExhibit(newExhibit.Id)
	return
}

func (e *EtcdState) watchExhibit(exhibitId string) {
	w := e.Client.Watch(context.Background(), "/"+e.Config.GetEtcdBaseKey()+"/"+exhibitId, etcd.WithPrefix())
	go func(exhibitId string, w etcd.WatchChan) {
		for {
			deleted := e.handleExhibitEvent(exhibitId, w)
			if deleted {
				return
			}
		}
	}(exhibitId, w)
}

func (e *EtcdState) watchRuntimeInfo(exhibitId string) {
	w := e.Client.Watch(context.Background(), "/"+e.Config.GetEtcdBaseKey()+"/"+exhibitId+"/"+"runtime_info")
	go func(exhibitId string, w etcd.WatchChan) {
		for {
			deleted := e.handleRuntimeInfoEvent(w)
			if deleted {
				return
			}
		}
	}(exhibitId, w)
}
