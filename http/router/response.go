package router

import (
	"encoding/json"
	"net/http"
)

type Response struct {
	http.ResponseWriter
}

func (r *Response) WriteJson(data any) error {
	b, err := json.Marshal(data)
	if err != nil {
		r.WriteHeader(http.StatusInternalServerError)
		return err
	}

	_, err = r.Write(b)
	if err != nil {
		r.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return nil
}
