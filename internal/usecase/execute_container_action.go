package usecase

import (
	"context"
	"errors"
	"log/slog"

	"github.com/dockscope/dockscope/internal/domain"
)

var allowedActions = map[string]bool{
	domain.ActionStart:   true,
	domain.ActionStop:    true,
	domain.ActionRestart: true,
	domain.ActionPause:   true,
	domain.ActionUnpause: true,
}

type ExecuteContainerActionInput struct {
	ContainerID string
	Action      string
}

type ExecuteContainerAction struct {
	ctrl domain.ContainerController
	log  *slog.Logger
}

func NewExecuteContainerAction(ctrl domain.ContainerController, log *slog.Logger) *ExecuteContainerAction {
	return &ExecuteContainerAction{ctrl: ctrl, log: log}
}

func (uc *ExecuteContainerAction) Execute(ctx context.Context, input ExecuteContainerActionInput) error {
	if input.ContainerID == "" {
		return errors.New("missing container id")
	}
	if !allowedActions[input.Action] {
		return errors.New("invalid action: must be one of start, stop, restart, pause, unpause")
	}
	err := uc.ctrl.ExecuteAction(ctx, input.ContainerID, input.Action)
	if err != nil {
		uc.log.ErrorContext(ctx, "container action failed", "container_id", input.ContainerID, "action", input.Action, "error", err)
		return err
	}
	uc.log.InfoContext(ctx, "container action ok", "container_id", input.ContainerID, "action", input.Action)
	return nil
}
