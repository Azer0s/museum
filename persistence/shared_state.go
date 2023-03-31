package persistence

import (
	"museum/domain"
)

type SharedPersistentState interface {
	WithLock(f func() error) (err error)
	GetExhibits() ([]domain.Exhibit, error)
	DeleteExhibit(app domain.Exhibit) error
	AddExhibit(app domain.Exhibit) error
}

type SharedPersistentEmittedState interface {
	GetExhibits() []domain.Exhibit
	AddExhibit(app domain.Exhibit) error
	RenewExhibitLease(app domain.Exhibit) error
	ExpireExhibitLease(app domain.Exhibit) error
}
