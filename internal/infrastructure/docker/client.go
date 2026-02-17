package docker

import (
	"context"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/docker/docker/client"
)

const (
	DefaultDockerHost = "unix:///var/run/docker.sock"
	EnvDockerHost     = "DOCKER_HOST"
)

func NewClient(ctx context.Context, log *slog.Logger) (*client.Client, error) {
	host := os.Getenv(EnvDockerHost)
	if host == "" {
		host = DefaultDockerHost
	}
	if len(host) > 7 && host[:7] == "unix://" {
		if path := host[7:]; path != "" && path[0] != '/' {
			abs, err := filepath.Abs(path)
			if err != nil {
				log.WarnContext(ctx, "resolving docker socket path", "path", path, "error", err)
			} else {
				host = "unix://" + abs
			}
		}
	}
	cli, err := client.NewClientWithOpts(
		client.WithHost(host),
		client.WithAPIVersionNegotiation(),
	)
	if err != nil {
		log.ErrorContext(ctx, "docker client creation failed", "host", host, "error", err)
		return nil, err
	}
	if _, err := cli.Ping(ctx); err != nil {
		_ = cli.Close()
		log.ErrorContext(ctx, "docker daemon ping failed", "host", host, "error", err)
		return nil, err
	}
	log.InfoContext(ctx, "docker client connected", "host", host)
	return cli, nil
}
