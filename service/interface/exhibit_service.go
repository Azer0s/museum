package service

import (
	"context"
	"museum/domain"
)

type ExhibitService interface {
	WithLock(ctx context.Context, id string, f func() error) (err error)

	GetExhibitById(ctx context.Context, id string) (domain.Exhibit, error)
	GetAllExhibits(ctx context.Context) []domain.Exhibit
	CreateExhibit(ctx context.Context, createExhibit domain.CreateExhibit) (string, error)
	UpdateExhibit(ctx context.Context, app domain.Exhibit) error
	DeleteExhibitById(ctx context.Context, id string) error

	GetLastAccessed(ctx context.Context, id string) (int64, error)
	SetLastAccessed(ctx context.Context, id string, lastAccessed int64) error
}
