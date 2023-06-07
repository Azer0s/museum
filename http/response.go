package http

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

func (r *Response) WriteErr(err error) {
	r.WriteHeader(http.StatusInternalServerError)
	_ = r.WriteJson(map[string]string{"status": "Internal Server Error", "error": err.Error()})
}
