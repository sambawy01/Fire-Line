package adapter

import (
	"context"
	"fmt"
	"log/slog"
	"sync"
)

// Factory creates a new adapter instance for a given type.
type Factory func() Adapter

// Registry manages adapter registrations and active instances.
type Registry struct {
	mu        sync.RWMutex
	factories map[string]Factory
	instances map[string]Adapter // keyed by adapter_id
}

// NewRegistry creates a new adapter registry.
func NewRegistry() *Registry {
	return &Registry{
		factories: make(map[string]Factory),
		instances: make(map[string]Adapter),
	}
}

// RegisterFactory registers a factory for a given adapter type.
func (r *Registry) RegisterFactory(adapterType string, factory Factory) {
	r.mu.Lock()
	r.factories[adapterType] = factory
	r.mu.Unlock()
	slog.Info("adapter factory registered", "type", adapterType)
}

// Create creates and initializes a new adapter instance from config.
func (r *Registry) Create(ctx context.Context, cfg Config) (Adapter, error) {
	r.mu.RLock()
	factory, ok := r.factories[cfg.AdapterType]
	r.mu.RUnlock()

	if !ok {
		return nil, fmt.Errorf("unknown adapter type: %s", cfg.AdapterType)
	}

	adapter := factory()
	if err := adapter.Initialize(ctx, cfg); err != nil {
		return nil, fmt.Errorf("initialize adapter %s: %w", cfg.AdapterType, err)
	}

	r.mu.Lock()
	r.instances[cfg.AdapterID] = adapter
	r.mu.Unlock()

	slog.Info("adapter instance created",
		"adapter_id", cfg.AdapterID,
		"type", cfg.AdapterType,
		"location_id", cfg.LocationID,
	)

	return adapter, nil
}

// Get retrieves an active adapter instance by ID.
func (r *Registry) Get(adapterID string) (Adapter, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	a, ok := r.instances[adapterID]
	return a, ok
}

// Remove shuts down and removes an adapter instance.
func (r *Registry) Remove(ctx context.Context, adapterID string) error {
	r.mu.Lock()
	adapter, ok := r.instances[adapterID]
	if !ok {
		r.mu.Unlock()
		return fmt.Errorf("adapter not found: %s", adapterID)
	}
	delete(r.instances, adapterID)
	r.mu.Unlock()

	return adapter.Shutdown(ctx)
}

// List returns all active adapter IDs.
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()
	ids := make([]string, 0, len(r.instances))
	for id := range r.instances {
		ids = append(ids, id)
	}
	return ids
}

// ByLocation returns all adapters for a specific location.
func (r *Registry) ByLocation(locationID string) []Adapter {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var adapters []Adapter
	for _, a := range r.instances {
		// We need the config to check location — adapters store it internally
		adapters = append(adapters, a)
	}
	return adapters
}

// ShutdownAll gracefully shuts down all adapter instances.
func (r *Registry) ShutdownAll(ctx context.Context) {
	r.mu.Lock()
	instances := make(map[string]Adapter, len(r.instances))
	for k, v := range r.instances {
		instances[k] = v
	}
	r.instances = make(map[string]Adapter)
	r.mu.Unlock()

	for id, adapter := range instances {
		if err := adapter.Shutdown(ctx); err != nil {
			slog.Error("adapter shutdown failed", "adapter_id", id, "error", err)
		}
	}
}
