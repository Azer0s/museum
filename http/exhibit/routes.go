package exhibit

import (
	_ "embed"
	"museum/http/router"
	"museum/http/router/path"
	"net/http"
	"text/template"
)

//go:embed loading.html
var loadingPage []byte

type LoadingPageTemplate struct {
	Exhibit string
}

func RegisterRoutes(router *router.Mux) {
	tmpl, _ := template.New("loading").Parse(string(loadingPage))

	router.AddRoute(path.Get("/exhibit/{id}/>>", func(w http.ResponseWriter, r *http.Request, pathParameters map[string]string) {
		idMap := map[string]string{
			"foo": "Foo App",
			"bar": "Bar App",
		}

		appName, ok := idMap[pathParameters["id"]]
		if !ok {
			appName = "Museum"
		}

		err := tmpl.Execute(w, LoadingPageTemplate{Exhibit: appName})
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}))
}
