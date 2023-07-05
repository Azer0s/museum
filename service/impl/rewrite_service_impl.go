package impl

import (
	"github.com/google/uuid"
	"go.uber.org/zap"
	"museum/config"
	"museum/domain"
	"museum/http"
	"museum/util"
	gohttp "net/http"
	"net/url"
	"strings"
)

type RewriteServiceImpl struct {
	Config config.Config
	Log    *zap.SugaredLogger
}

var placeHolderHost = strings.ReplaceAll(uuid.New().String(), "-", "")

func (r *RewriteServiceImpl) RewriteServerResponse(exhibit domain.Exhibit, ip string, res *gohttp.Response, body *[]byte) error {
	// alright, so we have to rewrite the response
	// 1: "http://172.168.0.3:9090/foo/bar" changes to "http://localhost:8080/exhibit/123/foo/bar"
	// 2: "http://localhost:8080/foo/bar" changes to "http://localhost:8080/exhibit/123/foo/bar"
	// 3: "http://localhost:8080/exhibit/123/foo/bar" changes to "http://localhost:8080/exhibit/123/foo/bar" (not "http://localhost:8080/exhibit/123/exhibit/123/foo/bar")

	// get encoding from header
	encoding := res.Header.Get("Content-Encoding")
	bodyDecoded, err := util.DecodeBody(*body, encoding)
	if err != nil {
		r.Log.Warnw("error decoding body", "error", err, "requestId", exhibit.Id)
		return err
	}

	bodyStr := string(bodyDecoded)

	// let's rewrite the IP case in the res headers
	for k := range res.Header {
		h := strings.ReplaceAll(res.Header.Get(k), ip, r.Config.GetHostname()+"/exhibit/"+exhibit.Id)
		h = strings.ReplaceAll(h, "//", "/")
		res.Header.Set(k, h)
	}
	bodyStr = strings.ReplaceAll(bodyStr, ip, r.Config.GetHostname()+"/exhibit/"+exhibit.Id)

	// now let's rewrite every case 3 to an uuid,
	// so we don't accidentally rewrite during case 2
	for k := range res.Header {
		h := strings.ReplaceAll(res.Header.Get(k), r.Config.GetHostname()+"/exhibit/"+exhibit.Id, placeHolderHost)
		h = strings.ReplaceAll(h, "//", "/")
		res.Header.Set(k, h)
	}
	bodyStr = strings.ReplaceAll(bodyStr, r.Config.GetHostname()+"/exhibit/"+exhibit.Id, placeHolderHost)

	// now let's rewrite every case 2
	for k := range res.Header {
		h := strings.ReplaceAll(res.Header.Get(k), r.Config.GetHostname(), r.Config.GetHostname()+"/exhibit/"+exhibit.Id)
		h = strings.ReplaceAll(h, "//", "/")
		res.Header.Set(k, h)
	}
	bodyStr = strings.ReplaceAll(bodyStr, r.Config.GetHostname(), r.Config.GetHostname()+"/exhibit/"+exhibit.Id)

	// now let's rewrite every uuid to the original path
	for k := range res.Header {
		h := strings.ReplaceAll(res.Header.Get(k), placeHolderHost, r.Config.GetHostname()+"/exhibit/"+exhibit.Id)
		h = strings.ReplaceAll(h, "//", "/")
		res.Header.Set(k, h)
	}
	bodyStr = strings.ReplaceAll(bodyStr, placeHolderHost, r.Config.GetHostname()+"/exhibit/"+exhibit.Id)

	copy(*body, bodyStr)

	return nil
}

func (r *RewriteServiceImpl) RewriteClientRequest(exhibit domain.Exhibit, ip string, req *http.Request, body *[]byte) error {
	// alright, so we have to rewrite the request
	// 1: "http://localhost:8080/exhibit/123/foo/bar" changes to "http://ip:port/foo/bar"
	// 2: "http://localhost:8080/foo/bar" changes to "http://ip:port/foo/bar"

	// get encoding from header
	encoding := req.Header.Get("Content-Encoding")
	bodyDecoded, err := util.DecodeBody(*body, encoding)
	if err != nil {
		r.Log.Warnw("error decoding body", "error", err, "requestId", exhibit.Id)
		return err
	}

	bodyStr, err := url.QueryUnescape(string(bodyDecoded))
	if err != nil {
		r.Log.Warnw("error unescaping body", "error", err, "requestId", exhibit.Id)
		return err
	}

	// let's rewrite case 1
	for k := range req.Header {
		h := strings.ReplaceAll(req.Header.Get(k), r.Config.GetHostname()+"/exhibit/"+exhibit.Id, ip)
		h = strings.ReplaceAll(h, "//", "/")
		req.Header.Set(k, h)
	}

	*body, err = util.EncodeBody([]byte(bodyStr), encoding)
	if err != nil {
		r.Log.Warnw("error encoding body", "error", err, "requestId", exhibit.Id)
		return err
	}
	bodyStr = strings.ReplaceAll(bodyStr, r.Config.GetHostname(), ip)

	copy(*body, bodyStr)

	return nil
}
