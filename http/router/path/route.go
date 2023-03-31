package path

import "net/http"

type MuxHandlerFunc func(http.ResponseWriter, *http.Request, map[string]string)

type Route struct {
	Path    path
	Handler MuxHandlerFunc
	Method  string
}

func Get(path string, handler MuxHandlerFunc) Route {
	return Route{
		Path:    constructPath(path),
		Handler: handler,
		Method:  http.MethodGet,
	}
}
