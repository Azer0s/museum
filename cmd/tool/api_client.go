package tool

import (
	"bytes"
	"encoding/json"
	"io"
	"museum/domain"
	"net/http"
)

type ApiClient interface {
	CreateExhibit(exhibit *domain.Exhibit) error
}

type ApiClientImpl struct {
	BaseUrl string
}

func (a *ApiClientImpl) CreateExhibit(exhibit *domain.Exhibit) error {
	b, err := json.Marshal(exhibit)
	if err != nil {
		return err
	}

	res, err := http.Post(a.BaseUrl+"/api/exhibits", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	b, err = io.ReadAll(res.Body)

	return nil
}
