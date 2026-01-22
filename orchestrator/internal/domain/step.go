package domain

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

type StepStatus string

const (
	StepWaiting    StepStatus = "WAITING"
	StepInProgress StepStatus = "IN_PROGRESS"
	StepDone       StepStatus = "DONE"
	StepFailed     StepStatus = "FAILED"  
	StepError      StepStatus = "ERROR"
	StepCancelled  StepStatus = "CANCELLED"
)

type Step struct {
	ID             uuid.UUID      `json:"id"`
	TaskID         uuid.UUID      `json:"task_id"`
	Agent          string         `json:"agent"`
	Input          json.RawMessage `json:"input"`
	Output         json.RawMessage `json:"output"`
	Status         StepStatus     `json:"status"`
	RetryCount     int            `json:"retry_count"`
	MaxRetries     int            `json:"max_retries"`
	DependsOn      []uuid.UUID    `json:"depends_on"`
	Attempt        int            `json:"attempt"`
	MaxAttempts    int            `json:"max_attempts"`
	NextRunAt      *time.Time     `json:"next_run_at"`
	LastError      string         `json:"last_error"`
	TimeoutSeconds int            `json:"timeout_seconds"`
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	LockedAt       *time.Time     `json:"locked_at"`    
	LockedBy       *string        `json:"locked_by"`    
	StartedAt      *time.Time     `json:"started_at"`   
	FinishedAt     *time.Time     `json:"finished_at"`  
}

func NewStep(taskID uuid.UUID, agent string, input json.RawMessage) *Step {
	now := time.Now()
	return &Step{
		ID:             uuid.New(),
		TaskID:         taskID,
		Agent:          agent,
		Input:          input,
		Status:         StepWaiting,
		RetryCount:     0,
		MaxRetries:     3,                     
		Attempt:        0,
		MaxAttempts:    3,                     
		TimeoutSeconds: 30,                    
		CreatedAt:      now,
		UpdatedAt:      now,
		DependsOn:      []uuid.UUID{},
	}
}

func (s *Step) CanRetry() bool {
	if s.Status != StepFailed && s.Status != StepError {
		return false
	}
	
	if s.RetryCount >= s.MaxRetries {
		return false
	}
	
	if s.Attempt >= s.MaxAttempts {
		return false
	}
	
	if s.NextRunAt != nil && time.Now().Before(*s.NextRunAt) {
		return false
	}
	
	return true
}

func (s *Step) MarkForRetry(delay time.Duration) {
	now := time.Now()
	nextRun := now.Add(delay)
	
	s.Status = StepWaiting
	s.RetryCount++
	s.Attempt++
	s.NextRunAt = &nextRun
	s.LockedAt = nil
	s.LockedBy = nil
	s.UpdatedAt = now
}

func (s *Step) MarkInProgress(workerID string) {
	now := time.Now()
	s.Status = StepInProgress
	s.StartedAt = &now
	s.LockedAt = &now
	workerIDCopy := workerID
	s.LockedBy = &workerIDCopy
	s.UpdatedAt = now
}

func (s *Step) MarkDone(output json.RawMessage) {
	now := time.Now()
	s.Status = StepDone
	s.Output = output
	s.FinishedAt = &now
	s.LockedAt = nil
	s.LockedBy = nil
	s.UpdatedAt = now
}

func (s *Step) MarkFailed(err error) {
	now := time.Now()
	s.Status = StepFailed
	s.LastError = err.Error()
	s.FinishedAt = &now
	s.LockedAt = nil
	s.LockedBy = nil
	s.UpdatedAt = now
}