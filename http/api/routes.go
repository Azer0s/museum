package api

import (
	"encoding/json"
	"museum/http/router"
	"museum/http/router/path"
	"museum/service"
	"net/http"
)

func RegisterRoutes(router *router.Mux, exhibitService service.ExhibitService) {
	router.AddRoute(path.Get("/api/exhibits", func(writer http.ResponseWriter, request *http.Request, m map[string]string) {
		exhibits := exhibitService.GetExhibits()

		b, err := json.Marshal(exhibits)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}

		_, err = writer.Write(b)
		if err != nil {
			writer.WriteHeader(http.StatusInternalServerError)
			return
		}
	}))
}
