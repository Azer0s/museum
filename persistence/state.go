package persistence

import (
	"context"
	"museum/domain"
	"museum/util"
)

// State handles persisting state to disk
// it does not care about the state an application is in, it is just responsible for
// communication between museum instances. No business logic shall be contained here.
type State interface {
	GetRwLock(ctx context.Context, id string, lockName string) util.RwErrMutex

	CreateExhibit(ctx context.Context, app domain.Exhibit) error
	GetExhibitById(ctx context.Context, id string) (domain.Exhibit, error)
	GetAllExhibits(ctx context.Context) []domain.Exhibit
	DeleteExhibitById(ctx context.Context, id string) error

	SetRuntimeInfo(ctx context.Context, id string, runtimeInfo domain.ExhibitRuntimeInfo) error
	GetRuntimeInfo(ctx context.Context, id string) (domain.ExhibitRuntimeInfo, error)

	GetLastAccessed(ctx context.Context, id string) (int64, error)
	SetLastAccessed(ctx context.Context, id string, lastAccessed int64) error
}
