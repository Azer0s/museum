package api

import (
	"encoding/json"
	"io"
	"museum/domain"
	"museum/http/router"
	"museum/service"
	"net/http"
)

func getExhibits(exhibitService service.ExhibitService) router.MuxHandlerFunc {
	return func(res *router.Response, req *http.Request, params map[string]string) {
		exhibits := exhibitService.GetExhibits()
		err := res.WriteJson(exhibits)
		if err != nil {
			//TODO: log error
		}
	}
}

func getExhibitById(exhibitService service.ExhibitService) router.MuxHandlerFunc {
	return func(res *router.Response, req *http.Request, params map[string]string) {
		exhibitId := params["id"]
		exhibit, err := exhibitService.GetExhibitById(exhibitId)
		if err != nil {
			res.WriteHeader(http.StatusNotFound)
			err := res.WriteJson(map[string]string{"status": "Not Found"})
			if err != nil {
				// TODO: log error
			}
		}

		err = res.WriteJson(exhibit)
		if err != nil {
			// TODO: log error
		}
	}
}

func createExhibit(exhibitService service.ExhibitService) router.MuxHandlerFunc {
	return func(res *router.Response, req *http.Request, params map[string]string) {
		body, err := io.ReadAll(req.Body)
		if err != nil {
			// TODO: log error
			return
		}

		exhibit := &domain.Exhibit{}
		err = json.Unmarshal(body, exhibit)
		if err != nil {
			// TODO: log error
			return
		}

		err = exhibitService.CreateExhibit(*exhibit)
		if err != nil {
			res.WriteHeader(http.StatusInternalServerError)
		}

		res.WriteHeader(http.StatusCreated)
		err = res.WriteJson(map[string]string{"status": "Created"})
		if err != nil {
			// TODO: log error
		}
	}
}

func RegisterRoutes(r *router.Mux, exhibitService service.ExhibitService) {
	r.AddRoute(router.Get("/api/exhibits", getExhibits(exhibitService)))
	r.AddRoute(router.Get("/api/exhibits/{id}", getExhibitById(exhibitService)))
	r.AddRoute(router.Post("/api/exhibits", createExhibit(exhibitService)))
}
