package service

import (
	"museum/domain"
	"museum/http/router"
)

type ApplicationProxyService interface {
	ForwardRequest(exhibit domain.Exhibit, path string, res *router.Response, req *router.Request) error
}
