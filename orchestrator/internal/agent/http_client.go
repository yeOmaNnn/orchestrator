package agent

import (
	"bytes"
	"context"
	"encoding/json"
	//"errors"
	"fmt"
	"net/http"
	"time"
)

type HTTPAgentClient struct {
	name         string
	baseURL      string
	client       *http.Client
	circuit      *CircuitBreaker
	healthURL    string
	timeout      time.Duration
	lastHealthOK time.Time
}

func NewHTTPAgent(name string, baseURL string, opts ...HTTPAgentOption) *HTTPAgentClient {
	h := &HTTPAgentClient{
		name:    name,
		baseURL: baseURL,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		healthURL: baseURL + "/health",
		timeout:   30 * time.Second,
	}

	for _, opt := range opts {
		opt(h)
	}

	if h.circuit == nil {
		h.circuit = NewCircuitBreaker(&CircuitBreakerConfig{
			FailureThreshold: 5,
			SuccessThreshold: 3,
			HalfOpenTimeout:  10 * time.Second,
			ResetTimeout:     60 * time.Second,
		})
	}

	return h
}

type HTTPAgentOption func(*HTTPAgentClient)

func WithTimeout(timeout time.Duration) HTTPAgentOption {
	return func(h *HTTPAgentClient) {
		h.timeout = timeout
		h.client.Timeout = timeout
	}
}

func WithCircuitBreaker(cb *CircuitBreaker) HTTPAgentOption {
	return func(h *HTTPAgentClient) {
		h.circuit = cb
	}
}

func WithHealthCheckURL(url string) HTTPAgentOption {
	return func(h *HTTPAgentClient) {
		h.healthURL = url
	}
}

func (c *HTTPAgentClient) Name() string {
	return c.name
}

func (c *HTTPAgentClient) Call(ctx context.Context, input json.RawMessage) (json.RawMessage, error) {
	var result json.RawMessage
	//var callErr error

	err := c.circuit.Execute(func() error {
		callCtx, cancel := context.WithTimeout(ctx, c.timeout)
		defer cancel()

		body, err := json.Marshal(map[string]any{
			"input": input,
			"metadata": map[string]any{
				"agent": c.name,
				"time":  time.Now().UTC(),
			},
		})
		if err != nil {
			return fmt.Errorf("marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(
			callCtx,
			http.MethodPost,
			c.baseURL+"/run",
			bytes.NewReader(body),
		)
		if err != nil {
			return fmt.Errorf("create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Agent-Name", c.name)

		resp, err := c.client.Do(req)
		if err != nil {
			return fmt.Errorf("execute request: %w", err)
		}
		defer resp.Body.Close()

		if resp.StatusCode != http.StatusOK {
			return fmt.Errorf("agent returned status %d", resp.StatusCode)
		}

		var out struct {
			Output  json.RawMessage `json:"output"`
			Error   string          `json:"error,omitempty"`
			Retry   bool            `json:"should_retry,omitempty"`
		}

		if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
			return fmt.Errorf("decode response: %w", err)
		}

		if out.Error != "" {
			return fmt.Errorf("agent error: %s", out.Error)
		}

		result = out.Output
		return nil
	})

	if err != nil {
		return nil, err
	}

	return result, nil
}

func (c *HTTPAgentClient) HealthCheck(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		c.healthURL,
		nil,
	)
	if err != nil {
		return fmt.Errorf("create health check request: %w", err)
	}

	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("health check failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("health check returned status %d", resp.StatusCode)
	}

	c.lastHealthOK = time.Now()
	return nil
}

func (c *HTTPAgentClient) IsHealthy() bool {
	if c.lastHealthOK.IsZero() {
		return false
	}
	return time.Since(c.lastHealthOK) < 5*time.Minute
}