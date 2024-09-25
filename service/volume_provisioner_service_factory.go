package service

import (
	"museum/service/impl"
	service "museum/service/interface"
)

type VolumeProvisionerFactoryService service.VolumeProvisionerFactoryService

func NewVolumeProvisionerFactoryService() VolumeProvisionerFactoryService {
	return &impl.VolumeProvisionerFactoryServiceImpl{}
}
