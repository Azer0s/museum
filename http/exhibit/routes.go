package exhibit

import (
	"context"
	_ "embed"
	"go.uber.org/zap"
	"io"
	"museum/config"
	"museum/domain"
	"museum/http/router"
	"museum/persistence"
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

func proxyHandler(state persistence.SharedPersistentEmittedState, resolver service.ApplicationResolverService, provisioner service.ApplicationProvisionerService, log *zap.SugaredLogger, c config.Config) router.MuxHandlerFunc {
	tmpl, _ := template.New("loading").Parse(string(loadingPage))
	//TODO: log everything

	return func(res *router.Response, req *router.Request) {
		id, ok := req.Params["id"]
		if !ok {
			res.WriteHeader(http.StatusBadRequest)
			log.Warn("no id provided", "requestId", req.RequestID)
			return
		}

		app, err := state.GetExhibitById(id)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			log.Warnw("error getting exhibit", "error", err, "requestId", req.RequestID)
			return
		}

		// if the application is stopping, return a 503
		if app.RuntimeInfo.Status == domain.Stopping {
			res.WriteHeader(http.StatusServiceUnavailable)
			return
		}

		// if the application is not running, start it and return the loading page
		// if the state is starting, only return the loading page
		if app.RuntimeInfo.Status != domain.Running {
			err := tmpl.Execute(res, LoadingPageTemplate{
				Exhibit:   app.Name,
				Host:      c.GetHostname() + ":" + c.GetPort(),
				ExhibitId: app.Id,
			})

			if app.RuntimeInfo.Status != domain.Starting {
				go func() {
					err := provisioner.StartApplication(context.Background(), id)
					if err != nil {
						log.Warnw("error starting application", "error", err, "requestId", req.RequestID)
						return
					}
				}()
			}

			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				log.Warnw("error executing template", "error", err, "requestId", req.RequestID)
				return
			}
			return
		}

		// forward to exhibit
		ip, err := resolver.ResolveApplication(id)
		if err != nil {
			log.Debugw("error resolving application", "error", err, "requestId", req.RequestID)
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// proxy the request
		proxyReq, err := http.NewRequest(req.Method, "http://"+ip, req.Body)
		if err != nil {
			log.Debugw("error creating proxy request", "error", err, "requestId", req.RequestID)
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
				log.Debugw("error doing proxy request", "error", err, "requestId", req.RequestID)
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
		case <-time.After(5 * time.Second):
			log.Debugw("timeout doing proxy request", "requestId", req.RequestID)
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
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = res.Write(body)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
func RegisterRoutes(r *router.Mux, state persistence.SharedPersistentEmittedState, resolver service.ApplicationResolverService, provisioner service.ApplicationProvisionerService, log *zap.SugaredLogger, config config.Config) {
	r.AddRoute(router.Get("/exhibit/{id}/>>", proxyHandler(state, resolver, provisioner, log, config)))
}
