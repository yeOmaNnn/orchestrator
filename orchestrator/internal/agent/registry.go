package agent

import (
	"fmt"
	"sync"
)

type Registry struct {
	mu      sync.RWMutex
	clients map[string]Client
}

func NewRegistry() *Registry {
	return &Registry{
		clients: make(map[string]Client),
	}
}

func (r *Registry) Register(client Client) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clients[client.Name()] = client
}

func (r *Registry) Get(name string) (Client, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.clients[name]
	if !ok {
		return nil, fmt.Errorf("agent %s not registered", name)
	}
	return c, nil
}

func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.clients))
	for name := range r.clients {
		names = append(names, name)
	}
	return names
}

func (r *Registry) Remove(name string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.clients, name)
}

func (r *Registry) Has(name string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()

	_, exists := r.clients[name]
	return exists
}