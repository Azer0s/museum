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

func (r *RewriteServiceImpl) RewriteServerResponse(exhibit domain.Exhibit, res *gohttp.Response, body *[]byte) error {
	// get encoding from header
	encoding := res.Header.Get("Content-Encoding")
	bodyDecoded, err := util.DecodeBody(*body, encoding)
	if err != nil {
		r.Log.Warnw("error decoding body", "error", err, "requestId", exhibit.Id)
		return err
	}

	bodyStr := string(bodyDecoded)

	// let's rewrite some paths in the body
	searchHost := r.Config.GetHostname() + ":" + r.Config.GetPort()
	replaceHost := r.Config.GetHostname() + ":" + r.Config.GetPort() + "/exhibit/" + exhibit.Id

	// replace all occurrences of the searchHost with the replaceHost
	// don't replace the replaceHost with the replaceHost
	placeholderHost := uuid.New().String()
	bodyStr = strings.ReplaceAll(bodyStr, replaceHost, placeholderHost)
	bodyStr = strings.ReplaceAll(bodyStr, searchHost, replaceHost)
	bodyStr = strings.ReplaceAll(bodyStr, placeholderHost, replaceHost)

	*body, err = util.EncodeBody([]byte(bodyStr), encoding)
	if err != nil {
		r.Log.Warnw("error encoding body", "error", err, "requestId", exhibit.Id)
		return err
	}

	return nil
}

func (r *RewriteServiceImpl) RewriteClientRequest(exhibit domain.Exhibit, req *http.Request, body *[]byte) error {
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

	fmt.Println(bodyStr)

	referer := req.Header.Get("Referer")
	if referer != "" {
		//TODO: rewrite referer
	}

	return nil
}
