package planner

import (
	"encoding/json"
	"github.com/google/uuid"
)

type PlanRequest struct {
	TaskID uuid.UUID `json:"task_id"`
	Goal   string    `json:"goal"`
}

type PlannedStep struct {
	ID uuid.UUID `json:"id"`
	Agent string `json:"agent"`
	Input json.RawMessage `json:"input"`
	DependsOn []uuid.UUID `json:"depends_on"`
}

type PlanResponse struct {
	Steps []PlannedStep `json:"steps"`
}

