package impl

import (
	"context"
	"museum/persistence"
	"museum/util"
)

type LockServiceImpl struct {
	State persistence.State
}

func (l LockServiceImpl) GetRwLock(ctx context.Context, id string, lockName string) util.RwErrMutex {
	return l.State.GetRwLock(ctx, id, lockName)
}
