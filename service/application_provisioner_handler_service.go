package service

import (
	"museum/service/impl"
	service "museum/service/interface"
)

type ApplicationProvisionerHandlerService service.ApplicationProvisionerHandlerService

func NewApplicationProvisionerHandlerService(applicationProvisionerService service.ApplicationProvisionerService) ApplicationProvisionerHandlerService {
	return &impl.ApplicationProvisionerHandlerServiceImpl{
		ApplicationProvisionerService: applicationProvisionerService,
	}
}
