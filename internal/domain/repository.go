package domain

import (
	"context"
	"io"
)

type ContainerRepository interface {
	ListActive(ctx context.Context, all bool) ([]*Container, error)
}

type ImageRepository interface {
	List(ctx context.Context) ([]*Image, error)
}

type VolumeRepository interface {
	List(ctx context.Context) ([]*Volume, error)
}

type SystemInfoProvider interface {
	GetMemTotal(ctx context.Context) (uint64, error)
}

type ContainerStatsStreamer interface {
	StreamStats(ctx context.Context, containerID string) (<-chan *ContainerMetrics, <-chan error)
	Snapshot(ctx context.Context, containerID string) (*ContainerMetrics, error)
}

type ContainerLogsStreamer interface {
	StreamLogs(ctx context.Context, containerID string, w io.Writer) error
}

type ContainerController interface {
	ExecuteAction(ctx context.Context, containerID, action string) error
}

const (
	ActionStart   = "start"
	ActionStop    = "stop"
	ActionRestart = "restart"
	ActionPause   = "pause"
	ActionUnpause = "unpause"
)
