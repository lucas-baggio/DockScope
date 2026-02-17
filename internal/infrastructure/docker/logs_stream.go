package docker

import (
	"context"
	"io"
	"log/slog"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
)

type LogsStreamer struct {
	cli *client.Client
	log *slog.Logger
}

func NewLogsStreamer(cli *client.Client, log *slog.Logger) *LogsStreamer {
	return &LogsStreamer{cli: cli, log: log}
}

func (s *LogsStreamer) StreamLogs(ctx context.Context, containerID string, w io.Writer) error {
	opts := types.ContainerLogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     true,
		Timestamps: false,
	}
	body, err := s.cli.ContainerLogs(ctx, containerID, opts)
	if err != nil {
		s.log.ErrorContext(ctx, "container logs failed", "container_id", containerID, "error", err)
		return err
	}
	defer body.Close()

	_, err = stdcopy.StdCopy(w, w, body)
	if err != nil && ctx.Err() == nil {
		s.log.DebugContext(ctx, "container logs stream ended", "container_id", containerID, "error", err)
		return err
	}
	return nil
}
