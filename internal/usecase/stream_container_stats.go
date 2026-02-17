package usecase

import (
	"context"
	"log/slog"

	"github.com/dockscope/dockscope/internal/domain"
)

type MetricsWriter interface {
	WriteJSON(v interface{}) error
}

func (uc *StreamContainerStats) Execute(ctx context.Context, containerID string, sink MetricsWriter) error {
	metricsCh, errCh := uc.streamer.StreamStats(ctx, containerID)

	for errCh != nil {
		select {
		case <-ctx.Done():
			uc.log.DebugContext(ctx, "stream stats use case: context cancelled", "container_id", containerID)
			return ctx.Err()
		case err, ok := <-errCh:
			if ok && err != nil {
				uc.log.ErrorContext(ctx, "stream stats error", "container_id", containerID, "error", err)
				return err
			}
			if !ok {
				errCh = nil
			}
		case m, ok := <-metricsCh:
			if !ok {
				return nil
			}
			if err := sink.WriteJSON(m); err != nil {
				uc.log.DebugContext(ctx, "stream stats: sink write failed (client gone?)", "error", err)
				return err
			}
		}
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case m, ok := <-metricsCh:
			if !ok {
				return nil
			}
			if err := sink.WriteJSON(m); err != nil {
				return err
			}
		}
	}
}

type StreamContainerStats struct {
	streamer domain.ContainerStatsStreamer
	log      *slog.Logger
}

func NewStreamContainerStats(streamer domain.ContainerStatsStreamer, log *slog.Logger) *StreamContainerStats {
	return &StreamContainerStats{streamer: streamer, log: log}
}

type jsonWriter struct {
	encode func(v interface{}) error
}

func (w *jsonWriter) WriteJSON(v interface{}) error {
	return w.encode(v)
}

func NewJSONMetricsWriter(encode func(v interface{}) error) MetricsWriter {
	return &jsonWriter{encode: encode}
}
