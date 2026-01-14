package engine

import (
	"encoding/json"

	"github.com/google/uuid"

	"orchestrator/internal/domain"
	"orchestrator/internal/planner"
)

func MapPlannedStepsToDomain(
	taskID uuid.UUID,
	planned []planner.PlannedStep,
) ([]domain.Step, error) {

	idMap := make(map[string]uuid.UUID)
	steps := make([]domain.Step, 0, len(planned))

	for _, p := range planned {
		idMap[p.ID] = uuid.New()
	}

	for _, p := range planned {
		inputBytes, err := json.Marshal(p.Input)
		if err != nil {
			return nil, err
		}

		deps := make([]uuid.UUID, 0, len(p.DependsOn))
		for _, dep := range p.DependsOn {
			deps = append(deps, idMap[dep])
		}

		step := domain.Step{
			ID:         idMap[p.ID],
			TaskID:     taskID,
			Agent:      p.Agent,
			Input:      inputBytes,
			Status:     domain.StepWaiting,
			DependsOn:  deps,
			RetryCount: 0,
		}
		steps = append(steps, step)
	}
	return steps, nil
}
