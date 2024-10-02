package service

import (
	"context"
	"museum/domain"
)

type VolumeProvisionerService interface {
	CheckValidity(config domain.StringMap) error
	ProvisionStorage(ctx context.Context, config domain.StringMap) (string, error)
	DeprovisionStorage(ctx context.Context, config domain.StringMap) error
}
