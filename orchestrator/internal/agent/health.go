package agent

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type HealthChecker struct {
	registry  *Registry
	statuses  map[string]*HealthStatus
	mu        sync.RWMutex
	interval  time.Duration
	stopCh    chan struct{}
}

type HealthStatus struct {
	Name      string        `json:"name"`
	Healthy   bool          `json:"healthy"`
	LastCheck time.Time     `json:"last_check"`
	Latency   time.Duration `json:"latency"`
	Error     string        `json:"error,omitempty"`
}

type healthCheckable interface {
	HealthCheck(ctx context.Context) error
}

func NewHealthChecker(registry *Registry, interval time.Duration) *HealthChecker {
	return &HealthChecker{
		registry: registry,
		statuses: make(map[string]*HealthStatus),
		interval: interval,
		stopCh:   make(chan struct{}),
	}
}

func (hc *HealthChecker) Start() {
	go hc.run()
}

func (hc *HealthChecker) run() {
	ticker := time.NewTicker(hc.interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			hc.checkAll(context.Background())
		case <-hc.stopCh:
			return
		}
	}
}

func (hc *HealthChecker) Stop() {
	close(hc.stopCh)
}

func (hc *HealthChecker) checkAll(ctx context.Context) {
	agents := hc.registry.List()

	var wg sync.WaitGroup
	for _, agentName := range agents {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			hc.checkAgent(ctx, name)
		}(agentName)
	}
	wg.Wait()
}

func (hc *HealthChecker) checkAgent(ctx context.Context, name string) {
	client, err := hc.registry.Get(name)
	if err != nil {
		hc.updateStatus(name, false, 0, err.Error())
		return
	}

	healthCheckable, ok := client.(healthCheckable)
	if !ok {
		hc.updateStatus(name, true, 0, "")
		return
	}

	start := time.Now()
	err = healthCheckable.HealthCheck(ctx)
	latency := time.Since(start)

	if err != nil {
		hc.updateStatus(name, false, latency, err.Error())
	} else {
		hc.updateStatus(name, true, latency, "")
	}
}

func (hc *HealthChecker) updateStatus(name string, healthy bool, latency time.Duration, error string) {
	hc.mu.Lock()
	defer hc.mu.Unlock()

	hc.statuses[name] = &HealthStatus{
		Name:      name,
		Healthy:   healthy,
		LastCheck: time.Now(),
		Latency:   latency,
		Error:     error,
	}
}

func (hc *HealthChecker) GetStatus(name string) (*HealthStatus, bool) {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	status, exists := hc.statuses[name]
	return status, exists
}

func (hc *HealthChecker) GetAllStatuses() map[string]*HealthStatus {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	result := make(map[string]*HealthStatus)
	for k, v := range hc.statuses {
		result[k] = v
	}
	return result
}

func (hc *HealthChecker) GetAllStatusesErrors() map[string]error {
	hc.mu.RLock()
	defer hc.mu.RUnlock()

	result := make(map[string]error)
	for k, v := range hc.statuses {
		if !v.Healthy {
			result[k] = fmt.Errorf(v.Error)
		}
	}
	return result
}