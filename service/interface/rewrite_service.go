package service

import (
	"museum/domain"
	"museum/http"
	gohttp "net/http"
)

type RewriteService interface {
	RewriteServerResponse(exhibit domain.Exhibit, hostname string, res *gohttp.Response, body *[]byte) (*[]byte, error)
	RewriteClientRequest(exhibit domain.Exhibit, hostname string, req *http.Request, body *[]byte) error
}
