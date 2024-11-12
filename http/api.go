package http

import (
	"net/http"
	"os"
)

var certFile, keyFile string

func ListenAndServe(addr string, handler http.Handler) error {
	if certFile != "" && keyFile != "" {
		return http.ListenAndServeTLS(addr, certFile, keyFile, handler)
	}

	return http.ListenAndServe(addr, handler)
}

func ConfigureTls(cf, kf string) error {
	if _, err := os.Stat(cf); os.IsNotExist(err) {
		return err
	}

	if _, err := os.Stat(kf); os.IsNotExist(err) {
		return err
	}

	certFile, keyFile = cf, kf
	return nil
}
