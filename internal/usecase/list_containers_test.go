package usecase

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"

	"github.com/dockscope/dockscope/internal/domain"
)

type mockContainerRepository struct {
	list []*domain.Container
	err  error
}

func (m *mockContainerRepository) ListActive(ctx context.Context, all bool) ([]*domain.Container, error) {
	return m.list, m.err
}

func TestListContainers_Execute(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	ctx := context.Background()

	t.Run("returns list from repo", func(t *testing.T) {
		want := []*domain.Container{
			{ID: "a", State: "running"},
			{ID: "b", State: "exited"},
		}
		repo := &mockContainerRepository{list: want}
		uc := NewListContainers(repo, log)

		got, err := uc.Execute(ctx, ListContainersInput{All: true})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(got) != 2 {
			t.Fatalf("len(got)=%d", len(got))
		}
		if got[0].ID != "a" || got[1].ID != "b" {
			t.Errorf("got %v", got)
		}
	})

	t.Run("propagates repo error", func(t *testing.T) {
		repoErr := errors.New("repo failed")
		repo := &mockContainerRepository{err: repoErr}
		uc := NewListContainers(repo, log)

		_, err := uc.Execute(ctx, ListContainersInput{All: false})
		if err != repoErr {
			t.Errorf("got err %v", err)
		}
	})
}
