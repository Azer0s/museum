package impl

import (
	"context"
	"museum/domain"
	"museum/persistence"
	service "museum/service/interface"
	"museum/util"
)

type RuntimeInfoServiceImpl struct {
	State       persistence.State
	LockService service.LockService
}

func (r RuntimeInfoServiceImpl) SetRuntimeInfo(ctx context.Context, id string, runtimeInfo domain.ExhibitRuntimeInfo) error {
	return r.State.SetRuntimeInfo(ctx, id, runtimeInfo)
}

func (r RuntimeInfoServiceImpl) GetRuntimeInfo(ctx context.Context, id string) (ri domain.ExhibitRuntimeInfo, err error) {
	lock := r.LockService.GetRwLock(ctx, id, "runtime_info")
	err = lock.RLock()
	if err != nil {
		return
	}

	//TODO: span

	defer func(lock util.RwErrMutex) {
		e := lock.RUnlock()
		if e != nil {
			err = e
		}
	}(lock)

	return r.State.GetRuntimeInfo(ctx, id)
}
