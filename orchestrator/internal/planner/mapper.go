package planner

import (
	"github.com/google/uuid"
	"github.com/yeOmaNnn/orchestrator/internal/domain"
)

func MapToDomainSteps(
	taskID uuid.UUID, 
	planned []PlannedStep, 
) []domain.Step {
	steps := make([]domain.Step, 0, len(planned))

	for _, ps := range planned {
		step :=  domain.Step{
			ID: ps.ID,
			TaskID: taskID,
			Agent: ps.Agent,
			Input: ps.Input,
			Status: domain.StepWaiting,
			RetryCount: 0,
			DependsOn: ps.DependsOn,
		}
		steps = append(steps, step)
	}
	return steps
}