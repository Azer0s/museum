package router

import (
	"encoding/json"
	"museum/http/router/path"
	"net/http"
)

type Mux struct {
	routes []path.Route
	mux    *http.ServeMux
}

type Status struct {
	Status string `json:"status"`
}

func WriteStatus(writer http.ResponseWriter, status Status) error {
	writer.Header().Set("Content-Type", "application/json")
	b, err := json.Marshal(status)
	if err != nil {
		return err
	}

	_, err = writer.Write(b)
	return err
}

func (r *Mux) ServeHTTP(writer http.ResponseWriter, request *http.Request) {
	for _, route := range r.routes {
		if segments, ok := route.Path.Match(request.URL.Path); ok {
			if route.Method == request.Method {
				pathParams := make(map[string]string)
				for _, segment := range segments {
					if w, ok := segment.(*path.WildcardPathSegment); ok {
						pathParams[w.VariableName] = w.Value
					}
				}

				route.Handler(writer, request, pathParams)
				return
			}
		}
	}

	writer.WriteHeader(http.StatusNotFound)
	err := WriteStatus(writer, Status{Status: "Not Found"})
	if err != nil {
		//TODO: log error
		return
	}
}

func NewMux() *Mux {
	return &Mux{}
}

func (r *Mux) AddRoute(route path.Route) {
	r.routes = append(r.routes, route)
}
