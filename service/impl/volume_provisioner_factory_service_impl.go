package impl

import (
	"errors"
	service "museum/service/interface"
)

type VolumeProvisionerFactoryServiceImpl struct {
}

func (v VolumeProvisionerFactoryServiceImpl) GetForDriverType(driver string) (service.VolumeProvisionerService, error) {
	switch driver {
	case "local":
		return &LocalVolumeProvisionerService{}, nil
	default:
		return nil, errors.New("unsupported driver type")
	}
}
