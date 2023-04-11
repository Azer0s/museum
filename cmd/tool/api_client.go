package tool

import (
	"bytes"
	"encoding/json"
	"errors"
	cloudevents "github.com/cloudevents/sdk-go/v2/event"
	"museum/domain"
	"net/http"
)

type ApiClient interface {
	CreateExhibit(exhibit *domain.Exhibit) (string, error)
	DeleteExhibitById(id string) error
	CreateEvent(event *cloudevents.Event) error
	GetBaseUrl() string
	GetExhibitById(id string) (*domain.Exhibit, error)
}

type ApiClientImpl struct {
	BaseUrl string
}

func (a *ApiClientImpl) CreateExhibit(exhibit *domain.Exhibit) (string, error) {
	b, err := json.Marshal(exhibit)
	if err != nil {
		return "", err
	}

	res, err := http.Post(a.BaseUrl+"/api/exhibits", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return "", err
	}

	status := make(map[string]string)
	err = json.NewDecoder(res.Body).Decode(&status)
	if err != nil {
		return "", err
	}

	if res.StatusCode != http.StatusCreated {
		return "", errors.New("could not create exhibit: " + status["error"])
	}

	return status["id"], nil
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

func (a *ApiClientImpl) CreateEvent(event *cloudevents.Event) error {
	b, err := event.MarshalJSON()
	if err != nil {
		return err
	}

	res, err := http.Post(a.BaseUrl+"/api/events", "application/json", bytes.NewBuffer(b))
	if err != nil {
		return err
	}

	status := make(map[string]string)
	err = json.NewDecoder(res.Body).Decode(&status)
	if err != nil {
		return err
	}

	if res.StatusCode != http.StatusCreated {
		return errors.New("could not create event: " + status["error"])
	}

	return nil
}

func (a *ApiClientImpl) GetExhibitById(id string) (*domain.Exhibit, error) {
	res, err := http.Get(a.BaseUrl + "/api/exhibits/" + id)
	if err != nil {
		return nil, err
	}

	exhibit := &domain.Exhibit{}
	err = json.NewDecoder(res.Body).Decode(exhibit)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("could not get exhibit")
	}

	return exhibit, nil
}

func (a *ApiClientImpl) GetBaseUrl() string {
	return a.BaseUrl
}
