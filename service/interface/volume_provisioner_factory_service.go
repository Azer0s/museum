package service

type VolumeProvisionerFactoryService interface {
	GetForDriverType(driver string) (VolumeProvisionerService, error)
}
