package health

import (
	_ "embed"
	"go.uber.org/zap"
	"museum/http/router"
	"net/http"
)

func healthEndpoint(log *zap.SugaredLogger) func(res *router.Response, req *router.Request) {
	return func(res *router.Response, req *router.Request) {
		log.Debug("handling health endpoint request", "requestId", req.RequestID)

		res.WriteHeader(http.StatusOK)
		err := router.WriteStatus(res, router.Status{Status: "OK"})
		if err != nil {
			log.Errorw("error writing status", "error", err, "requestId", req.RequestID)
			return
		}
	}
}

func RegisterRoutes(r *router.Mux, log *zap.SugaredLogger) {
	r.AddRoute(router.Get("/api/health", healthEndpoint(log)))
}
