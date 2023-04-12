package service

import (
	"context"
	"museum/domain"
)

type ExhibitService interface {
	GetExhibits() []domain.Exhibit
	GetExhibitById(id string) (*domain.Exhibit, error)
	CreateExhibit(ctx context.Context, createExhibit domain.CreateExhibit) (string, error)
}
