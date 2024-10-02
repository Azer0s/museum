package impl

import (
	"context"
	"errors"
	"museum/domain"
	"os"
)

type LocalVolumeProvisionerService struct {
}

func (l LocalVolumeProvisionerService) CheckValidity(config domain.StringMap) error {
	path, ok := config["path"]
	if !ok {
		return errors.New("path is required")
	}

	if path == "" {
		return errors.New("path cannot be empty")
	}

	//check if path exists
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return errors.New("path does not exist")
	}

	return nil
}

func (l LocalVolumeProvisionerService) ProvisionStorage(_ context.Context, config domain.StringMap) (string, error) {
	if _, err := os.Stat(config["path"]); os.IsNotExist(err) {
		return "", errors.New("path does not exist")
	}

	// check access rights
	if err := os.WriteFile(config["path"]+"/.museum", []byte("museum"), 0644); err != nil {
		return "", err
	}

	err := os.Remove(config["path"] + "/.museum")
	if err != nil {
		return "", err
	}

	return config["path"], nil
}

func (l LocalVolumeProvisionerService) DeprovisionStorage(context.Context, domain.StringMap) error {
	return nil
}
