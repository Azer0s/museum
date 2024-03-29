package impl

import (
	"bytes"
	"errors"
	"go.uber.org/zap"
	"io"
	"museum/config"
	"museum/domain"
	"museum/http"
	service "museum/service/interface"
	gohttp "net/http"
	"net/url"
	"strconv"
	"strings"
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

	port := ""
	for _, o := range exhibit.Objects {
		if exhibit.Expose == o.Name {
			if o.Port != nil {
				port = *o.Port
				break
			}

			port = "80"
			break
		}
	}

	//TODO: handle websocket
	//TODO: handle SSE

	reqBody, err := io.ReadAll(req.Body)
	if err != nil {
		d.Log.Warnw("error reading request body", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
		res.WriteHeader(gohttp.StatusInternalServerError)
		return err
	}

	// rewrite request
	if exhibit.Rewrite != nil && *exhibit.Rewrite {
		err = d.RewriteService.RewriteClientRequest(exhibit, ip+":"+port, req, &reqBody)
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
	proxyReq, err := gohttp.NewRequest(req.Method, "http://"+ip+":"+port+"/"+path+queryParams, bytes.NewReader(reqBody))
	if err != nil {
		d.Log.Warnw("error creating proxy request", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
		res.WriteHeader(gohttp.StatusInternalServerError)
		return err
	}

	proxyReq.Header = req.Header
	proxyReq.Host = req.Host

	//do request with timeout
	proxyRes := new(gohttp.Response)
	resultChan := make(chan error)
	go func() {
		var err error
		proxyRes, err = client.Do(proxyReq)
		resultChan <- err
	}()

	select {
	case err := <-resultChan:
		if err != nil {
			if proxyRes == nil {
				proxyRes = &gohttp.Response{
					StatusCode: gohttp.StatusInternalServerError,
				}
			}

			d.Log.Warnw("error doing proxy request", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id, "httpStatus", proxyRes.StatusCode)
			res.WriteHeader(gohttp.StatusInternalServerError)
			return err
		}
	case <-time.After(5 * time.Second):
		d.Log.Warnw("timeout doing proxy request", "requestId", req.RequestID, "exhibitId", exhibit.Id)
		res.WriteHeader(gohttp.StatusInternalServerError)
		return errors.New("timeout doing proxy request")
	}

	if proxyRes.Request.URL.Path != "/"+path && proxyReq.Method == "GET" {
		// the application redirected us to a different path
		// we need to redirect the user to the new path
		res.Header().Set("Location", "/exhibit/"+exhibit.Id+proxyRes.Request.URL.Path)
		res.WriteHeader(gohttp.StatusTemporaryRedirect)
		return nil
	}

	// read entire body
	resBody, err := func() (*[]byte, error) {
		// this might look a bit hacky, but it's actually
		// a good way to shadow the resBody with a different type
		resBody, err := io.ReadAll(proxyRes.Body)
		if err != nil {
			return nil, err
		}
		return &resBody, nil
	}()

	if err != nil {
		d.Log.Warnw("error reading body", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
		res.WriteHeader(gohttp.StatusInternalServerError)
		return err
	}

	// rewrite response
	if exhibit.Rewrite != nil && *exhibit.Rewrite {
		resBody, err = d.RewriteService.RewriteServerResponse(exhibit, ip+":"+port, proxyRes, resBody)
		if err != nil {
			d.Log.Warnw("error rewriting host", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
			res.WriteHeader(gohttp.StatusInternalServerError)
			return err
		}
		proxyRes.Header.Set("Content-Length", strconv.Itoa(len(*resBody)))
	}

	for k, v := range proxyRes.Header {
		res.Header().Set(k, v[0])
	}

	if proxyRes.StatusCode > 299 && proxyRes.StatusCode < 400 {
		// the application redirected us to a different path
		// we need to redirect the user to the new path
		redirectUrl, err := url.Parse(proxyRes.Header.Get("Location"))
		if err != nil {
			d.Log.Warnw("error parsing redirect url", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
			res.WriteHeader(gohttp.StatusInternalServerError)
			return err
		}

		// restructure from http://localhost:8080/foo/bar
		// to http://localhost:8080/exhibit/123/foo/bar
		// if, somehow, the redirect url already contains the exhibit id, we don't want to add it again

		if strings.Contains(redirectUrl.String(), "/exhibit/"+exhibit.Id) {
			res.WriteHeader(gohttp.StatusTemporaryRedirect)
			return nil
		}

		res.Header().Set("Location", "/exhibit/"+exhibit.Id+redirectUrl.Path)
		res.WriteHeader(gohttp.StatusTemporaryRedirect)
		return nil
	}

	res.WriteHeader(proxyRes.StatusCode)

	_, err = res.Write(*resBody)
	if err != nil {
		d.Log.Warnw("error writing body", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
		res.WriteHeader(gohttp.StatusInternalServerError)
		return err
	}

	return nil
}
