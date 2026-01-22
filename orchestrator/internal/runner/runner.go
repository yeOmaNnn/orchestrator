package runner

import (
	//"bytes"
	"context"
	"encoding/json"
	//"errors"
	//"net/http"
	"time"

	"github.com/yeOmaNnn/orchestrator/internal/domain"
	"github.com/yeOmaNnn/orchestrator/internal/storage"
)

type AgentClient interface {
	Call(
		ctx   context.Context, 
		agent string, 
		input json.RawMessage, 
	) (json.RawMessage, error) 
}

type Runner struct {
	stepsRepo  storage.StepRepository
	client     AgentClient

	maxRetries int 
	timeout    time.Duration
}

func New(
	stepsRepo storage.StepRepository, 
	client    AgentClient, 
	opts      ...Option, 
) *Runner {
	r := &Runner{
		stepsRepo:  stepsRepo, 
		client:     client, 
		maxRetries: 3,
		timeout:    30 * time.Second,
	}

	for _, opt := range opts {
		opt(r)
	}

	return r 
}

type Option func(*Runner) 

func WithMaxRetries(n int) Option {
	return func(r *Runner) {
		r.maxRetries = n
	}
}

func WithTimeout(d time.Duration) Option {
	return func(r *Runner) {
		r.timeout = d
	}
}

func (r *Runner) Run(
	ctx context.Context,
	step domain.Step,
) error {
		ctx, cancel := context.WithTimeout(ctx, r.timeout)
		defer cancel()
		output, err := r.client.Call(ctx, step.Agent, step.Input)
		if err != nil {
			step.RetryCount++

			if step.RetryCount >= r.maxRetries {
				step.Status = domain.StepError
			} else {
				step.Status = domain.StepWaiting
			}

			_ = r.stepsRepo.Update(ctx, &step)
			return err
		}
	step.Output = output
	step.Status = domain.StepDone

	return r.stepsRepo.Update(ctx, &step)
}


// type HTTPAgentClient struct {
// 	baseURL string
// 	client  *http.Client
// }

// func NewHTTPAgentClient(baseURL string) *HTTPAgentClient {
// 	return &HTTPAgentClient{
// 		baseURL: baseURL,
// 		client:  &http.Client{},
// 	}
// }

// func (c *HTTPAgentClient) Call(
// 	ctx context.Context,
// 	agent string,
// 	input json.RawMessage,
// ) (json.RawMessage, error) {

// 	body, _ := json.Marshal(map[string]any{
// 		"input": input,
// 	})

// 	req, err := http.NewRequestWithContext(
// 		ctx,
// 		http.MethodPost,
// 		c.baseURL+"/agents/"+agent,
// 		bytes.NewReader(body),
// 	)
// 	if err != nil {
// 		return nil, err
// 	}

// 	req.Header.Set("Content-Type", "application/json")

// 	resp, err := c.client.Do(req)
// 	if err != nil {
// 		return nil, err
// 	}
// 	defer resp.Body.Close()

// 	if resp.StatusCode != http.StatusOK {
// 		return nil, errors.New("agent returned non-200")
// 	}

// 	var result struct {
// 		Output json.RawMessage `json:"output"`
// 	}

// 	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
// 		return nil, err
// 	}

// 	return result.Output, nil
// }
