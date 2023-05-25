package service

import (
	"museum/persistence"
	"museum/service/impl"
	service "museum/service/interface"
)

type LastAccessedService service.LastAccessedService

func NewLastAccessedService(state persistence.State) LastAccessedService {
	return &impl.LastAccessedServiceImpl{
		State: state,
	}
}
