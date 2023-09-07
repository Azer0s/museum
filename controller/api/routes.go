package api

import (
	"encoding/json"
	cloudevents "github.com/cloudevents/sdk-go/v2/event"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"io"
	"museum/domain"
	"museum/http"
	"museum/persistence"
	"museum/service"
	gohttp "net/http"
	"time"
)

func getExhibits(exhibitService service.ExhibitService, log *zap.SugaredLogger, provider trace.TracerProvider) http.MuxHandlerFunc {
	return func(res *http.Response, req *http.Request) {
		subCtx, span := provider.
			Tracer("API request").
			Start(req.Context(), "HTTP GET /api/exhibits/", trace.WithAttributes(attribute.String("requestId", req.RequestID)))
		defer span.End()

		exhibits := exhibitService.GetAllExhibits(subCtx)
		dtos := make([]domain.ExhibitDto, len(exhibits))

		for i, exhibit := range exhibits {
			dtos[i] = exhibit.ToDto()
		}

		err := res.WriteJson(dtos)
		if err != nil {
			span.RecordError(err)
			log.Warnw("error writing json", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
		}
	}
}

func getExhibitById(exhibitService service.ExhibitService, log *zap.SugaredLogger, provider trace.TracerProvider) http.MuxHandlerFunc {
	return func(res *http.Response, req *http.Request) {
		subCtx, span := provider.
			Tracer("API request").
			Start(req.Context(), "HTTP GET /api/exhibits/"+req.Params["id"], trace.WithAttributes(attribute.String("requestId", req.RequestID)))
		defer span.End()

		exhibitId := req.Params["id"]
		exhibit, err := exhibitService.GetExhibitById(subCtx, exhibitId)
		if err != nil {
			span.RecordError(err)
			res.WriteErr(err)
			res.WriteHeader(gohttp.StatusNotFound)
			return
		}

		err = res.WriteJson(exhibit.ToDto())
		if err != nil {
			span.RecordError(err)
			log.Warnw("error writing json", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
		}
	}
}

func createExhibit(exhibitService service.ExhibitService, log *zap.SugaredLogger, provider trace.TracerProvider) http.MuxHandlerFunc {
	return func(res *http.Response, req *http.Request) {
		ctx, span := provider.
			Tracer("API request").
			Start(req.Context(), "HTTP POST /api/exhibits", trace.WithAttributes(attribute.String("requestId", req.RequestID)))
		defer span.End()

		body, err := io.ReadAll(req.Body)
		if err != nil {
			span.RecordError(err)
			log.Warnw("error reading request body", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
			return
		}

		exhibit := &domain.Exhibit{}
		err = json.Unmarshal(body, exhibit)
		if err != nil {
			span.RecordError(err)
			log.Warnw("error unmarshalling json", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
			return
		}

		span.AddEvent("request read")

		id, err := exhibitService.CreateExhibit(ctx, domain.CreateExhibit{
			Exhibit:   *exhibit,
			RequestID: req.RequestID,
		})
		if err != nil {
			span.RecordError(err)
			log.Warnw("error creating exhibit", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
			return
		}

		res.WriteHeader(gohttp.StatusCreated)
		err = res.WriteJson(map[string]string{"status": "Created", "id": id})
		if err != nil {
			span.RecordError(err)
			log.Warnw("error writing json", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
			return
		}

		span.AddEvent("response written")
	}
}

func handleEvents(handlerService service.ApplicationProvisionerHandlerService, log *zap.SugaredLogger, provider trace.TracerProvider) http.MuxHandlerFunc {
	return func(res *http.Response, req *http.Request) {
		ctx, span := provider.
			Tracer("API request").
			Start(req.Context(), "HTTP POST /api/events", trace.WithAttributes(attribute.String("requestId", req.RequestID)))
		defer span.End()

		event := &cloudevents.Event{}
		err := cloudevents.ReadJson(event, req.Body)

		if err != nil {
			span.RecordError(err)
			log.Warnw("error reading or parsing request body", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
			return
		}

		span.AddEvent("request read")

		exhibit := &domain.Exhibit{}
		err = json.Unmarshal(event.Data(), exhibit)
		if err != nil {
			span.RecordError(err)
			log.Warnw("error unmarshalling cloudevent data", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
			return
		}

		err = handlerService.HandleEvent(ctx, event, exhibit.Id)

		span.AddEvent("application started")

		res.WriteHeader(gohttp.StatusCreated)
		err = res.WriteJson(map[string]string{"status": "Started"})
		if err != nil {
			span.RecordError(err)
			log.Warnw("error writing json", "error", err, "requestId", req.RequestID)
			res.WriteErr(err)
			return
		}

		span.AddEvent("response written")
	}
}

func handleExhibitStatus(exhibitService service.ExhibitService, eventing persistence.Eventing, log *zap.SugaredLogger, provider trace.TracerProvider) http.MuxHandlerFunc {
	return func(res *http.Response, req *http.Request) {
		ctx, span := provider.
			Tracer("API request").
			Start(req.Context(), "HTTP GET /api/exhibits/"+req.Params["id"]+"/status", trace.WithAttributes(attribute.String("requestId", req.RequestID)))
		defer span.End()

		exhibitId := req.Params["id"]

		// get the exhibit
		_, err := exhibitService.GetExhibitById(ctx, exhibitId)
		if err != nil {
			span.RecordError(err)
			log.Warnw("error getting exhibit", "error", err, "requestId", req.RequestID, "exhibitId", exhibitId)
			res.WriteHeader(gohttp.StatusInternalServerError)
			return
		}

		span.AddEvent("setting up SSE")

		err = res.SetupSSE()
		if err != nil {
			span.RecordError(err)
			log.Warnw("error setting up SSE", "error", err, "requestId", req.RequestID)
			gohttp.Error(res, "Streaming unsupported!", gohttp.StatusInternalServerError)
			return
		}

		span.AddEvent("sending initial message")

		// send initial SSE message
		err = res.SendMessage("status.subscribed", map[string]string{"exhibitId": exhibitId})
		if err != nil {
			span.RecordError(err)
			log.Warnw("error sending message", "error", err, "requestId", req.RequestID)
			return
		}

		timeOut := time.After(5 * time.Second)
		events, cancel, err := eventing.GetExhibitStartingChannel(exhibitId, ctx)
		if err != nil {
			span.RecordError(err)
			log.Warnw("error getting exhibit starting channel", "error", err, "requestId", req.RequestID)
			return
		}

		defer func() {
			cancel()
			err = res.CloseSSE()
			if err != nil {
				log.Warnw("error closing SSE", "error", err, "requestId", req.RequestID)
			}
		}()

		//TODO: test this

		for {
			select {
			case <-timeOut:
				log.Warnw("timeout reached, stopping SSE", "requestId", req.RequestID)
				return
			case <-ctx.Done():
				log.Debugw("context cancelled, stopping SSE", "requestId", req.RequestID)
				return
			case event := <-events:
				err := res.SendMessage("status.update", event.ToMap())
				if err != nil {
					log.Warnw("error sending message", "error", err, "requestId", req.RequestID)
					return
				}
			}
		}
	}
}

func RegisterRoutes(r *http.Mux, exhibitService service.ExhibitService, eventing persistence.Eventing, provisionerHandlerService service.ApplicationProvisionerHandlerService, log *zap.SugaredLogger, provider trace.TracerProvider) {
	r.AddRoute(http.Get("/api/exhibits", getExhibits(exhibitService, log, provider)))
	r.AddRoute(http.Get("/api/exhibits/{id}", getExhibitById(exhibitService, log, provider)))
	r.AddRoute(http.Get("/api/exhibits/{id}/status", handleExhibitStatus(exhibitService, eventing, log, provider)))
	r.AddRoute(http.Post("/api/exhibits", createExhibit(exhibitService, log, provider)))
	r.AddRoute(http.Post("/api/events", handleEvents(provisionerHandlerService, log, provider)))
}
