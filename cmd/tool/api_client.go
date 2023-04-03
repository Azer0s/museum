package tool

import (
	"bytes"
	"encoding/json"
	"errors"
	"museum/domain"
	"net/http"
)

type ApiClient interface {
	CreateExhibit(exhibit *domain.Exhibit) (error, string)
	DeleteExhibitById(id string) error
	GetBaseUrl() string
}

type ApiClientImpl struct {
	BaseUrl string
}

func (a *ApiClientImpl) CreateExhibit(exhibit *domain.Exhibit) (error, string) {
	b, err := json.Marshal(exhibit)
	if err != nil {
		return err, ""
	}

	res, err := http.Post(a.BaseUrl+"/api/exhibits", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err, ""
	}

	status := make(map[string]string)
	err = json.NewDecoder(res.Body).Decode(&status)
	if err != nil {
		return err, ""
	}

	if res.StatusCode != http.StatusCreated {
		return errors.New("could not create exhibit: " + status["error"]), ""
	}

	return nil, status["id"]
}

func (a *ApiClientImpl) DeleteExhibitById(id string) error {
	req, err := http.NewRequest(http.MethodDelete, a.BaseUrl+"/api/exhibits/"+id, nil)
	if err != nil {
		return err
	}

	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	status := make(map[string]string)
	err = json.NewDecoder(res.Body).Decode(&status)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusNoContent {
		return errors.New("could not delete exhibit: " + status["error"])
	}

	return nil
}

func (a *ApiClientImpl) GetBaseUrl() string {
	return a.BaseUrl
}
