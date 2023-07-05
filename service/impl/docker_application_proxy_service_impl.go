package impl

import (
	"bytes"
	"errors"
	"fmt"
	"go.uber.org/zap"
	"io"
	"museum/config"
	"museum/domain"
	"museum/http"
	service "museum/service/interface"
	gohttp "net/http"
	"strconv"
	"time"
)

type DockerApplicationProxyService struct {
	Resolver       service.ApplicationResolverService
	RewriteService service.RewriteService
	Log            *zap.SugaredLogger
	Config         config.Config
}

func (d *DockerApplicationProxyService) ForwardRequest(exhibit domain.Exhibit, path string, res *http.Response, req *http.Request) error {
	// forward to exhibit
	ip, err := d.Resolver.ResolveApplication(req.Context(), exhibit.Id)
	if err != nil {
		d.Log.Warnw("error resolving application", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
		res.WriteHeader(gohttp.StatusInternalServerError)
		return err
	}

	//TODO: handle websocket
	//TODO: handle SSE

	reqUrl := req.URL.String()

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		d.Log.Warnw("error reading request body", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
		res.WriteHeader(gohttp.StatusInternalServerError)
		return err
	}

	// rewrite request
	if exhibit.Rewrite != nil && *exhibit.Rewrite {
		err = d.RewriteService.RewriteClientRequest(exhibit, ip, req, &reqBody)
		if err != nil {
			d.Log.Warnw("error rewriting request", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
			res.WriteHeader(gohttp.StatusInternalServerError)
			return err
		}
	}

	queryParams := ""
	if req.RawQueryParams != "" {
		queryParams = "?" + req.RawQueryParams
	}

	// create http client
	client := gohttp.Client{
		CheckRedirect: func(req *gohttp.Request, via []*gohttp.Request) error {
			return gohttp.ErrUseLastResponse
		},
	}

	// proxy the request
	proxyReq, err := gohttp.NewRequest(req.Method, "http://"+ip+"/"+path+queryParams, bytes.NewReader(reqBody))
	if err != nil {
		d.Log.Warnw("error creating proxy request", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
		res.WriteHeader(gohttp.StatusInternalServerError)
		return err
	}

	proxyReq.Header = req.Header
	proxyReq.Host = req.Host

	//do request with timeout
	var proxyRes *gohttp.Response
	resultChan := make(chan error)
	go func() {
		var err error
		proxyRes, err = client.Do(proxyReq)
		resultChan <- err
	}()

	select {
	case err := <-resultChan:
		if err != nil {
			d.Log.Warnw("error doing proxy request", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id, "httpStatus", proxyRes.StatusCode)
			res.WriteHeader(gohttp.StatusInternalServerError)
			return err
		}
	case <-time.After(5 * time.Second):
		d.Log.Warnw("timeout doing proxy request", "requestId", req.RequestID, "exhibitId", exhibit.Id)
		res.WriteHeader(gohttp.StatusInternalServerError)
		return errors.New("timeout doing proxy request")
	}

	// TODO: rewrite redirect

	if proxyRes.Request.URL.Path != "/"+path && proxyReq.Method == "GET" {
		// the application redirected us to a different path
		// we need to redirect the user to the new path
		res.Header().Set("Location", "/exhibit/"+exhibit.Id+proxyRes.Request.URL.Path)
		res.WriteHeader(gohttp.StatusTemporaryRedirect)
		return nil
	}

	// read entire body
	resBody, err := io.ReadAll(proxyRes.Body)
	if err != nil {
		d.Log.Warnw("error reading body", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
		res.WriteHeader(gohttp.StatusInternalServerError)
		return err
	}

	// rewrite response
	if exhibit.Rewrite != nil && *exhibit.Rewrite {
		//TODO: rewrite IP addresses in response
		err = d.RewriteService.RewriteServerResponse(exhibit, ip, proxyRes, &resBody)
		if err != nil {
			d.Log.Warnw("error rewriting host", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
			res.WriteHeader(gohttp.StatusInternalServerError)
			return err
		}
		proxyRes.Header.Set("Content-Length", strconv.Itoa(len(resBody)))
	}

	for k, v := range proxyRes.Header {
		res.Header().Set(k, v[0])
	}

	res.WriteHeader(proxyRes.StatusCode)

	_, err = res.Write(resBody)
	if err != nil {
		d.Log.Warnw("error writing body", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
		res.WriteHeader(gohttp.StatusInternalServerError)
		return err
	}

	fmt.Println("Rewriting " + reqUrl + " to " + proxyRes.Request.URL.String())

	return nil
}
