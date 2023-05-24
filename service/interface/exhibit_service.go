package service

import (
	"context"
	"museum/domain"
)

type ExhibitService interface {
	CreateExhibit(ctx context.Context, createExhibit domain.CreateExhibit) (string, error)
}
