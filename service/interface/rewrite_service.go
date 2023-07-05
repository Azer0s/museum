package service

import (
	"museum/domain"
	"museum/http"
	gohttp "net/http"
)

type RewriteService interface {
	RewriteServerResponse(exhibit domain.Exhibit, ip string, res *gohttp.Response, body *[]byte) error
	RewriteClientRequest(exhibit domain.Exhibit, ip string, req *http.Request, body *[]byte) error
}
