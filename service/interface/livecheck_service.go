package service

import (
	"context"
	"museum/domain"
)

type Livecheck interface {
	Check(ctx context.Context, exhibit domain.Exhibit, object domain.Object) (retry bool, err error)
}
