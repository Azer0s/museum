package health

import (
	_ "embed"
	"museum/http/router"
	"net/http"
)

func healthEndpoint(res *router.Response, _ *router.Request) {
	res.WriteHeader(http.StatusOK)
	err := router.WriteStatus(res, router.Status{Status: "OK"})
	if err != nil {
		//TODO: log error
		return
	}
}

func RegisterRoutes(r *router.Mux) {
	r.AddRoute(router.Get("/api/health", healthEndpoint))
}
