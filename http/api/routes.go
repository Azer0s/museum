package api

import (
	"encoding/json"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"io"
	"museum/domain"
	"museum/http/router"
	"museum/service"
	"net/http"
)

func getExhibits(exhibitService service.ExhibitService, log *zap.SugaredLogger, provider trace.TracerProvider) router.MuxHandlerFunc {
	return func(res *router.Response, req *router.Request) {
		_, span := provider.
			Tracer("API request").
			Start(req.Context(), "HTTP GET /exhibits/", trace.WithAttributes(attribute.String("requestId", req.RequestID)))
		defer span.End()

		exhibits := exhibitService.GetExhibits()
		err := res.WriteJson(exhibits)
		if err != nil {
			log.Warnw("error writing json", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
		}
	}
}

func getExhibitById(exhibitService service.ExhibitService, log *zap.SugaredLogger, provider trace.TracerProvider) router.MuxHandlerFunc {
	return func(res *router.Response, req *router.Request) {
		_, span := provider.
			Tracer("API request").
			Start(req.Context(), "HTTP GET /exhibits/"+req.Params["id"], trace.WithAttributes(attribute.String("requestId", req.RequestID)))
		defer span.End()

		exhibitId := req.Params["id"]
		exhibit, err := exhibitService.GetExhibitById(exhibitId)
		if err != nil {
			res.WriteErr(err)
			res.WriteHeader(http.StatusNotFound)
			return
		}

		err = res.WriteJson(exhibit)
		if err != nil {
			log.Warnw("error writing json", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
		}
	}
}

func createExhibit(exhibitService service.ExhibitService, log *zap.SugaredLogger, provider trace.TracerProvider) router.MuxHandlerFunc {
	return func(res *router.Response, req *router.Request) {
		ctx, span := provider.
			Tracer("API request").
			Start(req.Context(), "HTTP POST /api/exhibits", trace.WithAttributes(attribute.String("requestId", req.RequestID)))
		defer span.End()

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Warnw("error reading request body", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
			return
		}

		exhibit := &domain.Exhibit{}
		err = json.Unmarshal(body, exhibit)
		if err != nil {
			log.Warnw("error unmarshalling json", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
			return
		}

		span.AddEvent("request read")

		err, id := exhibitService.CreateExhibit(ctx, domain.CreateExhibit{
			Exhibit:   *exhibit,
			RequestID: req.RequestID,
		})
		if err != nil {
			log.Errorw("error creating exhibit", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
			return
		}

		res.WriteHeader(http.StatusCreated)
		err = res.WriteJson(map[string]string{"status": "Created", "id": id})
		if err != nil {
			log.Warnw("error writing json", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
			return
		}

		span.AddEvent("response written")
	}
}

type Event struct {
	ExhibitId string `json:"exhibitId"`
	EventType string `json:"eventType"`
}

func handleEvents(exhibitService service.ExhibitService, provisionerService service.ApplicationProvisionerService, log *zap.SugaredLogger, provider trace.TracerProvider) router.MuxHandlerFunc {
	return func(res *router.Response, req *router.Request) {
		ctx, span := provider.
			Tracer("API request").
			Start(req.Context(), "HTTP POST /api/events", trace.WithAttributes(attribute.String("requestId", req.RequestID)))
		defer span.End()

		body, err := io.ReadAll(req.Body)
		if err != nil {
			log.Warnw("error reading request body", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
			return
		}

		event := &Event{}
		err = json.Unmarshal(body, event)
		if err != nil {
			log.Warnw("error unmarshalling json", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
			return
		}

		span.AddEvent("request read")

		//TODO: move this to a service, controller should only be responsible for handling the request
		err = provisionerService.StartApplication(ctx, event.ExhibitId)
		if err != nil {
			log.Errorw("error starting application", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
			return
		}

		span.AddEvent("application started")

		res.WriteHeader(http.StatusCreated)
		err = res.WriteJson(map[string]string{"status": "Created"})
		if err != nil {
			log.Warnw("error writing json", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
			return
		}

		span.AddEvent("response written")
	}
}

func RegisterRoutes(r *router.Mux, exhibitService service.ExhibitService, provisionerService service.ApplicationProvisionerService, log *zap.SugaredLogger, provider trace.TracerProvider) {
	r.AddRoute(router.Get("/api/exhibits", getExhibits(exhibitService, log, provider)))
	r.AddRoute(router.Get("/api/exhibits/{id}", getExhibitById(exhibitService, log, provider)))
	r.AddRoute(router.Post("/api/exhibits", createExhibit(exhibitService, log, provider)))
	r.AddRoute(router.Post("/api/events", handleEvents(exhibitService, provisionerService, log, provider)))
}
