package exhibit

import (
	_ "embed"
	"museum/http/router"
	"net/http"
	"text/template"
)

//go:embed loading.html
var loadingPage []byte

type LoadingPageTemplate struct {
	Exhibit string
}

func loadingPageHandler() router.MuxHandlerFunc {
	tmpl, _ := template.New("loading").Parse(string(loadingPage))

	return func(res *router.Response, _ *http.Request, parameters map[string]string) {
		idMap := map[string]string{
			"foo": "Foo App",
			"bar": "Bar App",
		}

		appName, ok := idMap[parameters["id"]]
		if !ok {
			appName = "Museum"
		}

		err := tmpl.Execute(res, LoadingPageTemplate{Exhibit: appName})
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
			return
		}
	}
}
func RegisterRoutes(r *router.Mux) {
	r.AddRoute(router.Get("/exhibit/{id}/>>", loadingPageHandler()))
}
