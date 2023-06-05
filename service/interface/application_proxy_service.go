package service

import (
	"museum/http/router"
)

type ApplicationProxyService interface {
	ForwardRequest(exhibitId string, res *router.Response, req *router.Request) error
}
