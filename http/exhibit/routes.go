package exhibit

import (
	_ "embed"
	"io"
	"museum/config"
	"museum/domain"
	"museum/http/router"
	"museum/persistence"
	service "museum/service/interface"
	"net/http"
	"text/template"
)

//go:embed loading.html
var loadingPage []byte

type LoadingPageTemplate struct {
	Exhibit   string
	Host      string
	ExhibitId string
}

func proxyHandler(state persistence.SharedPersistentEmittedState, resolver service.ApplicationResolverService, c config.Config) router.MuxHandlerFunc {
	tmpl, _ := template.New("loading").Parse(string(loadingPage))
	//TODO: log everything

	return func(res *router.Response, req *router.Request) {
		id, ok := req.Params["id"]
		if !ok {
			res.WriteHeader(http.StatusBadRequest)
			return
		}

		app, err := state.GetExhibitById(id)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		if app.RuntimeInfo.Status != domain.Running {
			err := tmpl.Execute(res, LoadingPageTemplate{
				Exhibit:   app.Name,
				Host:      c.GetHostname() + ":" + c.GetPort(),
				ExhibitId: app.Id,
			})
			//TODO: start exhibit
			if err != nil {
				res.WriteHeader(http.StatusInternalServerError)
				return
			}
			return
		}

		// forward to exhibit
		ip, err := resolver.ResolveApplication(id)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		// proxy the request
		proxyReq, err := http.NewRequest(req.Method, "http://"+ip, req.Body)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}

		proxyReq.Header = req.Header
		proxyReq.Host = req.Host

		proxyRes, err := http.DefaultClient.Do(proxyReq)
		if err != nil {
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
func RegisterRoutes(r *router.Mux, state persistence.SharedPersistentEmittedState, resolver service.ApplicationResolverService, config config.Config) {
	r.AddRoute(router.Get("/exhibit/{id}/>>", proxyHandler(state, resolver, config)))
}
