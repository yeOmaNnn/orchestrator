package planner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/envoyproxy/protoc-gen-validate/validate"
)

type PlanRequest struct {
	Goal string `json:"goal"`
}

type PlanResponse struct {
	Steps []PlannedStep `json:"steps"`
}

type PlannedStep struct {
	ID        string   `json:"id"`
	Agent     string   `json:"agent"`
	Input     any      `json:"input"`
	DependsOn []string `json:"depends_on"`
}

type Client interface {
	CreatePlan(ctx context.Context, goal string) (*PlanResponse, error)
}

type HTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewHTTPClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 10 * time.Second,
		},
	}
}

func (c *HTTPClient) CreatePlan(ctx context.Context, goal string) (*PlanResponse, error) {
	reqBody, err := json.Marshal(PlanRequest{
		Goal: goal,
	})
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/plan",
		bytes.NewBuffer(reqBody),
	)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, errors.New("200")
	}

	var plan PlanResponse
	if err := json.NewDecoder(resp.Body).Decode(&plan); err != nil {
		return nil, err
	}

	if err := validatePlan(plan); err != nil {
		return nil, err
	}

	return &plan, nil
}

func validatePlan(plan PlanResponse) error {
	if len(plan.Steps) == 0 {
		return errors.New("empty plan")
	}

	ids := make(map[string]struct{})

	for _, step := range plan.Steps {
		if step.ID == "" {
			return errors.New("no step-id")
		}

		if _, exists := ids[step.ID]; exists {
			return errors.New("duplicate ids: " + step.ID)
		}

		ids[step.ID] = struct{}{}

		for _, dep := range step.DependsOn {
			if dep == step.ID {
				return errors.New("step depends on itself")
			}
		}
	}

	for _, step := range plan.Steps {
		for _, dep := range step.DependsOn {
			if _, ok := ids[dep]; !ok {
				return errors.New("unknow deps: " + dep)
			}
		}
	}

	return nil
}
