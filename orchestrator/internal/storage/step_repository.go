package storage

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/yeOmaNnn/orchestrator/internal/domain"
)

type StepRepository interface {
	CreateMany(
		ctx context.Context,
		steps []domain.Step,
	) error

	GetByTask(
		ctx context.Context,
		taskID uuid.UUID,
	) ([]domain.Step, error)

	Update(
		ctx context.Context,
		step *domain.Step,
	) error 

	AcquireReadySteps(
		ctx context.Context, 
		taskID uuid.UUID, 
		limit int,
		workerID string, 
	) ([]domain.Step, error) 

	CancelByTask(
		ctx context.Context,
		 taskID uuid.UUID,
	) error 

	ReleaseStaleLocks(
		ctx context.Context, 
		ttl time.Duration,
	) error 
}

