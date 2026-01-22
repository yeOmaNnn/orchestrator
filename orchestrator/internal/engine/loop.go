package engine

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yeOmaNnn/orchestrator/internal/domain"
)

const schedulerInterval = 300 * time.Millisecond

func (e *Engine) RunTaskLoop(
	ctx context.Context,
	taskID uuid.UUID,
) error {

	if err := e.taskRepo.UpdateStatus(
		ctx,
		taskID,
		domain.TaskRunning,
	); err != nil {
		return err
	}

	ticker := time.NewTicker(schedulerInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()

		case <-ticker.C:

			if err := e.scheduler.Schedule(ctx, taskID); err != nil {
				_ = e.taskRepo.UpdateStatus(ctx, taskID, domain.TaskFailed)
				return err
			}

			steps, err := e.stepRepo.GetByTask(ctx, taskID)
			if err != nil {
				return err
			}

			var (
				hasActive bool
				hasError  bool
			)

			for _, s := range steps {
				switch s.Status {
				case domain.StepWaiting, domain.StepInProgress:
					hasActive = true
				case domain.StepError:
					hasError = true
				}
			}

			if hasError {
				_ = e.taskRepo.UpdateStatus(ctx, taskID, domain.TaskFailed)
				return nil
			}

			if !hasActive {
				_ = e.taskRepo.UpdateStatus(ctx, taskID, domain.TaskCompleted)
				return nil
			}
		}
	}
}
