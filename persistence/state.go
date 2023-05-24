package persistence

import (
	"context"
	"museum/domain"
)

// State handles persisting state to disk
// it does not care about the state an application is in, it is just responsible for
// communication between museum instances. No business logic shall be contained here.
type State interface {
	WithLock(ctx context.Context, id string, f func() error) (err error)

	CreateExhibit(ctx context.Context, app domain.Exhibit) error
	GetExhibitById(ctx context.Context, id string) (domain.Exhibit, error)
	GetAllExhibits(ctx context.Context) []domain.Exhibit
	UpdateExhibit(ctx context.Context, app domain.Exhibit) error
	DeleteExhibitById(ctx context.Context, id string) error

	GetLastAccessed(ctx context.Context, id string) (int64, error)
	SetLastAccessed(ctx context.Context, id string, lastAccessed int64) error
}
