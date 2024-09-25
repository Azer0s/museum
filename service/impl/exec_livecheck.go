package impl

import (
	"context"
	"github.com/docker/docker/api/types/container"
	docker "github.com/docker/docker/client"
	"io"
	"museum/domain"
)

type ExecLivecheck struct {
	Client *docker.Client
}

func (e *ExecLivecheck) Check(ctx context.Context, exhibit domain.Exhibit, object domain.Object) (retry bool, err error) {
	objectContainerName := exhibit.Name + "_" + object.Name
	command, ok := object.Livecheck.Config["command"]

	if !ok {
		command = "true"
	}

	exec, err := e.Client.ContainerExecCreate(ctx, objectContainerName, container.ExecOptions{
		Cmd: []string{"sh", "-c", command},
	})
	if err != nil {
		return false, err
	}

	res, err := e.Client.ContainerExecAttach(ctx, exec.ID, container.ExecStartOptions{})
	if err != nil {
		return false, err
	}

	_, err = io.ReadAll(res.Reader)
	if err != nil {
		return false, err
	}

	res.Close()

	inspect, err := e.Client.ContainerExecInspect(ctx, exec.ID)
	if err != nil {
		return false, err
	}

	if inspect.ExitCode != 0 {
		return true, nil
	}

	return false, nil
}
