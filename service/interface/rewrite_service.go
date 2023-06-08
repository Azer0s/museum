package service

import (
	"museum/domain"
	"museum/http"
	gohttp "net/http"
)

type RewriteService interface {
	RewriteServerResponse(exhibit domain.Exhibit, res *gohttp.Response, body *[]byte) error
	RewriteClientRequest(exhibit domain.Exhibit, req *http.Request, body *[]byte) error
}
