package domain

import (
	"github.com/google/uuid"
	"time"
)

type TaskStatus string

const (
	TaskPending   TaskStatus = "PENDING"
	TaskRunning   TaskStatus = "RUNNING"
	TaskCompleted TaskStatus = "COMPLETED"
	TaskFailed    TaskStatus = "FAILED"
	TaskCanceled  TaskStatus = "CANCELED"
)

type Task struct {
	ID        uuid.UUID
	Goal      string
	Status    TaskStatus
	CreatedAt time.Time
}
