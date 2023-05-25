package service

import (
	"museum/persistence"
	"museum/service/impl"
	service "museum/service/interface"
)

type RuntimeInfoService service.RuntimeInfoService

func NewRuntimeInfoService(state persistence.State) RuntimeInfoService {
	return &impl.RuntimeInfoServiceImpl{
		State: state,
	}
}
