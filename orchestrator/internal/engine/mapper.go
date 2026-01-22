package engine

import (
	"github.com/google/uuid"

	"github.com/yeOmaNnn/orchestrator/internal/domain"
	"github.com/yeOmaNnn/orchestrator/internal/planner"
)

func MapPlannedStepsToDomain(
	taskID uuid.UUID,
	planned []planner.PlannedStep,
) ([]domain.Step, error) {

	steps := make([]domain.Step, 0, len(planned))

	for _, p := range planned {
		step := domain.Step{
			ID:         p.ID,
			TaskID:     taskID,
			Agent:      p.Agent,
			Input:      p.Input,
			Status:     domain.StepWaiting,
			DependsOn:  p.DependsOn,
			RetryCount: 0,
		}

		steps = append(steps, step)
	}

	return steps, nil
}
