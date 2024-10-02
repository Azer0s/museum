package impl

import (
	"github.com/google/uuid"
	"github.com/yosssi/gohtml"
	"go.uber.org/zap"
	"museum/config"
	"museum/domain"
	"museum/http"
	"museum/util"
	gohttp "net/http"
	"net/url"
	"regexp"
	"strings"
)

type RewriteServiceImpl struct {
	Config config.Config
	Log    *zap.SugaredLogger
}

var placeHolderHost = strings.ReplaceAll(uuid.New().String(), "-", "")
var hrefSrcReg = regexp.MustCompile("(href|src|action) *= *([\"'])(\\w+[\\w=?/&.]*)([\"'])")
var hrefSrcBaseReg = regexp.MustCompile("(href|src|action) *= *([\"'])/")

// gets the FQHN (Fully Qualified Host Name) of the museum
func (r *RewriteServiceImpl) getFqhn() string {
	return r.Config.GetHostname() + ":" + r.Config.GetPort()
}

/*
 FIXME: currently http://localhost:8080/exhibit/fd8006cd-39b9-4c8b-816b-50152bbef02b/wp-login.php?redirect_to=http%3A%2F%2Flocalhost%3A8080%2Fwp-admin%2F&reauth=1
 does not work, because the redirect_to parameter is not rewritten because we are not idiots that rewrite query params
 but apparently WP devs are idiots that use query params for redirects :/
*/

func (r *RewriteServiceImpl) RewriteServerResponse(exhibit domain.Exhibit, hostname string, res *gohttp.Response, body *[]byte) (*[]byte, error) {
	// alright, so we have to rewrite the response
	// 1: "http://172.168.0.3:9090/foo/bar" changes to "http://localhost:8080/exhibit/123/foo/bar"
	// 2: "http://localhost:8080/foo/bar" changes to "http://localhost:8080/exhibit/123/foo/bar"
	// 3: "http://localhost:8080/exhibit/123/foo/bar" changes to "http://localhost:8080/exhibit/123/foo/bar" (not "http://localhost:8080/exhibit/123/exhibit/123/foo/bar")
	// 4: change hrefs from "/foo/bar" to "/exhibit/123/foo/bar"
	// 5: change srcs from "/foo/bar" to "/exhibit/123/foo/bar"

	util.Nop(hostname)

	// check if res is a redirect
	if res.StatusCode >= 300 && res.StatusCode < 400 {
		// get the redirect url
		redirectUrl, err := res.Location()
		if err != nil {
			return nil, err
		}

		// rewrite the redirect url
		redirectUrlStr := redirectUrl.String()
		redirectUrlStr = strings.ReplaceAll(redirectUrlStr, r.getFqhn(), r.getFqhn()+"/exhibit/"+exhibit.Id)

		res.Header.Set("Location", redirectUrlStr)
	}

	// check content type
	if !strings.Contains(res.Header.Get("Content-Type"), "text/html") {
		return body, nil
	}

	// get encoding from header
	encoding := res.Header.Get("Content-Encoding")
	bodyDecoded, err := util.DecodeBody(*body, encoding)
	if err != nil {
		r.Log.Warnw("error decoding body", "error", err, "requestId", exhibit.Id)
		return nil, err
	}

	bodyStr := string(bodyDecoded)
	bodyStr = gohtml.Format(bodyStr)

	bodyStr = hrefSrcBaseReg.ReplaceAllString(bodyStr, "$1=$2/exhibit/"+exhibit.Id+"/")
	bodyStr = hrefSrcReg.ReplaceAllString(bodyStr, "$1=$2/exhibit/"+exhibit.Id+"/$3$4")

	b, err := util.EncodeBody([]byte(bodyStr), encoding)
	if err != nil {
		return nil, err
	}

	return &b, nil
}

func (r *RewriteServiceImpl) RewriteClientRequest(exhibit domain.Exhibit, hostname string, req *http.Request, body *[]byte) error {
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
		h := strings.ReplaceAll(req.Header.Get(k), r.getFqhn()+"/exhibit/"+exhibit.Id, hostname)
		req.Header.Set(k, h)
	}

	*body, err = util.EncodeBody([]byte(bodyStr), encoding)
	if err != nil {
		r.Log.Warnw("error encoding body", "error", err, "requestId", exhibit.Id)
		return err
	}
	bodyStr = strings.ReplaceAll(bodyStr, r.getFqhn(), hostname)

	b, err := util.EncodeBody([]byte(bodyStr), encoding)
	if err != nil {
		return err
	}

	copy(*body, b)

	return nil
}
