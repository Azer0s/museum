package service

import (
	"context"
	docker "github.com/docker/docker/client"
	"go.uber.org/zap"
	"museum/config"
)

func NewDockerClient(config config.Config, log *zap.SugaredLogger) *docker.Client {
	ctx := context.Background()
	c, err := docker.NewClientWithOpts(docker.WithHost(config.GetDockerHost()))
	if err != nil {
		log.Panicw("failed to create docker client", "error", err)
	}

	c.NegotiateAPIVersion(ctx)

	info, err := c.Info(ctx)
	if err != nil {
		log.Panicw("failed to get docker info", "error", err)
	}

	if info.Swarm.LocalNodeState != "active" {
		log.Panic("docker swarm is not active")
	}

	log.Debugw("connected to docker", "host", config.GetDockerHost())

	return c
}
