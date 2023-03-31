package health

import (
	_ "embed"
	"museum/http/router"
	"museum/http/router/path"
	"net/http"
)

func healthEndpoint(w http.ResponseWriter, r *http.Request, _ map[string]string) {
	w.WriteHeader(http.StatusOK)
	err := router.WriteStatus(w, router.Status{Status: "OK"})
	if err != nil {
		//TODO: log error
		return
	}
}

func RegisterRoutes(router *router.Mux) {
	router.AddRoute(path.Get("/api/health", healthEndpoint))
}
