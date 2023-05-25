package service

import (
	"museum/persistence"
	"museum/service/impl"
	service "museum/service/interface"
)

type LockService service.LockService

func NewLockService(state persistence.State) LockService {
	return &impl.LockServiceImpl{
		State: state,
	}
}
