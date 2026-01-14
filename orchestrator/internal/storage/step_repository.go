package storage

import (
	"context"

	"github.com/google/uuid"
	"orchestrator/internal/domain"
)

type StepRepository interface {
	CreateMany(ctx context.Context, steps []domain.Step) error
	GetByTask(ctx context.Context, taskID uuid.UUID) ([]domain.Step, error)
	Update(ctx context.Context, step *domain.Step) error
}
