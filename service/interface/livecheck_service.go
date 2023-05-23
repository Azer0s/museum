package service

import "museum/domain"

type LivecheckService interface {
	Check(exhibit domain.Exhibit, object domain.Object) (retry bool, err error)
}
