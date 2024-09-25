package service

import "museum/domain"

type VolumeProvisionerService interface {
	CheckValidity(config domain.StringMap) error
}
