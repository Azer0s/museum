package http

import (
	"encoding/json"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"museum/http/path"
	"net/http"
	"strings"
)

type FallbackHandler func(writer http.ResponseWriter, request *Request) error

type Mux struct {
	routes     []Route
	mux        *http.ServeMux
	log        *zap.SugaredLogger
	fallbackFn *FallbackHandler
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
	requestId := uuid.New().String()
	r.log.Debugw("request received", "method", request.Method, "path", request.URL.Path, "requestId", requestId)

	for _, route := range r.routes {
		if segments, ok := route.Path.Match(request.URL.Path); ok {
			if route.Method == request.Method || route.Method == "*" {
				restPath := ""

				pathParams := make(map[string]string)
				for _, segment := range segments {
					if w, ok := segment.(*path.WildcardPathSegment); ok {
						pathParams[w.VariableName] = w.Value
					}

					if w, ok := segment.(*path.RestPathSegment); ok {
						restPath = w.Value
					}
				}

				queryParams := strings.Split(request.URL.RequestURI(), "?")
				rawQueryParams := ""
				if len(queryParams) > 1 {
					rawQueryParams = queryParams[1]
				}

				route.Handler(&Response{writer}, &Request{
					Request:        request,
					Params:         pathParams,
					RequestID:      requestId,
					RestPath:       restPath,
					RawQueryParams: rawQueryParams,
				})
				return
			}
		}
	}

	if r.fallbackFn != nil {
		err := (*r.fallbackFn)(writer, &Request{
			Request:   request,
			RequestID: requestId,
		})

		if err == nil {
			return
		}

		r.log.Warnw("error in not found handler", "error", err, "requestId", requestId)
	}

	writer.WriteHeader(http.StatusNotFound)
	err := WriteStatus(writer, Status{Status: "Not Found"})
	if err == nil {
		r.log.Warnw("no route found", "method", request.Method, "path", request.URL.Path, "requestId", requestId)
		return
	}
}

func NewMux(log *zap.SugaredLogger) *Mux {
	return &Mux{
		log: log,
	}
}

func (r *Mux) AddRoute(route Route) {
	r.routes = append(r.routes, route)
}

func (r *Mux) SetFallbackHandler(fallbackHandler FallbackHandler) {
	r.fallbackFn = &fallbackHandler
}
