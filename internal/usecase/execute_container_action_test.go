package usecase

import (
	"context"
	"errors"
	"log/slog"
	"os"
	"testing"
)

type mockController struct {
	lastID     string
	lastAction string
	err        error
}

func (m *mockController) ExecuteAction(ctx context.Context, containerID, action string) error {
	m.lastID = containerID
	m.lastAction = action
	return m.err
}

func TestExecuteContainerAction_Validation(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	ctrl := &mockController{}
	uc := NewExecuteContainerAction(ctrl, log)
	ctx := context.Background()

	t.Run("missing container id", func(t *testing.T) {
		err := uc.Execute(ctx, ExecuteContainerActionInput{ContainerID: "", Action: "start"})
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "missing container id" {
			t.Errorf("got %q", err.Error())
		}
		if ctrl.lastID != "" {
			t.Errorf("controller should not be called, got lastID %q", ctrl.lastID)
		}
	})

	t.Run("invalid action", func(t *testing.T) {
		err := uc.Execute(ctx, ExecuteContainerActionInput{ContainerID: "abc123", Action: "invalid"})
		if err == nil {
			t.Fatal("expected error")
		}
		if err.Error() != "invalid action: must be one of start, stop, restart, pause, unpause" {
			t.Errorf("got %q", err.Error())
		}
		if ctrl.lastID != "" {
			t.Errorf("controller should not be called, got lastID %q", ctrl.lastID)
		}
	})

	for _, action := range []string{"start", "stop", "restart", "pause", "unpause"} {
		t.Run("allowed_"+action, func(t *testing.T) {
			ctrl.err = nil
			err := uc.Execute(ctx, ExecuteContainerActionInput{ContainerID: "cid", Action: action})
			if err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if ctrl.lastID != "cid" || ctrl.lastAction != action {
				t.Errorf("controller called with %q, %q", ctrl.lastID, ctrl.lastAction)
			}
		})
	}
}

func TestExecuteContainerAction_ControllerError(t *testing.T) {
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: slog.LevelError}))
	ctrl := &mockController{err: errors.New("docker error")}
	uc := NewExecuteContainerAction(ctrl, log)
	ctx := context.Background()

	err := uc.Execute(ctx, ExecuteContainerActionInput{ContainerID: "cid", Action: "start"})
	if err == nil {
		t.Fatal("expected error from controller")
	}
	if err.Error() != "docker error" {
		t.Errorf("got %q", err.Error())
	}
}
