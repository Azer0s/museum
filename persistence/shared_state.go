package persistence

import (
	"museum/domain"
)

type SharedPersistentState interface {
	GetApplications() ([]domain.Application, error)
	DeleteApplication(app domain.Application) error
	AddApplication(app domain.Application) error
}

type SharedPersistentEmittedState interface {
	GetApplications() ([]domain.Application, error)
	AddApplication(app domain.Application) error
	RenewApplicationLease(app domain.Application) error
	ExpireApplicationLease(app domain.Application) error
}
