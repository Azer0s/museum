package exhibit

import (
	"context"
	_ "embed"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
	"io"
	"museum/config"
	"museum/domain"
	"museum/http/router"
	service "museum/service/interface"
	"net/http"
	"text/template"
	"time"
)

//go:embed loading.html
var loadingPage []byte

type LoadingPageTemplate struct {
	Exhibit   string
	Host      string
	ExhibitId string
}

func proxyHandler(exhibitService service.ExhibitService, lastAccessedService service.LastAccessedService, resolver service.ApplicationResolverService, provisioner service.ApplicationProvisionerService, log *zap.SugaredLogger, c config.Config, provider trace.TracerProvider) router.MuxHandlerFunc {
	tmpl, _ := template.New("loading").Parse(string(loadingPage))

	return func(res *router.Response, req *router.Request) {
		id, ok := req.Params["id"]
		if !ok {
			log.Warn("no id provided", "requestId", req.RequestID)
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		app, err := exhibitService.GetExhibitById(req.Context(), id)
		if err != nil {
			log.Warnw("error getting exhibit", "error", err, "requestId", req.RequestID, "exhibitId", id)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// if the application is stopping, return a 503
		if app.RuntimeInfo.Status == domain.Stopping {
			log.Warnw("application is stopping, returning 503", "requestId", req.RequestID, "status", app.RuntimeInfo.Status, "exhibitId", app.Id)
			res.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// if the application is not running, start it and return the loading page
		// if the state is "starting", only return the loading page
		if app.RuntimeInfo.Status != domain.Running {
			log.Infow("application is not running, returning loading page", "requestId", req.RequestID, "status", app.RuntimeInfo.Status, "exhibitId", app.Id)

			ctx, span := provider.
				Tracer("API request to non-running application").
				Start(context.Background(), "HTTP "+req.Method+" "+req.URL.Path, trace.WithAttributes(attribute.String("requestId", req.RequestID), attribute.String("exhibitId", app.Id)))
			defer span.End()

			// if the application is not starting, start it
			err := tmpl.Execute(res, LoadingPageTemplate{
				Exhibit:   app.Name,
				Host:      c.GetHostname() + ":" + c.GetPort(),
				ExhibitId: app.Id,
			})
			span.AddEvent("loading page rendered")

			if app.RuntimeInfo.Status != domain.Starting {
				log.Infow("starting application", "requestId", req.RequestID, "exhibitId", app.Id)
				go func() {
					// create a new span for starting the application, the tracer should outlive the request
					subCtx, subSpan := provider.
						Tracer("Starting application"+app.Name+" ("+app.Id+")").
						Start(ctx, "Starting application", trace.WithAttributes(attribute.String("requestId", req.RequestID), attribute.String("exhibitId", app.Id)))
					defer subSpan.End()

					subSpan.AddEvent("starting application")
					err := provisioner.StartApplication(subCtx, id)
					if err != nil {
						log.Warnw("error starting application", "error", err, "requestId", req.RequestID, "exhibitId", app.Id)
						return
					}
					log.Infow("application started", "requestId", req.RequestID, "exhibitId", app.Id)
					subSpan.AddEvent("application started")
				}()
			}

			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				log.Warnw("error executing template", "error", err, "requestId", req.RequestID, "exhibitId", app.Id)
				return
			}

			span.AddEvent("loading page returned")

			return
		}

		// forward to exhibit
		ip, err := resolver.ResolveApplication(req.Context(), id)
		if err != nil {
			log.Warnw("error resolving application", "error", err, "requestId", req.RequestID, "exhibitId", app.Id)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// proxy the request
		proxyReq, err := http.NewRequest(req.Method, "http://"+ip, req.Body)
		if err != nil {
			log.Warnw("error creating proxy request", "error", err, "requestId", req.RequestID, "exhibitId", app.Id)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		proxyReq.Header = req.Header
		proxyReq.Host = req.Host

		//do request with timeout
		var proxyRes *http.Response
		resultChan := make(chan error)
		go func() {
			var err error
			proxyRes, err = http.DefaultClient.Do(proxyReq)
			resultChan <- err
		}()

		select {
		case err := <-resultChan:
			if err != nil {
				log.Warnw("error doing proxy request", "error", err, "requestId", req.RequestID, "exhibitId", app.Id, "httpStatus", proxyRes.StatusCode)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		case <-time.After(5 * time.Second):
			log.Warnw("timeout doing proxy request", "requestId", req.RequestID, "exhibitId", app.Id)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		for k, v := range proxyRes.Header {
			res.Header().Set(k, v[0])
		}

		res.WriteHeader(proxyRes.StatusCode)

		// read entire body
		body, err := io.ReadAll(proxyRes.Body)
		if err != nil {
			log.Warnw("error reading body", "error", err, "requestId", req.RequestID, "exhibitId", app.Id)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = res.Write(body)
		if err != nil {
			log.Warnw("error writing body", "error", err, "requestId", req.RequestID, "exhibitId", app.Id)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		go func() {
			err := lastAccessedService.SetLastAccessed(context.Background(), id, time.Now().Unix())
			if err != nil {
				return
			}
			log.Debugw("finished proxy request", "requestId", req.RequestID, "exhibitId", app.Id)
		}()
	}
}

func RegisterRoutes(r *router.Mux, exhibitService service.ExhibitService, lastAccessedService service.LastAccessedService, resolver service.ApplicationResolverService, provisioner service.ApplicationProvisionerService, log *zap.SugaredLogger, config config.Config, provider trace.TracerProvider) {
	r.AddRoute(router.Get("/exhibit/{id}/>>", proxyHandler(exhibitService, lastAccessedService, resolver, provisioner, log, config, provider)))
}
