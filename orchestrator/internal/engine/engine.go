package engine

import (
	"context"
	"github.com/google/uuid"

	"github.com/yeOmaNnn/orchestrator/internal/domain"
	"github.com/yeOmaNnn/orchestrator/internal/planner"
	"github.com/yeOmaNnn/orchestrator/internal/scheduler"
	"github.com/yeOmaNnn/orchestrator/internal/storage"
)

type Engine struct {
	planner   planner.Client
	scheduler *scheduler.Scheduler
	taskRepo  storage.TaskRepository
	stepRepo  storage.StepRepository
}

func New(
	planner planner.Client,
	scheduler *scheduler.Scheduler,
	taskRepo storage.TaskRepository,
	stepRepo storage.StepRepository,
) *Engine {
	return &Engine{
		planner:   planner,
		scheduler: scheduler,
		taskRepo:  taskRepo,
		stepRepo:  stepRepo,
	}
}

func (e *Engine) InitTaskExecution(
	ctx  context.Context,
	task domain.Task,
) error {

	plan, err := e.planner.Plan(ctx, planner.PlanRequest{
		TaskID: task.ID,
		Goal:   task.Goal,
	})
	if err != nil {
		return err
	}

	steps := planner.MapToDomainSteps(task.ID, plan.Steps)

	return e.stepRepo.CreateMany(ctx, steps)
}

func (e *Engine) RunTask(
	ctx context.Context,
	taskID uuid.UUID,
) error {
	return e.RunTaskLoop(ctx, taskID)
}


func (e *Engine) CancelTask(
	ctx context.Context, 
	taskID uuid.UUID, 
) error {
	if err := e.stepRepo.CancelByTask(ctx, taskID); err != nil {
		return err 
	}

	if err := e.taskRepo.UpdateStatus(
		ctx, 
		taskID, 
		domain.TaskCanceled,
	); err != nil {
		return err 
	}

	return nil 
}