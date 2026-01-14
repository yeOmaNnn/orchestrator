package engine

import (
	"context"
	"orchestrator/internal/scheduler"

	"github.com/google/uuid"
	"orchestrator/internal/domain"
	"orchestrator/internal/planner"
	"orchestrator/internal/storage"
)

type Engine struct {
	scheduler *scheduler.Scheduler
	planner   planner.Client
	stepsRepo storage.StepRepository
}

func New(
	planner planner.Client,
	stepsRepo storage.StepRepository,
	scheduler *scheduler.Scheduler,
) *Engine {
	return &Engine{
		planner:   planner,
		stepsRepo: stepsRepo,
		scheduler: scheduler,
	}
}

func (e *Engine) InitTaskExecution(
	ctx context.Context,
	task domain.Task,
) error {

	plan, err := e.planner.CreatePlan(ctx, task.Goal)
	if err != nil {
		return err
	}

	steps, err := MapPlannedStepsToDomain(task.ID, plan.Steps)
	if err != nil {
		return err
	}

	return e.stepsRepo.CreateMany(ctx, steps)
}

func (e *Engine) RunTask(ctx context.Context, taskID uuid.UUID) error {
	return e.scheduler.Scheduler(ctx, taskID)
}
