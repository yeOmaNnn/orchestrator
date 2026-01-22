package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/yeOmaNnn/orchestrator/internal/agent"
	"github.com/yeOmaNnn/orchestrator/internal/api"
	"github.com/yeOmaNnn/orchestrator/internal/domain"
	"github.com/yeOmaNnn/orchestrator/internal/engine"
	"github.com/yeOmaNnn/orchestrator/internal/planner"
	"github.com/yeOmaNnn/orchestrator/internal/runner"
	"github.com/yeOmaNnn/orchestrator/internal/scheduler"
	// "github.com/yeOmaNnn/orchestrator/internal/storage"
)

type InMemoryTaskRepo struct {
	tasks map[uuid.UUID]*domain.Task
}

func NewInMemoryTaskRepo() *InMemoryTaskRepo {
	return &InMemoryTaskRepo{
		tasks: make(map[uuid.UUID]*domain.Task),
	}
}

func (r *InMemoryTaskRepo) Create(ctx context.Context, task *domain.Task) error {
	r.tasks[task.ID] = task
	return nil
}

func (r *InMemoryTaskRepo) GetByID(ctx context.Context, id uuid.UUID) (*domain.Task, error) {
	t, ok := r.tasks[id]
	if !ok {
		return nil, fmt.Errorf("task not found")
	}
	return t, nil
}

func (r *InMemoryTaskRepo) UpdateStatus(
	ctx context.Context,
	id uuid.UUID,
	status domain.TaskStatus,
) error {
	t, ok := r.tasks[id]
	if !ok {
		return fmt.Errorf("task not found")
	}
	t.Status = status
	return nil
}


type InMemoryStepRepo struct {
	steps map[uuid.UUID]domain.Step
}

func NewInMemoryStepRepo() *InMemoryStepRepo {
	return &InMemoryStepRepo{
		steps: make(map[uuid.UUID]domain.Step),
	}
}

func (r *InMemoryStepRepo) CreateMany(
	ctx context.Context,
	steps []domain.Step,
) error {
	for _, s := range steps {
		r.steps[s.ID] = s
	}
	return nil
}

func (r *InMemoryStepRepo) CancelByTask(
	ctx context.Context, 
	taskID uuid.UUID,
) error {
	for id, step := range r.steps {
		if step.TaskID != taskID {
			continue
		}

		if step.Status == domain.StepDone || 
		step.Status == domain.StepError || 
		step.Status == domain.StepCancelled {
			continue
		}

		step.Status = domain.StepCancelled
		step.UpdatedAt = time.Now()
		r.steps[id] = step 
	}
	return nil 
}

func (r *InMemoryStepRepo) GetByTask(
	ctx context.Context,
	taskID uuid.UUID,
) ([]domain.Step, error) {

	var out []domain.Step
	for _, s := range r.steps {
		if s.TaskID == taskID {
			out = append(out, s)
		}
	}
	return out, nil
}

func (r *InMemoryStepRepo) Update(
	ctx context.Context,
	step *domain.Step,
) error {
		if step.Output == nil {
			step.Output = json.RawMessage(`{}`)

		}
		r.steps[step.ID] = *step
		return nil
	}

func (r *InMemoryStepRepo) AcquireReadySteps(
    ctx context.Context,
    taskID uuid.UUID,
    limit int,
) ([]domain.Step, error) {
    
    var readySteps []domain.Step
    stepsByID := make(map[uuid.UUID]domain.Step)
    
    for _, s := range r.steps {
        if s.TaskID == taskID {
            stepsByID[s.ID] = s
        }
    }
    
    for _, s := range stepsByID {
        if s.Status != domain.StepWaiting {
            continue
        }
        
        ready := true
        for _, depID := range s.DependsOn {
            if dep, exists := stepsByID[depID]; exists && dep.Status != domain.StepDone {
                ready = false
                break
            }
        }
        
        if ready {
            s.Status = domain.StepInProgress
            r.steps[s.ID] = s
            readySteps = append(readySteps, s)
            
            if len(readySteps) >= limit {
                break
            }
        }
    }
    
    return readySteps, nil
}
 

type DummyPlanner struct{}

func (p *DummyPlanner) Plan(
	ctx context.Context,
	req planner.PlanRequest,
) (planner.PlanResponse, error) {

	inputBytes, err := json.Marshal(map[string]any{
		"text": "hello",
	})
	if err != nil {
		return planner.PlanResponse{}, err
	}

	step := planner.PlannedStep{
		ID:        uuid.New(),
		Agent:     "agent1",
		Input:     inputBytes,
		DependsOn: []uuid.UUID{},
	}

	return planner.PlanResponse{
		Steps: []planner.PlannedStep{step},
	}, nil
}


type DummyAgent struct{}

func (a *DummyAgent) Name() string {
	return "agent1"
}

func (a *DummyAgent) Call(
	ctx context.Context,
	input json.RawMessage,
) (json.RawMessage, error) {

	return json.RawMessage(`{"result":"ok"}`), nil
}

func (r *InMemoryTaskRepo) ListActive (
	ctx context.Context, 
) ([]domain.Task, error) {
	var out []domain.Task
	for _, t := range r.tasks {
		if t.Status == domain.TaskPending || 
		t.Status == domain.TaskRunning {
			out = append(out, *t)
		}
	}

	return out, nil
}  

func main() {
	ctx := context.Background()

	taskRepo := NewInMemoryTaskRepo()
	stepRepo := NewInMemoryStepRepo()

	registry := agent.NewRegistry()
	registry.Register(&DummyAgent{})

	router := agent.NewRouter(registry)

	runnerService := runner.New(
		stepRepo,
		router,
	)

	schedulerService := scheduler.New(
		stepRepo,
		taskRepo,
		runnerService,
		2,

	)
	go schedulerService.Run(ctx)

	plannerClient := &DummyPlanner{}

	eng := engine.New(
		plannerClient,
		schedulerService,
		taskRepo,
		stepRepo,		
	)

	handler := api.NewHandler(
		eng, 
		taskRepo, 
		stepRepo,
	)

	mux := http.NewServeMux()
	handler.Register(mux)

	

	task := domain.Task{
		ID:     uuid.New(),
		Goal:   "Say Hello",
		Status: domain.TaskPending,
	}

	_ = taskRepo.Create(ctx, &task)
	

	finalTask, _ := taskRepo.GetByID(ctx, task.ID)
	fmt.Println("Task finished with status:", finalTask.Status)

	if err := taskRepo.Create(ctx, &task); err != nil {
		log.Fatal(err)
	}

	if err := eng.InitTaskExecution(ctx, task); err != nil {
		log.Fatal(err)
	}

	steps, _ := stepRepo.GetByTask(ctx, task.ID)
	for _, s := range steps {
		fmt.Println(
			"Step:",
			s.ID,
			"Status:",
			s.Status,
			"Output:",
			string(s.Output),
		)
	}
	
	log.Println("api listening :8080")
	log.Fatal(http.ListenAndServe(":8080", mux))
	
}

