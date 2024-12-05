package dependency

import (
	"reflect"
	"sync"
)

// Provider represents a dependency provider.
type Provider interface {
	// Name returns the name of the provider.
	Name() string
	// Type returns the type of the provider.
	Type() reflect.Type
	// Resolve returns the value of the provider.
	Resolve() any
	// Named sets the name of the provider.
	Named(name string) Provider
}

// baseProvider provides the base implementation of the [Provider] interface.
type baseProvider struct {
	name  string
	iface reflect.Type
}

// newBaseProvider creates a new [baseProvider] with the provided name and interface.
func newBaseProvider[T any]() baseProvider {
	var iface reflect.Type
	t := reflect.TypeFor[T]()
	if t.Kind() == reflect.Interface {
		iface = t
	}

	return baseProvider{iface: iface}
}

// Name returns the name of the provider.
func (b *baseProvider) Name() string {
	return b.name
}

// Singleton is a [Provider] that returns the same value every time it is resolved.
type Singleton struct {
	baseProvider
	value reflect.Value
}

// NewSingleton creates a new [Singleton] provider with the provided value.
func NewSingleton[T any](value T) Provider {
	return &Singleton{
		baseProvider: newBaseProvider[T](),
		value:        reflect.ValueOf(value),
	}
}

// Named sets the name of the provider.
func (s *Singleton) Named(name string) Provider {
	s.name = name
	return s
}

// Type returns the type of the provider.
func (s *Singleton) Type() reflect.Type {
	if s.iface != nil {
		return s.iface
	}
	return s.value.Type()
}

// Resolve returns the singleton value.
func (s *Singleton) Resolve() any {
	if s.iface != nil {
		return s.value.Convert(s.iface).Interface()
	}
	return s.value.Interface()
}

// Factory is a [Provider] that returns a new value every time it is resolved.
type Factory struct {
	baseProvider
	factory func() any
}

// NewFactory creates a new [Factory] provider with the provided factory function.
func NewFactory[T any](factory func() T) Provider {
	return &Factory{
		baseProvider: newBaseProvider[T](),
		factory:      func() any { return factory() },
	}
}

// Named sets the name of the provider.
func (f *Factory) Named(name string) Provider {
	f.name = name
	return f
}

// Type returns the type of the provider.
func (f *Factory) Type() reflect.Type {
	if f.iface != nil {
		return f.iface
	}
	return reflect.TypeOf(f.factory())
}

// Resolve returns a new value from the factory.
func (f *Factory) Resolve() any {
	v := f.factory()
	if f.iface != nil {
		return reflect.ValueOf(v).Convert(f.iface).Interface()
	}
	return v
}

// SingletonFunc is a [Provider] that returns the same value every time it is resolved.
type SingletonFunc struct {
	Factory
	value any
	once  sync.Once
}

// NewSingletonFunc creates a new [SingletonFunc] provider with the provided factory function.
func NewSingletonFunc[T any](factory func() T) Provider {
	return &SingletonFunc{
		Factory: Factory{
			baseProvider: newBaseProvider[T](),
			factory:      func() any { return factory() },
		},
	}
}

// Named sets the name of the provider.
func (f *SingletonFunc) Named(name string) Provider {
	f.name = name
	return f
}

// Type returns the type of the provider.
func (f *SingletonFunc) Type() reflect.Type {
	if f.iface != nil {
		return f.iface
	}
	f.once.Do(func() {
		f.value = f.factory()
	})
	return reflect.TypeOf(f.value)
}

// Resolve returns the singleton value.
func (f *SingletonFunc) Resolve() any {
	f.once.Do(func() {
		f.value = f.factory()
	})
	if f.iface != nil {
		return reflect.ValueOf(f.value).Convert(f.iface).Interface()
	}
	return f.value
}
