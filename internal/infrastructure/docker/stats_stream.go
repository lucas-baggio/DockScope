package docker

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/dockscope/dockscope/internal/domain"
)

type StatsStreamer struct {
	cli *client.Client
	log *slog.Logger
}

func NewStatsStreamer(cli *client.Client, log *slog.Logger) *StatsStreamer {
	return &StatsStreamer{cli: cli, log: log}
}

func (s *StatsStreamer) StreamStats(ctx context.Context, containerID string) (<-chan *domain.ContainerMetrics, <-chan error) {
	outCh := make(chan *domain.ContainerMetrics, 8)
	errCh := make(chan error, 1)

	go func() {
		defer close(outCh)
		defer close(errCh)

		statsResp, err := s.cli.ContainerStats(ctx, containerID, true)
		if err != nil {
			s.log.ErrorContext(ctx, "container stats stream failed", "container_id", containerID, "error", err)
			errCh <- err
			return
		}
		defer func() { _ = statsResp.Body.Close() }()

		dec := json.NewDecoder(statsResp.Body)

		for {
			select {
			case <-ctx.Done():
				s.log.DebugContext(ctx, "stats stream cancelled", "container_id", containerID)
				return
			default:
			}

			var raw types.StatsJSON
			if err := dec.Decode(&raw); err != nil {
				if err == io.EOF {
					return
				}
				s.log.DebugContext(ctx, "stats decode error (stream ended?)", "error", err)
				errCh <- err
				return
			}

			m := statsToMetrics(&raw)
			if m == nil {
				continue
			}

			select {
			case <-ctx.Done():
				return
			case outCh <- m:
			}
		}
	}()

	return outCh, errCh
}

func (s *StatsStreamer) Snapshot(ctx context.Context, containerID string) (*domain.ContainerMetrics, error) {
	statsResp, err := s.cli.ContainerStats(ctx, containerID, true)
	if err != nil {
		return nil, err
	}
	defer func() { _ = statsResp.Body.Close() }()

	dec := json.NewDecoder(statsResp.Body)
	var raw types.StatsJSON
	if err := dec.Decode(&raw); err != nil {
		return nil, err
	}
	return statsToMetrics(&raw), nil
}

func statsToMetrics(raw *types.StatsJSON) *domain.ContainerMetrics {
	cpuPercent := 0.0
	systemDelta := float64(raw.CPUStats.SystemUsage - raw.PreCPUStats.SystemUsage)
	if systemDelta > 0 {
		cpuDelta := float64(raw.CPUStats.CPUUsage.TotalUsage - raw.PreCPUStats.CPUUsage.TotalUsage)
		cpuPercent = (cpuDelta / systemDelta) * 100.0
		if raw.CPUStats.OnlineCPUs > 0 {
			cpuPercent /= float64(raw.CPUStats.OnlineCPUs)
		}
	}

	memUsage := raw.MemoryStats.Usage
	memLimit := raw.MemoryStats.Limit
	memPercent := 0.0
	if memLimit > 0 {
		memPercent = (float64(memUsage) / float64(memLimit)) * 100.0
	}

	return &domain.ContainerMetrics{
		CPUPercentage: round2(cpuPercent),
		MemoryUsage:    memUsage,
		MemoryLimit:    memLimit,
		MemoryPercent:  round2(memPercent),
		Timestamp:      raw.Read,
	}
}

func round2(v float64) float64 {
	return float64(int(v*100+0.5)) / 100
}
