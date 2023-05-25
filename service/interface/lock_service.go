package service

import (
	"context"
	"museum/util"
)

type LockService interface {
	GetRwLock(ctx context.Context, id string, lockName string) util.RwErrMutex
}
