package http

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
)

type Response struct {
	http.ResponseWriter
}

func (r *Response) SendMessage(event string, data map[string]string) error {
	flusher, ok := r.ResponseWriter.(http.Flusher)
	if !ok {
		return errors.New("streaming unsupported")
	}

	_, err := fmt.Fprintf(r.ResponseWriter, "event: %s\n", event)
	if err != nil {
		return err
	}

	dataStr, err := json.Marshal(data)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(r.ResponseWriter, "data: %s\n\n", dataStr)
	if err != nil {
		return err
	}

	flusher.Flush()

	return nil
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

func (r *Response) SetupSSE() error {
	_, ok := r.ResponseWriter.(http.Flusher)
	if !ok {
		return errors.New("streaming unsupported")
	}

	// initialize SSE
	r.Header().Set("Content-Type", "text/event-stream")
	r.Header().Set("Cache-Control", "no-cache")
	r.Header().Set("Connection", "keep-alive")
	r.Header().Set("Access-Control-Allow-Origin", "*")

	return nil
}

func (r *Response) CloseSSE() error {
	flusher, ok := r.ResponseWriter.(http.Flusher)
	if !ok {
		return errors.New("streaming unsupported")
	}

	_, err := fmt.Fprintf(r.ResponseWriter, "event: close\n")
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(r.ResponseWriter, "data: {}\n\n")
	if err != nil {
		return err
	}

	flusher.Flush()

	// get underlying TCP connection
	hijacker, ok := r.ResponseWriter.(http.Hijacker)
	if !ok {
		return errors.New("hijacking unsupported")
	}

	hijack, _, err := hijacker.Hijack()
	if err != nil {
		return err
	}

	err = hijack.Close()
	if err != nil {
		return err
	}

	return nil
}
