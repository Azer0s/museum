package http

import (
	"go.opentelemetry.io/otel/trace"
	"museum/config"
	"net/http"
)

func HealthHandler(_ config.Config, _ trace.TracerProvider) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	}
}
