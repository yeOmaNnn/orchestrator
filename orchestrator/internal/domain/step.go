package domain

import (
	"encoding/json"
	"github.com/google/uuid"
)

type StepStatus string

const (
	StepWaiting    StepStatus = "WAITING"
	StepInProgress StepStatus = "IN_PROGRESS"
	StepDone       StepStatus = "DONE"
	StepError      StepStatus = "ERROR"
)

type Step struct {
	ID         uuid.UUID
	TaskID     uuid.UUID
	Agent      string
	Input      json.RawMessage
	Output     json.RawMessage
	Status     StepStatus
	RetryCount int
	DependsOn  []uuid.UUID
}
