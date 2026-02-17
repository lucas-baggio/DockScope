package usecase

import (
	"context"
	"log/slog"

	"github.com/dockscope/dockscope/internal/domain"
)

type StreamContainerLogs struct {
	streamer domain.ContainerLogsStreamer
	log      *slog.Logger
}

func NewStreamContainerLogs(streamer domain.ContainerLogsStreamer, log *slog.Logger) *StreamContainerLogs {
	return &StreamContainerLogs{streamer: streamer, log: log}
}

func (uc *StreamContainerLogs) Execute(ctx context.Context, containerID string, w interface {
	Write([]byte) (int, error)
}) error {
	return uc.streamer.StreamLogs(ctx, containerID, w)
}
