package service

import (
	"museum/domain"
	gohttp "net/http"
)

type RewriteService interface {
	RewriteRequest(exhibit domain.Exhibit, proxyRes *gohttp.Response, body *[]byte) error
}
