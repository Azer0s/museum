package health

import (
	_ "embed"
	"go.uber.org/zap"
	http2 "museum/http"
	"net/http"
)

func healthEndpoint(log *zap.SugaredLogger) func(res *http2.Response, req *http2.Request) {
	return func(res *http2.Response, req *http2.Request) {
		log.Debug("handling health endpoint request", "requestId", req.RequestID)

		res.WriteHeader(http.StatusOK)
		err := http2.WriteStatus(res, http2.Status{Status: "OK"})
		if err != nil {
			log.Errorw("error writing status", "error", err, "requestId", req.RequestID)
			return
		}
	}
}

func RegisterRoutes(r *http2.Mux, log *zap.SugaredLogger) {
	r.AddRoute(http2.Get("/api/health", healthEndpoint(log)))
}
