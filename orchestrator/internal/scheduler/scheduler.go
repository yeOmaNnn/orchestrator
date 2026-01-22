package scheduler

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/google/uuid"

	"github.com/yeOmaNnn/orchestrator/internal/domain"
	"github.com/yeOmaNnn/orchestrator/internal/runner"
	"github.com/yeOmaNnn/orchestrator/internal/storage"
)

type Scheduler struct {
	workerID    string
	stepsRepo 	storage.StepRepository
	taskRepo  	storage.TaskRepository
	runner    	*runner.Runner
	 
	maxParallel int

	queue 		chan domain.Step

	ticker      *time.Ticker
	stop        chan struct{}
	wg          sync.WaitGroup
}

func New(
	stepsRepo   storage.StepRepository,
	taskRepo    storage.TaskRepository,
	runner      *runner.Runner,
	maxParallel int,
) *Scheduler {
	return &Scheduler{
		workerID: uuid.NewString(),
		stepsRepo:   stepsRepo,
		taskRepo:    taskRepo,
		runner:      runner,
		maxParallel: maxParallel,
		queue: make(chan domain.Step, maxParallel*2),
		stop: make(chan struct{}),
	}
}


func (s *Scheduler) Schedule(
	ctx context.Context,
	taskID uuid.UUID,
) error {

	steps, err := s.stepsRepo.GetByTask(ctx, taskID)
	if err != nil {
		return err
	}

	stepByID := make(map[uuid.UUID]domain.Step, len(steps))
	for _, step := range steps {
		stepByID[step.ID] = step
	}

	var readySteps []domain.Step
	now := time.Now()

	for _, step := range steps {
		if step.Status != domain.StepWaiting {
			continue
		}

		if step.NextRunAt != nil && step.NextRunAt.After(now) {
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
		readySteps[i].UpdatedAt = time.Now()

		if err := s.stepsRepo.Update(ctx, &readySteps[i]); err != nil {
			return err
		}
	}

	sem := make(chan struct{}, s.maxParallel)
	var wg sync.WaitGroup

	for _, step := range readySteps {
		wg.Add(1)

		select {
		case sem <- struct{}{}:
		case <-ctx.Done():
			wg.Done()
			return ctx.Err()
		}

		go func(st domain.Step) {
			defer wg.Done()
			defer func() { <-sem }()

			stepCtx := ctx
			if st.TimeoutSeconds > 0 {
				var cancel context.CancelFunc
				stepCtx, cancel = context.WithTimeout(
					ctx,
					time.Duration(st.TimeoutSeconds)*time.Second,
				)
				defer cancel()
			}

			err := s.runner.Run(stepCtx, st)

			if err != nil {
				now := time.Now()
				st.Attempt++

				if errors.Is(err, context.DeadlineExceeded) {
					st.LastError = "timeout exceeded"
				} else {
					st.LastError = err.Error()
				}

				if st.Attempt >= st.MaxAttempts {
					st.Status = domain.StepError
				} else {
					delay := nextBackoff(st.Attempt)
					next := now.Add(delay)

					st.Status = domain.StepWaiting
					st.NextRunAt = &next
				}

				st.UpdatedAt = time.Now()
				_ = s.stepsRepo.Update(ctx, &st)
				return
			}

			st.Status = domain.StepDone
			st.UpdatedAt = time.Now()
			_ = s.stepsRepo.Update(ctx, &st)

		}(step)
	}

	wg.Wait()
	return nil
}

func dependenciesDone(
	step domain.Step,
	all map[uuid.UUID]domain.Step,
) bool {
	for _, depID := range step.DependsOn {
		dep, ok := all[depID]
		if !ok || dep.Status != domain.StepDone {
			return false
		}
	}
	return true
}

func nextBackoff(attempt int) time.Duration {
	switch {
	case attempt <= 1:
		return 1 * time.Second
	case attempt == 2:
		return 3 * time.Second
	default:
		return time.Duration(attempt*attempt) * time.Second
	}
}

func (s *Scheduler) worker (
	ctx context.Context, 
	id int, 
) {
	for {
		select {
		case <-ctx.Done():
			return 
		case step := <-s.queue:
			s.executeStep(ctx, step)
		}
	}
}

func (s *Scheduler) executeStep(
	ctx context.Context,
	st domain.Step,
) {
	stepCtx := ctx

	if st.TimeoutSeconds > 0 {
		var cancel context.CancelFunc
		stepCtx, cancel = context.WithTimeout(
			ctx,
			time.Duration(st.TimeoutSeconds)*time.Second,
		)
		defer cancel()
	}

	err := s.runner.Run(stepCtx, st)

	now := time.Now()

	if err != nil {
		st.Attempt++
		st.LastError = err.Error()

		if st.Attempt >= st.MaxAttempts {
			st.Status = domain.StepError
			st.FinishedAt = &now
		} else {
			delay := nextBackoff(st.Attempt)
			next := now.Add(delay)

			st.Status = domain.StepWaiting
			st.NextRunAt = &next
		}

		st.UpdatedAt = now
		_ = s.stepsRepo.Update(ctx, &st)
		return
	}

	st.Status = domain.StepDone
	st.FinishedAt = &now
	st.UpdatedAt = now
	_ = s.stepsRepo.Update(ctx, &st)
}


func (s *Scheduler) Run(ctx context.Context) {
	for i := 0; i < s.maxParallel; i++ {
		go s.worker(ctx, i)
	}
	ticker := time.NewTicker(500 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			s.tick(ctx)
		}
	}
}

func (s *Scheduler) RunOnce(
	ctx context.Context,
	taskID uuid.UUID,
) error {

	steps, err := s.stepsRepo.AcquireReadySteps(
		ctx,
		taskID,
		s.maxParallel,
		s.workerID,
	)
	if err != nil {
		return err
	}

	for _, step := range steps {
		go func(st domain.Step) {
			_ = s.runner.Run(ctx, st)
		}(step)
	}

	return nil
}


func (s *Scheduler) tick(ctx context.Context) {

	tasks, err := s.taskRepo.ListActive(ctx)
	if err != nil {
		return
	}

	for _, task := range tasks {
		steps, err := s.stepsRepo.AcquireReadySteps(
			ctx,
			task.ID,
			s.maxParallel,
			s.workerID,
		)
		if err != nil {
			continue
		}

		for _, step := range steps {
			select {
			case s.queue <- step:
			default:
			}
		}
	}
}

