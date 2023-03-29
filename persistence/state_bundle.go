package persistence

import (
	"museum/domain"
)

type StateBundle struct {
	SharedPersistentState SharedPersistentState
	Emitter               Emitter
	Consumer              Consumer
}

func (s StateBundle) GetApplications() ([]domain.Application, error) {
	//TODO implement me
	panic("implement me")
}

func (s StateBundle) AddApplication(app domain.Application) error {
	//TODO implement me
	panic("implement me")
}

func (s StateBundle) RenewApplicationLease(app domain.Application) error {
	//TODO implement me
	panic("implement me")
}

func (s StateBundle) ExpireApplicationLease(app domain.Application) error {
	//TODO implement me
	panic("implement me")
}
