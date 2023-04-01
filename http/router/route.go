package router

import (
	"museum/http/router/path"
	"net/http"
)

type MuxHandlerFunc func(*Response, *Request)

type Route struct {
	Path    path.Path
	Handler MuxHandlerFunc
	Method  string
}

func Get(p string, handler MuxHandlerFunc) Route {
	return Route{
		Path:    path.ConstructPath(p),
		Handler: handler,
		Method:  http.MethodGet,
	}
}

func Post(p string, handler MuxHandlerFunc) Route {
	return Route{
		Path:    path.ConstructPath(p),
		Handler: handler,
		Method:  http.MethodPost,
	}
}
