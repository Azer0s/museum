package tool

import (
	"gopkg.in/yaml.v3"
	"museum/domain"
	"museum/ioc"
	"os"
)

func createToolContainer() *ioc.Container {
	c := ioc.NewContainer()
	ioc.RegisterSingleton[ApiClient](c, func() ApiClient {
		return &ApiClientImpl{
			BaseUrl: "http://localhost:8080",
		}
	})
	return c
}

func Create(filePath string) (error, *domain.Exhibit, string) {
	c := createToolContainer()

	_, err := os.Open(filePath)
	if err != nil {
		return err, nil, ""
	}
	content, err := os.ReadFile(filePath)
	if err != nil {
		return err, nil, ""
	}

	exhibit := &domain.Exhibit{}
	err = yaml.Unmarshal(content, exhibit)

	if err != nil {
		return err, nil, ""
	}

	a := ioc.Get[ApiClient](c)
	err, id := a.CreateExhibit(exhibit)
	if err != nil {
		return err, nil, ""
	}

	exhibit.Id = id

	return nil, exhibit, a.GetBaseUrl() + "/exhibits/" + id
}

func Delete(id string) error {
	c := createToolContainer()

	a := ioc.Get[ApiClient](c)
	err := a.DeleteExhibitById(id)
	if err != nil {
		return err
	}

	return nil
}

func List() (error, []domain.Exhibit) {
	//TODO implement me
	panic("implement me")
}

func Warmup(id string) error {
	//TODO implement me
	panic("implement me")
}
