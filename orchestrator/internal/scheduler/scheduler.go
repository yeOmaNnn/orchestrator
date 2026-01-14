package scheduler

import (
	"context"
	"orchestrator/internal/domain"
	"orchestrator/internal/storage"

	"github.com/google/uuid"
	"orchestrator/internal/runner"
	"honnef.co/go/tools/lintcmd/runner"
)

type Scheduler struct {
	stepsRepo  storage.StepRepository
	runner *runner.Runner
}

func New(stepsRepo storage.StepRepository, 
		 runner *runner.Runner,) *Scheduler {
	return &Scheduler{stepsRepo: stepsRepo,
					  runner: runner,}
}

func (s *Scheduler) Scheduler(ctx context.Context, taskID uuid.UUID) error {
	steps, err := s.stepsRepo.GetByTask(ctx, taskID)
	if err != nil {
		return err 
	}

	stepByID := make(map[*uuid.UUID]domain.Step, len(steps))
	for _, step := range steps {
		stepByID[step.ID] = step
	}

		var readySteps []domain.Step

	for _, step := range steps {
		if step.Status != domain.StepWaiting {
			continue
		}

		if dependenciesDone(step, stepByID) {
			readySteps = append(readySteps, step)
		}
	}

		if len(readySteps) == 0 {
		return nil
	}

		for i := range readySteps {
		readySteps[i].Status = domain.StepInProgress

		if err := s.stepsRepo.Update(ctx, &readySteps[i]); err != nil {
			return err
		}
	}

		for _, step := range readySteps {
		if err := s.runner.Run(ctx, step); err != nil {
			return err
		}
	}
	return nil
}

func dependenciesDone(
	step domain.Step,
	all map[uuid.UUID]domain.Step,
) bool {

	for _, depID := range step.DependsOn {
		dep, ok := all[depID]
		if !ok {
			return false
		}

		if dep.Status != domain.StepDone {
			return false
		}
	}

	return true
}
