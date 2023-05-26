package impl

import (
	"museum/domain"
	service "museum/service/interface"
	"net/http"
	"strconv"
	"strings"
)

type HttpLivecheck struct {
	ApplicationResolverService service.ApplicationResolverService
	ExhibitService             service.ExhibitService
}

func (h *HttpLivecheck) Check(exhibit domain.Exhibit, object domain.Object) (retry bool, err error) {
	exhibit.RuntimeInfo.Status = domain.Running

	ip, err := h.ApplicationResolverService.ResolveExhibitObject(exhibit, object)
	if err != nil {
		retry = false
		return
	}

	// do http request to ip
	port := "80"
	if object.Port != nil {
		port = *object.Port
	}

	method, ok := object.Livecheck.Config["method"]
	if !ok {
		method = "GET"
	}

	path, ok := object.Livecheck.Config["path"]
	if !ok {
		path = "/"
	}

	req, err := http.NewRequest(strings.ToUpper(method), "http://"+ip+":"+port+path, nil)
	if err != nil {
		retry = false
		return
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		retry = true
		return
	}

	statusStr, ok := object.Livecheck.Config["status"]
	if !ok {
		statusStr = "200"
	}

	status, err := strconv.Atoi(statusStr)
	if err != nil {
		retry = false
		return
	}

	if res.StatusCode != status {
		retry = true
		return
	}

	retry = false
	return
}
