package docker

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/dockscope/dockscope/internal/domain"
)

type ContainerController struct {
	cli *client.Client
	log *slog.Logger
}

func NewContainerController(cli *client.Client, log *slog.Logger) *ContainerController {
	return &ContainerController{cli: cli, log: log}
}

func (c *ContainerController) ExecuteAction(ctx context.Context, containerID, action string) error {
	timeout := 10 * time.Second
	switch action {
	case domain.ActionStart:
		return c.cli.ContainerStart(ctx, containerID, types.ContainerStartOptions{})
	case domain.ActionStop:
		return c.cli.ContainerStop(ctx, containerID, &timeout)
	case domain.ActionRestart:
		return c.cli.ContainerRestart(ctx, containerID, &timeout)
	case domain.ActionPause:
		return c.cli.ContainerPause(ctx, containerID)
	case domain.ActionUnpause:
		return c.cli.ContainerUnpause(ctx, containerID)
	default:
		return errors.New("invalid action")
	}
}
