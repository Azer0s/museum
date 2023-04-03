package tool

import (
	"errors"
	"fmt"
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
	err, id := a.CreateExhibit(exhibit)
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	fmt.Println("🧑‍🎨 exhibit " + exhibit.Name + " created successfully")
	fmt.Println("👉 " + a.GetBaseUrl() + "/exhibits/" + id)

	return nil
}

func Delete() error {
	c := createToolContainer()

	if len(os.Args) < 3 {
		return errors.New("missing id argument")
	}

	a := ioc.Get[ApiClient](c)
	err := a.DeleteExhibitById(os.Args[2])
	if err != nil {
		fmt.Println(err.Error())
		return nil
	}

	fmt.Println("🗑️ exhibit deleted successfully")

	return nil
}

func List() error {
	//TODO implement me
	panic("implement me")
}
