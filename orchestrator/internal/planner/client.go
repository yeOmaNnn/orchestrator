package planner

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
)

type Client interface {
	Plan(ctx context.Context, req PlanRequest) (PlanResponse, error) 
}

type HTTPClient struct {
	baseURL string
	client  *http.Client
}

func NewClient(baseURL string) *HTTPClient {
	return &HTTPClient{
		baseURL: baseURL,
		client:  &http.Client{},
	}
}

func (c *HTTPClient) Plan(
	ctx context.Context,
	req PlanRequest,
) (PlanResponse, error) {

	body, err := json.Marshal(req)
	if err != nil {
		return PlanResponse{}, err
	}

	httpReq, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		c.baseURL+"/plan",
		bytes.NewReader(body),
	)
	if err != nil {
		return PlanResponse{}, err
	}

	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return PlanResponse{}, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return PlanResponse{}, errors.New("planner returned non-200")
	}

	var out PlanResponse
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		return PlanResponse{}, err
	}

	return out, nil
}
