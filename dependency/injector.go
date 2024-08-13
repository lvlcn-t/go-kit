package dependency

import (
	"errors"
	"reflect"
	"sync"
)

// Injector is a dependency injector that can store and provide dependencies.
type Injector interface {
	// Provide registers a dependency with the DI container under a specified interface type.
	//
	// Parameters:
	//   - dep: The concrete implementation of the dependency you want to register.
	//   - iface: The [reflect.Type] of the interface under which the dependency should be registered. Defaults to the type of the provided dependency.
	//   - singleton: If true, the DI container will ensure that only one instance of the dependency is created and shared.
	//   - factory: An optional function that creates a new instance of the dependency. If provided, this function is used to create instances instead of the passed 'dep' value.
	//
	// Usage:
	//
	//	err := container.Provide(&MyConcreteType{}, reflect.TypeOf((*MyInterface)(nil)).Elem(), nil)
	//	if err != nil {
	//	    log.Fatal(err)
	//	}
	Provide(dep any, iface reflect.Type, singleton bool, factory func() any) error
	// Resolve returns either the first dependency of the given type or the given index.
	// Returns an error if the type is nil or if the dependency is not found.
	Resolve(t reflect.Type, i ...int) (any, error)
	// ResolveAll returns all dependencies of the given interface type.
	// Returns an error if the type is nil or if no dependencies are found.
	ResolveAll(t reflect.Type) ([]any, error)
	// Delete removes a dependency from the DI container.
	Delete(t reflect.Type)
}

// container is the default DI container
var container Injector = NewDIContainer()

// Provide registers a dependency with the DI container under a specified interface type.
//
// Parameters:
//   - dep: The dependency you want to register.
//   - singleton: If true, the DI container will ensure that only one instance of the dependency is created and shared.
//   - factory: An optional function that creates a new instance of the dependency. If provided, this function is used to create instances instead of the passed 'dep' value.
//
// Usage:
//
//	// Initialize a new instance of an implementation of MyInterface that should be registered as the ladder.
//	var instance MyInterface = &MyConcreteType{}
//	err := Provide(instance, false, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
func Provide[T any](dep T, singleton bool, factory func() T) error {
	fun := (func() any)(nil)
	if factory != nil {
		fun = func() any {
			return factory()
		}
	}
	return container.Provide(dep, reflect.TypeOf((*T)(nil)).Elem(), singleton, fun)
}

// Resolve returns either the first dependency of the given type or the given index.
// Returns an error if the type is nil or if the dependency is not found.
func Resolve[T any](i ...int) (T, error) {
	var empty T
	v, err := container.Resolve(reflect.TypeOf((*T)(nil)).Elem(), i...)
	if err != nil {
		return empty, err
	}
	return v.(T), nil
}

// ResolveAll returns all dependencies of the given interface type.
// Returns an error if the type is nil or if no dependencies are found.
func ResolveAll[T any]() ([]T, error) {
	v, err := container.ResolveAll(reflect.TypeOf((*T)(nil)).Elem())
	if err != nil {
		return nil, err
	}

	var results []T
	for _, dep := range v {
		results = append(results, dep.(T))
	}
	return results, nil
}

// Delete removes a dependency from the DI container.
func Delete[T any]() {
	container.Delete(reflect.TypeOf((*T)(nil)).Elem())
}

// injector is a dependency injector that can store and provide dependencies.
// It is the default implementation of the Injector interface.
type injector struct {
	// mu is a mutex that protects the dependencies map
	mu sync.RWMutex
	// dependencies is a map of types to their respective list of dependencies
	dependencies map[reflect.Type][]dependency
}

// dependency is a dependency that can be stored in the injector.
type dependency struct {
	// value is the value of the dependency
	value reflect.Value
	// factory is a function that creates a new instance of the dependency
	factory func() any
	// singleton is a flag that indicates if the dependency is a singleton
	singleton bool
	// once ensures the singleton is initialized only once
	once sync.Once
}

// NewDIContainer creates a new DI container.
func NewDIContainer() Injector {
	return &injector{
		mu:           sync.RWMutex{},
		dependencies: map[reflect.Type][]dependency{},
	}
}

// Provide registers a dependency with the DI container under a specified interface type.
//
// Parameters:
//   - dep: The concrete implementation of the dependency you want to register.
//   - iface: The [reflect.Type] of the interface under which the dependency should be registered. Defaults to the type of the provided dependency.
//   - singleton: If true, the DI container will ensure that only one instance of the dependency is created and shared.
//   - factory: An optional function that creates a new instance of the dependency. If provided, this function is used to create instances instead of the passed 'dep' value.
//
// Usage:
//
//	err := container.Provide(&MyConcreteType{}, reflect.TypeOf((*MyInterface)(nil)).Elem(), nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (c *injector) Provide(dep any, iface reflect.Type, singleton bool, factory func() any) error {
	if dep == nil && factory == nil {
		return errors.New("dependency and factory are both nil")
	}

	if iface == nil {
		iface = reflect.TypeOf(dep)
		if iface.Kind() == reflect.Pointer {
			iface = iface.Elem()
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.dependencies[iface] = append(c.dependencies[iface], dependency{
		value:     reflect.ValueOf(dep),
		factory:   factory,
		singleton: singleton,
	})
	return nil
}

// Resolve returns either the first dependency of the given type or the given index.
// Returns an error if the type is nil or if the dependency is not found.
func (c *injector) Resolve(t reflect.Type, i ...int) (any, error) {
	if t == nil {
		return nil, errors.New("type is nil")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	deps, ok := c.dependencies[t]
	if !ok || len(deps) == 0 {
		return nil, &errDependencyNotFound{}
	}

	index := 0
	if len(i) > 0 {
		index = i[0]
	}
	if index < -1 || index >= len(deps) {
		return nil, errors.New("index out of range")
	}
	if index == -1 {
		index = len(deps) - 1
	}

	dep := &deps[index]
	if dep.singleton {
		dep.once.Do(func() {
			if dep.factory != nil {
				instance := dep.factory()
				dep.value = reflect.ValueOf(instance)
			}
		})
		return dep.value.Interface(), nil
	}

	if dep.factory != nil {
		return dep.factory(), nil
	}
	return reflect.New(dep.value.Type()).Elem().Interface(), nil
}

// ResolveAll returns all dependencies of the given interface type.
// Returns an error if the type is nil or if no dependencies are found.
func (c *injector) ResolveAll(t reflect.Type) ([]any, error) {
	if t == nil {
		return nil, errors.New("type is nil")
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	deps, exists := c.dependencies[t]
	if !exists || len(deps) == 0 {
		return nil, &errDependencyNotFound{}
	}

	var results []any
	for i := range deps {
		dep := &deps[i]
		if dep.singleton {
			dep.once.Do(func() {
				if dep.factory != nil {
					instance := dep.factory()
					dep.value = reflect.ValueOf(instance)
				}
			})
			results = append(results, dep.value.Interface())
			continue
		}

		if dep.factory != nil {
			results = append(results, dep.factory())
			continue
		}

		results = append(results, dep.value.Interface())
	}

	return results, nil
}

// Delete removes a dependency from the DI container.
func (c *injector) Delete(t reflect.Type) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.dependencies, t)
}

var _ error = (*errDependencyNotFound)(nil)

type errDependencyNotFound struct{}

func (e errDependencyNotFound) Error() string {
	return "dependency not found"
}

func (e *errDependencyNotFound) Is(target error) bool {
	_, ok := target.(*errDependencyNotFound)
	if !ok {
		_, ok = target.(errDependencyNotFound)
	}
	return ok
}
