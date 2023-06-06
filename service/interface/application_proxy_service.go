package service

import (
	"museum/http/router"
)

type ApplicationProxyService interface {
	ForwardRequest(exhibitId string, path string, res *router.Response, req *router.Request) error
}
