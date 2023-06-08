package health

import (
	_ "embed"
	"go.uber.org/zap"
	"museum/http"
	gohttp "net/http"
)

func healthEndpoint(log *zap.SugaredLogger) func(res *http.Response, req *http.Request) {
	return func(res *http.Response, req *http.Request) {
		log.Debug("handling health endpoint request", "requestId", req.RequestID)

		res.WriteHeader(gohttp.StatusOK)
		err := http.WriteStatus(res, http.Status{Status: "OK"})
		if err != nil {
			log.Errorw("error writing status", "error", err, "requestId", req.RequestID)
			return
		}
	}
}

func RegisterRoutes(r *http.Mux, log *zap.SugaredLogger) {
	r.AddRoute(http.Get("/api/health", healthEndpoint(log)))
}
