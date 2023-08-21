package service

import (
	"context"
	"museum/domain"
)

type ExhibitService interface {
	GetExhibitById(ctx context.Context, id string) (domain.Exhibit, error)
	GetAllExhibits(ctx context.Context) []domain.Exhibit
	CreateExhibit(ctx context.Context, createExhibit domain.CreateExhibit) (string, error)
	DeleteExhibitById(ctx context.Context, id string) error
	Count() int
}
