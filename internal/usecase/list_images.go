package usecase

import (
	"context"
	"log/slog"

	"github.com/dockscope/dockscope/internal/domain"
)

type ListImages struct {
	repo domain.ImageRepository
	log  *slog.Logger
}

func NewListImages(repo domain.ImageRepository, log *slog.Logger) *ListImages {
	return &ListImages{repo: repo, log: log}
}

func (uc *ListImages) Execute(ctx context.Context) ([]*domain.Image, error) {
	list, err := uc.repo.List(ctx)
	if err != nil {
		uc.log.ErrorContext(ctx, "list images use case failed", "error", err)
		return nil, err
	}
	uc.log.DebugContext(ctx, "list images ok", "count", len(list))
	return list, nil
}
