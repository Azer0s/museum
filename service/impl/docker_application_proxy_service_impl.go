package impl

import (
	"errors"
	"github.com/google/uuid"
	"go.uber.org/zap"
	"io"
	"museum/config"
	"museum/domain"
	"museum/http"
	service "museum/service/interface"
	"museum/util"
	gohttp "net/http"
	"strconv"
	"strings"
	"time"
)

type DockerApplicationProxyService struct {
	Resolver service.ApplicationResolverService
	Log      *zap.SugaredLogger
	Config   config.Config
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
	queryParams := ""
	if req.RawQueryParams != nil {
		queryParams = "?" + *req.RawQueryParams
	}

	// proxy the request
	proxyReq, err := gohttp.NewRequest(req.Method, "http://"+ip+"/"+path+queryParams, req.Body)
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
		proxyRes, err = gohttp.DefaultClient.Do(proxyReq)
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

	if proxyRes.Request.URL.Path != "/"+path {
		// the application redirected us to a different path
		// we need to redirect the user to the new path
		res.Header().Set("Location", "/exhibit/"+exhibit.Id+proxyRes.Request.URL.Path)
		res.WriteHeader(gohttp.StatusTemporaryRedirect)
		return nil
	}

	// read entire body
	body, err := io.ReadAll(proxyRes.Body)
	if err != nil {
		d.Log.Warnw("error reading body", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
		res.WriteHeader(gohttp.StatusInternalServerError)
		return err
	}

	if exhibit.Rewrite != nil && *exhibit.Rewrite {
		err := d.rewriteHost(exhibit, proxyRes, &body)
		if err != nil {
			d.Log.Warnw("error rewriting host", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
			res.WriteHeader(gohttp.StatusInternalServerError)
			return err
		}
		proxyRes.Header.Set("Content-Length", strconv.Itoa(len(body)))
	}

	for k, v := range proxyRes.Header {
		res.Header().Set(k, v[0])
	}

	res.WriteHeader(proxyRes.StatusCode)

	_, err = res.Write(body)
	if err != nil {
		d.Log.Warnw("error writing body", "error", err, "requestId", req.RequestID, "exhibitId", exhibit.Id)
		res.WriteHeader(gohttp.StatusInternalServerError)
		return err
	}

	return nil
}

func (d *DockerApplicationProxyService) rewriteHost(exhibit domain.Exhibit, proxyRes *gohttp.Response, body *[]byte) error {
	// get encoding from header
	encoding := proxyRes.Header.Get("Content-Encoding")
	bodyDecoded, err := util.DecodeBody(*body, encoding)
	if err != nil {
		d.Log.Warnw("error decoding body", "error", err, "requestId", exhibit.Id)
		return err
	}

	bodyStr := string(bodyDecoded)

	// let's rewrite some paths in the body
	searchHost := d.Config.GetHostname() + ":" + d.Config.GetPort()
	replaceHost := d.Config.GetHostname() + ":" + d.Config.GetPort() + "/exhibit/" + exhibit.Id

	// replace all occurrences of the searchHost with the replaceHost
	// don't replace the replaceHost with the replaceHost
	placeholderHost := uuid.New().String()
	bodyStr = strings.ReplaceAll(bodyStr, replaceHost, placeholderHost)
	bodyStr = strings.ReplaceAll(bodyStr, searchHost, replaceHost)
	bodyStr = strings.ReplaceAll(bodyStr, placeholderHost, replaceHost)

	*body, err = util.EncodeBody([]byte(bodyStr), encoding)
	if err != nil {
		d.Log.Warnw("error encoding body", "error", err, "requestId", exhibit.Id)
		return err
	}

	return nil
}
