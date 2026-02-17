package usecase

import (
	"context"
	"log/slog"
	"sync"

	"github.com/dockscope/dockscope/internal/domain"
)

type ContainerMemoryEntry struct {
	ID           string  `json:"id"`
	Name         string  `json:"name"`
	MemoryUsage  uint64  `json:"memory_usage"`
	MemoryPercent float64 `json:"memory_percent,omitempty"`
}

type ContainerMetricsEntry struct {
	ID              string  `json:"id"`
	Name            string  `json:"name"`
	CPUPercentage   float64 `json:"cpu_percentage"`
	MemoryUsage     uint64  `json:"memory_usage"`
	MemoryLimit     uint64  `json:"memory_limit"`
	MemoryPercent   float64 `json:"memory_percent,omitempty"`
}

type GetSystemSummaryOutput struct {
	ContainersTotal    int                     `json:"containers_total"`
	ContainersRunning  int                     `json:"containers_running"`
	ContainersStopped  int                     `json:"containers_stopped"`
	CPUPercentTotal    float64                 `json:"cpu_percent_total"`
	MemoryUsageBytes   uint64                  `json:"memory_usage_bytes"`
	MemoryLimitBytes   uint64                  `json:"memory_limit_bytes"`
	ImagesCount        int                     `json:"images_count"`
	VolumesCount       int                     `json:"volumes_count"`
	TopContainersByMemory []ContainerMemoryEntry `json:"top_containers_by_memory"`
	ContainerMetrics   []ContainerMetricsEntry `json:"container_metrics"`
}

const topContainersByMemoryN = 10

type GetSystemSummary struct {
	containers domain.ContainerRepository
	images     domain.ImageRepository
	volumes    domain.VolumeRepository
	stats      domain.ContainerStatsStreamer
	sysInfo    domain.SystemInfoProvider
	log        *slog.Logger
}

func NewGetSystemSummary(
	containers domain.ContainerRepository,
	images domain.ImageRepository,
	volumes domain.VolumeRepository,
	stats domain.ContainerStatsStreamer,
	sysInfo domain.SystemInfoProvider,
	log *slog.Logger,
) *GetSystemSummary {
	return &GetSystemSummary{
		containers: containers,
		images:     images,
		volumes:    volumes,
		stats:      stats,
		sysInfo:    sysInfo,
		log:        log,
	}
}

func (uc *GetSystemSummary) Execute(ctx context.Context) (*GetSystemSummaryOutput, error) {
	containers, err := uc.containers.ListActive(ctx, true)
	if err != nil {
		return nil, err
	}
	images, err := uc.images.List(ctx)
	if err != nil {
		return nil, err
	}
	volumes, err := uc.volumes.List(ctx)
	if err != nil {
		return nil, err
	}
	memTotal, err := uc.sysInfo.GetMemTotal(ctx)
	if err != nil {
		uc.log.WarnContext(ctx, "could not get host mem total, using sum of containers", "error", err)
		memTotal = 0
	}

	out := &GetSystemSummaryOutput{
		ContainersTotal:   len(containers),
		ImagesCount:       len(images),
		VolumesCount:      len(volumes),
		MemoryLimitBytes:  memTotal,
		TopContainersByMemory: make([]ContainerMemoryEntry, 0, topContainersByMemoryN),
		ContainerMetrics:  make([]ContainerMetricsEntry, 0),
	}

	var running []*domain.Container
	for _, c := range containers {
		switch c.State {
		case "running":
			out.ContainersRunning++
			running = append(running, c)
		default:
			out.ContainersStopped++
		}
	}

	if len(running) == 0 {
		return out, nil
	}

	type result struct {
		c   *domain.Container
		m   *domain.ContainerMetrics
		err error
	}
	ch := make(chan result, len(running))
	var wg sync.WaitGroup
	for _, c := range running {
		c := c
		wg.Add(1)
		go func() {
			defer wg.Done()
			m, err := uc.stats.Snapshot(ctx, c.ID)
			ch <- result{c: c, m: m, err: err}
		}()
	}
	go func() {
		wg.Wait()
		close(ch)
	}()

	var totalCPU float64
	var totalMem uint64
	var metricsWithMem []struct {
		c *domain.Container
		m *domain.ContainerMetrics
	}
	for r := range ch {
		if r.err != nil {
			uc.log.DebugContext(ctx, "stats snapshot failed", "container_id", r.c.ID, "error", r.err)
			continue
		}
		if r.m == nil {
			continue
		}
		totalCPU += r.m.CPUPercentage
		totalMem += r.m.MemoryUsage
		metricsWithMem = append(metricsWithMem, struct {
			c *domain.Container
			m *domain.ContainerMetrics
		}{r.c, r.m})
		out.ContainerMetrics = append(out.ContainerMetrics, ContainerMetricsEntry{
			ID:            r.c.ID,
			Name:          containerDisplayName(r.c),
			CPUPercentage:  r.m.CPUPercentage,
			MemoryUsage:   r.m.MemoryUsage,
			MemoryLimit:   r.m.MemoryLimit,
			MemoryPercent: r.m.MemoryPercent,
		})
	}

	out.CPUPercentTotal = totalCPU
	out.MemoryUsageBytes = totalMem
	if memTotal == 0 && totalMem > 0 {
		out.MemoryLimitBytes = totalMem
	}

	for i := 0; i < len(metricsWithMem); i++ {
		for j := i + 1; j < len(metricsWithMem); j++ {
			if metricsWithMem[j].m.MemoryUsage > metricsWithMem[i].m.MemoryUsage {
				metricsWithMem[i], metricsWithMem[j] = metricsWithMem[j], metricsWithMem[i]
			}
		}
	}
	for i := 0; i < len(metricsWithMem) && i < topContainersByMemoryN; i++ {
		c, m := metricsWithMem[i].c, metricsWithMem[i].m
		percent := 0.0
		if out.MemoryLimitBytes > 0 {
			percent = (float64(m.MemoryUsage) / float64(out.MemoryLimitBytes)) * 100
		}
		out.TopContainersByMemory = append(out.TopContainersByMemory, ContainerMemoryEntry{
			ID:            c.ID,
			Name:          containerDisplayName(c),
			MemoryUsage:   m.MemoryUsage,
			MemoryPercent: percent,
		})
	}

	return out, nil
}

func containerDisplayName(c *domain.Container) string {
	if len(c.Names) > 0 && c.Names[0] != "" {
		name := c.Names[0]
		if len(name) > 0 && name[0] == '/' {
			name = name[1:]
		}
		return name
	}
	if len(c.ID) >= 12 {
		return c.ID[:12]
	}
	return c.ID
}
