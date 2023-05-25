package service

import (
	"context"
	"museum/domain"
)

type RuntimeInfoService interface {
	SetRuntimeInfo(ctx context.Context, id string, runtimeInfo domain.ExhibitRuntimeInfo) error
	GetRuntimeInfo(ctx context.Context, id string) (domain.ExhibitRuntimeInfo, error)
}
