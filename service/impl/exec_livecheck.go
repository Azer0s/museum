package impl

import (
	"museum/domain"
	"os/exec"
)

type ExecLivecheck struct {
}

func (e *ExecLivecheck) Check(exhibit domain.Exhibit, object domain.Object) (retry bool, err error) {
	objectContainerName := exhibit.Name + "_" + object.Name
	command, ok := object.Livecheck.Config["command"]

	if !ok {
		command = "true"
	}

	dockerCommand := []string{"docker", "exec", objectContainerName, "sh", "-c", command}
	_, err = exec.Command(dockerCommand[0], dockerCommand[1:]...).Output()
	if err != nil {
		// if the error is a exit code error, we can check the exit code
		// if it is not, we can't do anything
		if exitError, ok := err.(*exec.ExitError); ok {
			if exitError.ExitCode() != 0 {
				retry = true
				err = nil
				return
			}
		}

		return false, err
	}

	return false, nil
}
