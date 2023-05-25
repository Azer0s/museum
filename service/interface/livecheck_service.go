package service

import "museum/domain"

type Livecheck interface {
	Check(exhibit domain.Exhibit, object domain.Object) (retry bool, err error)
}
