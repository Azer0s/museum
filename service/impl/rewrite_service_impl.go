package impl

import (
	"fmt"
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

var placeholderHost = strings.ReplaceAll(uuid.New().String(), "-", "")

func replaceHostInString(str string, searchHost string, replaceHost string) string {
	// replace all occurrences of the searchHost with the replaceHost
	// don't replace the replaceHost with the replaceHost
	str = strings.ReplaceAll(str, replaceHost, placeholderHost)
	str = strings.ReplaceAll(str, searchHost, replaceHost)
	str = strings.ReplaceAll(str, placeholderHost, replaceHost)

	return str
}

func (r *RewriteServiceImpl) getSearchAndReplaceHost(exhibit domain.Exhibit) (string, string) {
	searchHost := r.Config.GetHostname() + ":" + r.Config.GetPort()
	replaceHost := r.Config.GetHostname() + ":" + r.Config.GetPort() + "/exhibit/" + exhibit.Id
	return searchHost, replaceHost
}

func (r *RewriteServiceImpl) RewriteServerResponse(exhibit domain.Exhibit, res *gohttp.Response, body *[]byte) error {
	// get encoding from header
	encoding := res.Header.Get("Content-Encoding")
	bodyDecoded, err := util.DecodeBody(*body, encoding)
	if err != nil {
		r.Log.Warnw("error decoding body", "error", err, "requestId", exhibit.Id)
		return err
	}

	bodyStr := string(bodyDecoded)

	//TODO: check if we have to rewrite the request headers from searchHost to replaceHost

	// let's rewrite some paths in the body
	searchHost, replaceHost := r.getSearchAndReplaceHost(exhibit)
	bodyStr = replaceHostInString(bodyStr, searchHost, replaceHost)

	*body, err = util.EncodeBody([]byte(bodyStr), encoding)
	if err != nil {
		r.Log.Warnw("error encoding body", "error", err, "requestId", exhibit.Id)
		return err
	}

	return nil
}

func (r *RewriteServiceImpl) RewriteClientRequest(exhibit domain.Exhibit, req *http.Request, body *[]byte) error {
	searchHost, replaceHost := r.getSearchAndReplaceHost(exhibit)

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

	bodyStr = replaceHostInString(bodyStr, replaceHost, searchHost)
	for k, v := range req.Header {
		req.Header.Set(k, replaceHostInString(strings.Join(v, ","), replaceHost, searchHost))
	}

	fmt.Println(bodyStr)

	return nil
}
