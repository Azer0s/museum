package impl

import (
	"context"
	"museum/persistence"
)

type LastAccessedServiceImpl struct {
	State persistence.State
}

func (e LastAccessedServiceImpl) GetLastAccessed(ctx context.Context, id string) (int64, error) {
	return e.State.GetLastAccessed(ctx, id)
}

func (e LastAccessedServiceImpl) SetLastAccessed(ctx context.Context, id string, lastAccessed int64) error {
	return e.State.SetLastAccessed(ctx, id, lastAccessed)
}
