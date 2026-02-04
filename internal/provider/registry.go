package provider

import (
	"fmt"
)

// Registry holds all available providers
type Registry struct {
	providers map[string]Provider
	defaultID string
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
		defaultID: "claude",
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(id string, p Provider) {
	r.providers[id] = p
}

// Get returns a provider by ID
func (r *Registry) Get(id string) (Provider, error) {
	p, ok := r.providers[id]
	if !ok {
		return nil, fmt.Errorf("provider not found: %s", id)
	}
	return p, nil
}

// GetDefault returns the default provider
func (r *Registry) GetDefault() Provider {
	p, _ := r.providers[r.defaultID]
	return p
}

// SetDefault sets the default provider ID
func (r *Registry) SetDefault(id string) error {
	if _, ok := r.providers[id]; !ok {
		return fmt.Errorf("provider not found: %s", id)
	}
	r.defaultID = id
	return nil
}

// List returns all registered provider IDs
func (r *Registry) List() []string {
	ids := make([]string, 0, len(r.providers))
	for id := range r.providers {
		ids = append(ids, id)
	}
	return ids
}

// DefaultID returns the default provider ID
func (r *Registry) DefaultID() string {
	return r.defaultID
}
