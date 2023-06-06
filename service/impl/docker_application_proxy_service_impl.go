package impl

import (
	"errors"
	"go.uber.org/zap"
	"io"
	"museum/http/router"
	service "museum/service/interface"
	"net/http"
	"time"
)

type DockerApplicationProxyService struct {
	Resolver service.ApplicationResolverService
	Log      *zap.SugaredLogger
}

func (d *DockerApplicationProxyService) ForwardRequest(exhibitId string, path string, res *router.Response, req *router.Request) error {
	// forward to exhibit
	ip, err := d.Resolver.ResolveApplication(req.Context(), exhibitId)
	if err != nil {
		d.Log.Warnw("error resolving application", "error", err, "requestId", req.RequestID, "exhibitId", exhibitId)
		res.WriteHeader(http.StatusInternalServerError)
		return err
	}

	// proxy the request
	proxyReq, err := http.NewRequest(req.Method, "http://"+ip+"/"+path, req.Body)
	if err != nil {
		d.Log.Warnw("error creating proxy request", "error", err, "requestId", req.RequestID, "exhibitId", exhibitId)
		res.WriteHeader(http.StatusInternalServerError)
		return err
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
			d.Log.Warnw("error doing proxy request", "error", err, "requestId", req.RequestID, "exhibitId", exhibitId, "httpStatus", proxyRes.StatusCode)
			res.WriteHeader(http.StatusInternalServerError)
			return err
		}
	case <-time.After(5 * time.Second):
		d.Log.Warnw("timeout doing proxy request", "requestId", req.RequestID, "exhibitId", exhibitId)
		res.WriteHeader(http.StatusInternalServerError)
		return errors.New("timeout doing proxy request")
	}

	for k, v := range proxyRes.Header {
		res.Header().Set(k, v[0])
	}

	res.WriteHeader(proxyRes.StatusCode)

	// read entire body
	body, err := io.ReadAll(proxyRes.Body)
	if err != nil {
		d.Log.Warnw("error reading body", "error", err, "requestId", req.RequestID, "exhibitId", exhibitId)
		res.WriteHeader(http.StatusInternalServerError)
		return err
	}

	_, err = res.Write(body)
	if err != nil {
		d.Log.Warnw("error writing body", "error", err, "requestId", req.RequestID, "exhibitId", exhibitId)
		res.WriteHeader(http.StatusInternalServerError)
		return err
	}

	return nil
}
