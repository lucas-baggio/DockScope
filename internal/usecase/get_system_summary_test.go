package usecase

import (
	"context"
	"log/slog"
	"os"
	"testing"

	"github.com/dockscope/dockscope/internal/domain"
)

type mockContainerRepo struct {
	list []*domain.Container
	err  error
}

func (m *mockContainerRepo) ListActive(ctx context.Context, all bool) ([]*domain.Container, error) {
	return m.list, m.err
}

type mockImageRepo struct {
	list []*domain.Image
	err  error
}

func (m *mockImageRepo) List(ctx context.Context) ([]*domain.Image, error) {
	return m.list, m.err
}

type mockVolumeRepo struct {
	list []*domain.Volume
	err  error
}

func (m *mockVolumeRepo) List(ctx context.Context) ([]*domain.Volume, error) {
	return m.list, m.err
}

type mockStatsStreamer struct {
	snapshot *domain.ContainerMetrics
	err      error
}

func (m *mockStatsStreamer) StreamStats(ctx context.Context, containerID string) (<-chan *domain.ContainerMetrics, <-chan error) {
	ch := make(chan *domain.ContainerMetrics, 1)
	errCh := make(chan error, 1)
	close(ch)
	close(errCh)
	return ch, errCh
}

func (m *mockStatsStreamer) Snapshot(ctx context.Context, containerID string) (*domain.ContainerMetrics, error) {
	return m.snapshot, m.err
}

type mockSysInfo struct {
	mem  uint64
	err  error
}

func (m *mockSysInfo) GetMemTotal(ctx context.Context) (uint64, error) {
	return m.mem, m.err
}

func TestGetSystemSummary_Empty(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	uc := NewGetSystemSummary(
		&mockContainerRepo{list: []*domain.Container{}},
		&mockImageRepo{list: []*domain.Image{}},
		&mockVolumeRepo{list: []*domain.Volume{}},
		&mockStatsStreamer{},
		&mockSysInfo{mem: 1024 * 1024 * 1024},
		log,
	)
	ctx := context.Background()

	out, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ContainersTotal != 0 || out.ContainersRunning != 0 || out.ContainersStopped != 0 {
		t.Errorf("expected zeros, got total=%d running=%d stopped=%d", out.ContainersTotal, out.ContainersRunning, out.ContainersStopped)
	}
	if out.ImagesCount != 0 || out.VolumesCount != 0 {
		t.Errorf("expected 0 images/volumes, got %d %d", out.ImagesCount, out.VolumesCount)
	}
	if out.MemoryLimitBytes != 1024*1024*1024 {
		t.Errorf("expected 1GB limit, got %d", out.MemoryLimitBytes)
	}
}

func TestGetSystemSummary_Counts(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	containers := []*domain.Container{
		{ID: "1", State: "running", Names: []string{"/a"}},
		{ID: "2", State: "exited", Names: []string{"/b"}},
	}
	images := []*domain.Image{{ID: "img1"}}
	volumes := []*domain.Volume{{Name: "v1"}}

	uc := NewGetSystemSummary(
		&mockContainerRepo{list: containers},
		&mockImageRepo{list: images},
		&mockVolumeRepo{list: volumes},
		&mockStatsStreamer{snapshot: &domain.ContainerMetrics{CPUPercentage: 1.5, MemoryUsage: 100}},
		&mockSysInfo{mem: 2048},
		log,
	)
	ctx := context.Background()

	out, err := uc.Execute(ctx)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if out.ContainersTotal != 2 {
		t.Errorf("containers_total: got %d", out.ContainersTotal)
	}
	if out.ContainersRunning != 1 || out.ContainersStopped != 1 {
		t.Errorf("running=%d stopped=%d", out.ContainersRunning, out.ContainersStopped)
	}
	if out.ImagesCount != 1 || out.VolumesCount != 1 {
		t.Errorf("images=%d volumes=%d", out.ImagesCount, out.VolumesCount)
	}
	if out.CPUPercentTotal != 1.5 || out.MemoryUsageBytes != 100 {
		t.Errorf("cpu=%f mem=%d", out.CPUPercentTotal, out.MemoryUsageBytes)
	}
	if len(out.ContainerMetrics) != 1 {
		t.Errorf("expected 1 container metric, got %d", len(out.ContainerMetrics))
	}
	if len(out.TopContainersByMemory) != 1 {
		t.Errorf("expected 1 top by memory, got %d", len(out.TopContainersByMemory))
	}
}

func TestGetSystemSummary_ContainerRepoError(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	uc := NewGetSystemSummary(
		&mockContainerRepo{err: context.DeadlineExceeded},
		&mockImageRepo{list: nil},
		&mockVolumeRepo{list: nil},
		&mockStatsStreamer{},
		&mockSysInfo{},
		log,
	)
	ctx := context.Background()

	_, err := uc.Execute(ctx)
	if err != context.DeadlineExceeded {
		t.Errorf("expected DeadlineExceeded, got %v", err)
	}
}
