package usecase

import (
	"context"
	"log/slog"

	"github.com/dockscope/dockscope/internal/domain"
)

type ListVolumes struct {
	repo domain.VolumeRepository
	log  *slog.Logger
}

func NewListVolumes(repo domain.VolumeRepository, log *slog.Logger) *ListVolumes {
	return &ListVolumes{repo: repo, log: log}
}

func (uc *ListVolumes) Execute(ctx context.Context) ([]*domain.Volume, error) {
	list, err := uc.repo.List(ctx)
	if err != nil {
		uc.log.ErrorContext(ctx, "list volumes use case failed", "error", err)
		return nil, err
	}
	uc.log.DebugContext(ctx, "list volumes ok", "count", len(list))
	return list, nil
}
