package service

import (
	"museum/domain"
)

type EnvironmentTemplateResolverService interface {
	FillEnvironmentTemplate(exhibit *domain.Exhibit, o domain.Object, templateContainer *map[string]string) (error, map[string]string)
}
