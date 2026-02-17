package docker

import (
	"context"
	"log/slog"

	"github.com/docker/docker/client"
	"github.com/dockscope/dockscope/internal/domain"
)

type SystemInfoProvider struct {
	cli *client.Client
	log *slog.Logger
}

func NewSystemInfoProvider(cli *client.Client, log *slog.Logger) *SystemInfoProvider {
	return &SystemInfoProvider{cli: cli, log: log}
}

func (s *SystemInfoProvider) GetMemTotal(ctx context.Context) (uint64, error) {
	info, err := s.cli.Info(ctx)
	if err != nil {
		return 0, err
	}
	return uint64(info.MemTotal), nil
}

var _ domain.SystemInfoProvider = (*SystemInfoProvider)(nil)
