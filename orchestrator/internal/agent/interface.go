package agent

import (
	"context"
	"encoding/json"
	"errors"
	"time"
)

var (
	ErrNotCallable = errors.New("agent not call")
	ErrCircuitOpen = errors.New("breaker not found")
	ErrAgentNotFound = errors.New("agent is not health")
)

type Client interface {
	Name() string 
}
type Callable interface {
	Call(
		ctx context.Context, 
		input json.RawMessage, 
	) (json.RawMessage, error)
}

type HealthCheckable interface {
	HealthChecker(ctx context.Context) error 
}

type Agent interface {
	Client 
	Callable
	HealthCheckable
}

type Config struct {
	Name string 
	BaseURL string 
	Timeout time.Duration 
	MaxRetries int 
	CircuitBreaker bool 
	HealthCheck bool 
	HealthCheckURL string 
}
