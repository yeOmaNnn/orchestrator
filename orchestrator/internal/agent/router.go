package agent

import (
	"context"
	"encoding/json"
	"fmt"
)

type Router struct {
	registry *Registry
}

func NewRouter(registry *Registry) *Router {
	return &Router{
		registry: registry,
	}
}

func (r *Router) Call(
	ctx context.Context,
	agent string,
	input json.RawMessage,
) (json.RawMessage, error) {
	client, err := r.registry.Get(agent)
	if err != nil {
		return nil, err
	}

	callable, ok := client.(Callable)
	if !ok {
		return nil, ErrNotCallable
	}

	return callable.Call(ctx, input)
}

func (r *Router) CallWithHealthCheck(
	ctx context.Context,
	agent string,
	input json.RawMessage,
) (json.RawMessage, error) {
	client, err := r.registry.Get(agent)
	if err != nil {
		return nil, err
	}

	if healthCheckable, ok := client.(interface{ HealthCheck(context.Context) error }); ok {
		if err := healthCheckable.HealthCheck(ctx); err != nil {
			return nil, fmt.Errorf("agent %s health check failed: %w", agent, err)
		}
	}

	callable, ok := client.(Callable)
	if !ok {
		return nil, ErrNotCallable
	}

	return callable.Call(ctx, input)
}