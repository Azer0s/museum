package service

import (
	"context"
)

type LastAccessedService interface {
	GetLastAccessed(ctx context.Context, id string) (int64, error)
	SetLastAccessed(ctx context.Context, id string, lastAccessed int64) error
}
