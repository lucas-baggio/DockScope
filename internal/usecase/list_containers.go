package usecase

import (
	"context"
	"log/slog"

	"github.com/dockscope/dockscope/internal/domain"
)

type ListContainersInput struct {
	All bool
}

type ListContainers struct {
	repo domain.ContainerRepository
	log  *slog.Logger
}

func NewListContainers(repo domain.ContainerRepository, log *slog.Logger) *ListContainers {
	return &ListContainers{repo: repo, log: log}
}

func (uc *ListContainers) Execute(ctx context.Context, input ListContainersInput) ([]*domain.Container, error) {
	list, err := uc.repo.ListActive(ctx, input.All)
	if err != nil {
		uc.log.ErrorContext(ctx, "list containers use case failed", "error", err)
		return nil, err
	}
	uc.log.DebugContext(ctx, "list containers ok", "count", len(list))
	return list, nil
}
