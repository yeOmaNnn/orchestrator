package storage

import (
	"context"
	"github.com/google/uuid"
	"github.com/yeOmaNnn/orchestrator/internal/domain"
)

type TaskRepository interface {
	Create(
		ctx context.Context,
		task *domain.Task,
		) error

	GetByID(
		ctx context.Context,
		id uuid.UUID,
		) (*domain.Task, error)

	UpdateStatus(
		ctx context.Context,
		id uuid.UUID,
		status domain.TaskStatus,
		) error

	ListActive(
		ctx context.Context) ([]domain.Task, error)
}
