package tool

import (
	"errors"
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

func Create() error {
	c := createToolContainer()

	if len(os.Args) < 3 {
		return errors.New("missing file argument")
	}

	_, err := os.Open(os.Args[2])
	if err != nil {
		return err
	}
	content, err := os.ReadFile(os.Args[2])
	if err != nil {
		return err
	}

	exhibit := &domain.Exhibit{}
	err = yaml.Unmarshal(content, exhibit)

	if err != nil {
		return err
	}

	a := ioc.Get[ApiClient](c)
	err = a.CreateExhibit(exhibit)
	if err != nil {
		return err
	}

	return nil
}

func Delete() error {
	//TODO implement me
	panic("implement me")

	/*isId := false
	f, err := os.Open(os.Args[2])
	if err != nil {
		if !os.IsNotExist(err) {
			return err
		}

		isId = true
	}*/
}

func List() error {
	//TODO implement me
	panic("implement me")
}
