package persistence

import (
	"context"
	"museum/domain"
)

type SharedPersistentState interface {
	WithLock(f func() error) (err error)
	GetExhibits() ([]domain.Exhibit, error)
	DeleteExhibitById(id string) error
	AddExhibit(ctx context.Context, app domain.Exhibit) error
}

type SharedPersistentEmittedState interface {
	GetExhibits() []domain.Exhibit
	GetExhibitById(id string) (*domain.Exhibit, error)
	EventReceived(eventId string) (<-chan struct{}, error)
	AddExhibit(ctx context.Context, app domain.CreateExhibit) error
	RenewExhibitLeaseById(id string) error
	ExpireExhibitLease(id string) error
}
