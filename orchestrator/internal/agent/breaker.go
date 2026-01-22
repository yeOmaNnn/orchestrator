package agent

import (
	//"errors"
	"sync"
	"time"
)

type CircuitBreaker struct {
	mu                sync.RWMutex
	state             CircuitState
	failureCount      int
	successCount      int
	lastFailure       time.Time
	lastSuccess       time.Time
	halfOpenTimeout   time.Duration
	failureThreshold  int
	successThreshold  int
	consecutiveErrors int
	resetTimeout      time.Duration
}

func NewCircuitBreaker(config *CircuitBreakerConfig) *CircuitBreaker {
	if config == nil {
		config = &CircuitBreakerConfig{
			FailureThreshold: 5,
			SuccessThreshold: 3,
			HalfOpenTimeout:  10 * time.Second,
			ResetTimeout:     60 * time.Second,
		}
	}

	return &CircuitBreaker{
		state:             StateClosed,
		failureThreshold:  config.FailureThreshold,
		successThreshold:  config.SuccessThreshold,
		halfOpenTimeout:   config.HalfOpenTimeout,
		resetTimeout:      config.ResetTimeout,
	}
}

type CircuitState int

const (
	StateClosed CircuitState = iota
	StateHalfOpen
	StateOpen
)

type CircuitBreakerConfig struct {
	FailureThreshold int           `json:"failure_threshold"`
	SuccessThreshold int           `json:"success_threshold"`
	HalfOpenTimeout  time.Duration `json:"half_open_timeout"`
	ResetTimeout     time.Duration `json:"reset_timeout"`
}

func (cb *CircuitBreaker) Execute(fn func() error) error {
	cb.mu.Lock()

	if cb.state == StateOpen && time.Since(cb.lastFailure) > cb.resetTimeout {
		cb.state = StateHalfOpen
		cb.successCount = 0
	}

	if cb.state == StateOpen {
		cb.mu.Unlock()
		return ErrCircuitOpen
	}

	cb.mu.Unlock()

	err := fn()

	cb.mu.Lock()
	defer cb.mu.Unlock()

	if err != nil {
		cb.handleFailure()
		return err
	}

	cb.handleSuccess()
	return nil
}

func (cb *CircuitBreaker) handleFailure() {
	cb.failureCount++
	cb.consecutiveErrors++
	cb.lastFailure = time.Now()
	cb.successCount = 0

	switch cb.state {
	case StateClosed:
		if cb.consecutiveErrors >= cb.failureThreshold {
			cb.state = StateOpen
			cb.consecutiveErrors = 0
		}
	case StateHalfOpen:
		cb.state = StateOpen
		cb.consecutiveErrors = 0
	}
}

func (cb *CircuitBreaker) handleSuccess() {
	cb.successCount++
	cb.lastSuccess = time.Now()
	cb.consecutiveErrors = 0

	if cb.state == StateHalfOpen && cb.successCount >= cb.successThreshold {
		cb.state = StateClosed
		cb.failureCount = 0
	}
}

func (cb *CircuitBreaker) State() CircuitState {
	cb.mu.RLock()
	defer cb.mu.RUnlock()
	return cb.state
}

func (cb *CircuitBreaker) Metrics() map[string]interface{} {
	cb.mu.RLock()
	defer cb.mu.RUnlock()

	return map[string]interface{}{
		"state":              cb.state,
		"failure_count":      cb.failureCount,
		"success_count":      cb.successCount,
		"consecutive_errors": cb.consecutiveErrors,
		"last_failure":       cb.lastFailure,
		"last_success":       cb.lastSuccess,
		"is_open":            cb.state == StateOpen,
	}
}