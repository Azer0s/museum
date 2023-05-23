package service

import "museum/domain"

type LivecheckService interface {
	Check(object domain.Object) error
}
