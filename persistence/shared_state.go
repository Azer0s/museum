package persistence

import (
	"museum/domain"
)

type SharedPersistentState interface {
	GetApplications() ([]domain.Application, error)
	DeleteApplication(app domain.Application) error
	AddApplication(app domain.Application) error
}
