package service

import (
	"museum/domain"
	"museum/http"
)

type ApplicationProxyService interface {
	ForwardRequest(exhibit domain.Exhibit, path string, res *http.Response, req *http.Request) error
}
