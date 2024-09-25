package impl

import (
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
