package api

import (
	"encoding/json"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"io"
	"museum/domain"
	"museum/http/router"
	"museum/service"
	"net/http"
)

func getExhibits(exhibitService service.ExhibitService, log *zap.SugaredLogger) router.MuxHandlerFunc {
	return func(res *router.Response, req *router.Request) {
		exhibits := exhibitService.GetExhibits()
		err := res.WriteJson(exhibits)
		if err != nil {
			log.Warnw("error writing json", "error", err, "requestId", req.RequestID)
		}
	}
}

func getExhibitById(exhibitService service.ExhibitService, log *zap.SugaredLogger) router.MuxHandlerFunc {
	return func(res *router.Response, req *router.Request) {
		exhibitId := req.Params["id"]
		exhibit, err := exhibitService.GetExhibitById(exhibitId)
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			err := res.WriteJson(map[string]string{"status": "Not Found"})
			if err != nil {
				log.Warnw("error writing json", "error", err, "requestId", req.RequestID)
			}
		}

		err = res.WriteJson(exhibit)
		if err != nil {
			log.Warnw("error writing json", "error", err, "requestId", req.RequestID)
		}
	}
}

func createExhibit(exhibitService service.ExhibitService, log *zap.SugaredLogger, provider trace.TracerProvider) router.MuxHandlerFunc {
	return func(res *router.Response, req *router.Request) {
		ctx, span := provider.
			Tracer("API request").
			Start(req.Context(), "HTTP POST /api/exhibits")
		defer span.End()

		body, err := io.ReadAll(req.Body)
		if err != nil {
			return
		}

		exhibit := &domain.Exhibit{}
		err = json.Unmarshal(body, exhibit)
		if err != nil {
			log.Warnw("error unmarshalling json", "error", err, "requestId", req.RequestID)
			return
		}

		span.AddEvent("request read")

		err, id := exhibitService.CreateExhibit(ctx, domain.CreateExhibit{
			Exhibit:   *exhibit,
			RequestID: req.RequestID,
		})
		if err != nil {
			log.Errorw("error creating exhibit", "error", err, "requestId", req.RequestID)
			res.WriteHeader(http.StatusInternalServerError)
		}

		res.WriteHeader(http.StatusCreated)
		err = res.WriteJson(map[string]string{"status": "Created", "id": id})
		if err != nil {
			log.Warnw("error writing json", "error", err, "requestId", req.RequestID)
		}

		span.AddEvent("response written")
	}
}

func RegisterRoutes(r *router.Mux, exhibitService service.ExhibitService, log *zap.SugaredLogger, provider trace.TracerProvider) {
	r.AddRoute(router.Get("/api/exhibits", getExhibits(exhibitService, log)))
	r.AddRoute(router.Get("/api/exhibits/{id}", getExhibitById(exhibitService, log)))
	r.AddRoute(router.Post("/api/exhibits", createExhibit(exhibitService, log, provider)))
}
