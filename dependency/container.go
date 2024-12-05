package dependency

import (
	"iter"
	"reflect"
	"sync"
)

// Container is the dependency injection Container.
type Container struct {
	// mu protects the providers map from concurrent access.
	mu sync.RWMutex
	// providers is a map of providers by type.
	// One type can have multiple providers to support multiple implementations of the same interface.
	providers map[reflect.Type][]Provider
	// namedProviders is a map of providers by name.
	// Named providers can be resolved by their name.
	// Each provider must have a unique name otherwise it will be overwritten.
	namedProviders map[string]Provider
}

// NewContainer creates a new [Container].
func NewContainer() *Container {
	return &Container{
		providers:      map[reflect.Type][]Provider{},
		namedProviders: map[string]Provider{},
	}
}

// Provide adds the dependency providers to the container.
func (c *Container) Provide(providers ...Provider) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, provider := range providers {
		t := provider.Type()
		c.providers[t] = append(c.providers[t], provider)
		if provider.Name() != "" {
			c.namedProviders[provider.Name()] = provider
		}
	}
}

// Resolve returns the value of the provider with the provided type.
// If there are multiple providers with the same type, the first one is returned.
// Returns nil if the provider is not found.
//
// Note: You should always use [reflect.TypeFor][T]() to get the type you want to resolve.
func (c *Container) Resolve(t reflect.Type) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if providers, ok := c.providers[t]; ok {
		if len(providers) > 0 {
			return providers[0].Resolve()
		}
	}

	return nil
}

// ResolveNamed returns the value of the provider with the provided name.
// Returns nil if the provider is not found.
func (c *Container) ResolveNamed(name string) any {
	c.mu.RLock()
	defer c.mu.RUnlock()
	if provider, ok := c.namedProviders[name]; ok {
		return provider.Resolve()
	}

	return nil
}

// ResolveAll returns all the values of the provider with the provided type.
// Returns an empty slice if the provider is not found.
func (c *Container) ResolveAll(t reflect.Type) iter.Seq2[reflect.Type, any] {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return func(yield func(reflect.Type, any) bool) {
		if providers, ok := c.providers[t]; ok {
			for _, provider := range providers {
				if !yield(t, provider.Resolve()) {
					return
				}
			}
		}
	}
}
