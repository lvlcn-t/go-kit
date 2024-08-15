package dependency

import (
	"errors"
	"fmt"
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
	//   - name: An optional name for the dependency. If not provided, the name defaults to the type name of the dependency.
	//
	// Usage:
	//
	//	err := container.Provide(&MyConcreteType{}, reflect.TypeOf((*MyInterface)(nil)).Elem(), nil)
	//	if err != nil {
	//	    log.Fatal(err)
	//	}
	Provide(dep any, iface reflect.Type, singleton bool, factory func() any, name ...string) error
	// Resolve returns a dependency of the given interface type and name.
	// Returns an error if the type is nil or if the dependency is not found.
	//
	// The following rules apply when resolving a dependency:
	//   - If no name is provided, it tries to resolve the dependency by the type name.
	//     If the name is not found, it resolves the first dependency of the given type.
	//   - If an empty string is provided as the name, it resolves the first dependency of the given type.
	//   - If "-1" is provided as the name, it resolves the last dependency of the given type.
	Resolve(t reflect.Type, name ...string) (any, error)
	// ResolveAll returns all dependencies of the given interface type.
	// Returns an error if the type is nil or if no dependencies are found.
	ResolveAll(t reflect.Type) ([]any, error)
	// Delete removes a dependency from the DI container.
	Delete(t reflect.Type)
}

// container is the default DI container
var container Injector = NewDIContainer()

// Provide registers a dependency with the DI container under the given type.
//
// Parameters:
//   - dep: The dependency you want to register.
//   - singleton: If true, the DI container will ensure that only one instance of the dependency is created and shared.
//   - factory: An optional function that creates a new instance of the dependency. If provided, this function is used to create instances instead of the passed 'dep' value.
//   - name: An optional name for the dependency. If not provided, the name defaults to the type name of the dependency.
//
// Usage:
//
//	err := dependency.Provide[MyInterface](&MyConcreteType{}, false, nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
func Provide[T any](dep T, singleton bool, factory func() T, name ...string) error {
	fun := (func() any)(nil)
	if factory != nil {
		fun = func() any {
			return factory()
		}
	}
	return container.Provide(dep, reflect.TypeOf((*T)(nil)).Elem(), singleton, fun, name...)
}

// Resolve returns a dependency of the given type and name.
// Returns an error if the type is nil or if the dependency is not found.
//
// The following rules apply when resolving a dependency:
//   - If no name is provided, it tries to resolve the dependency by the type name.
//     If the name is not found, it resolves the first dependency of the given type.
//   - If an empty string is provided as the name, it resolves the first dependency of the given type.
//   - If "-1" is provided as the name, it resolves the last dependency of the given type.
func Resolve[T any](name ...string) (T, error) {
	var empty T
	v, err := container.Resolve(reflect.TypeOf((*T)(nil)).Elem(), name...)
	if err != nil {
		return empty, err
	}
	return v.(T), nil
}

// ResolveAll returns all dependencies of the given type.
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
	// registry is a map of dependency names to their respective interface type and implementation type
	registry map[string]dependencyInfo
}

// dependencyInfo is a struct that holds information about a dependency.
type dependencyInfo struct {
	// iface is the interface type of the dependency
	iface reflect.Type
	// dep is the implementation type of the dependency
	dep reflect.Type
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
		registry:     map[string]dependencyInfo{},
	}
}

// Provide registers a dependency with the DI container under a specified interface type.
//
// Parameters:
//   - dep: The concrete implementation of the dependency you want to register.
//   - iface: The [reflect.Type] of the interface under which the dependency should be registered. Defaults to the type of the provided dependency.
//   - singleton: If true, the DI container will ensure that only one instance of the dependency is created and shared.
//   - factory: An optional function that creates a new instance of the dependency. If provided, this function is used to create instances instead of the passed 'dep' value.
//   - name: An optional name for the dependency. If not provided, the name defaults to the type name of the dependency.
//
// Usage:
//
//	err := container.Provide(&MyConcreteType{}, reflect.TypeOf((*MyInterface)(nil)).Elem(), nil)
//	if err != nil {
//	    log.Fatal(err)
//	}
func (c *injector) Provide(dep any, iface reflect.Type, singleton bool, factory func() any, name ...string) error {
	if len(name) > 0 {
		if name[0] == "-1" {
			return fmt.Errorf("name cannot be a reserved keyword: %q", name[0])
		}
		return c.provide(dep, iface, name[0], singleton, factory)
	}
	return c.provide(dep, iface, "", singleton, factory)
}

// Resolve returns a dependency of the given interface type and name.
// Returns an error if the type is nil or if the dependency is not found.
//
// The following rules apply when resolving a dependency:
//   - If no name is provided, it tries to resolve the dependency by the type name.
//     If the name is not found, it resolves the first dependency of the given type.
//   - If an empty string is provided as the name, it resolves the first dependency of the given type.
//   - If "-1" is provided as the name, it resolves the last dependency of the given type.
func (c *injector) Resolve(t reflect.Type, name ...string) (any, error) {
	if t == nil {
		return nil, errors.New("type is nil")
	}

	if len(name) > 0 {
		return c.resolve(t, name[0])
	}
	return c.resolve(t, t.String())
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
		return nil, &ErrDependencyNotFound{}
	}

	var results []any
	for i := range deps {
		instance, err := deps[i].instantiate()
		if err != nil {
			return nil, err
		}
		results = append(results, instance)
	}

	return results, nil
}

// Delete removes a dependency from the DI container.
func (c *injector) Delete(t reflect.Type) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.dependencies, t)
	for name, info := range c.registry {
		if info.iface == t {
			delete(c.registry, name)
		}
	}
}

// provide registers a dependency with the DI container.
func (c *injector) provide(dep any, iface reflect.Type, name string, singleton bool, factory func() any) error {
	if dep == nil && factory == nil {
		return errors.New("dependency and factory are both nil")
	}

	if dep == nil {
		dep = factory()
		if dep == nil {
			return errors.New("dependency is nil")
		}
	}

	if iface == nil {
		iface = reflect.TypeOf(dep)
		if iface.Kind() == reflect.Pointer {
			iface = iface.Elem()
		}
	}

	switch iface.Kind() {
	case reflect.Interface:
		if !reflect.TypeOf(dep).Implements(iface) {
			return fmt.Errorf("dependency (%s) does not implement interface (%s)", reflect.TypeOf(dep), iface)
		}
	default:
		t := reflect.TypeOf(dep)
		if t.Kind() == reflect.Pointer {
			t = t.Elem()
		}

		if t.Kind() != iface.Kind() {
			return fmt.Errorf("dependency type (%s) does not match interface type (%s)", reflect.TypeOf(dep), iface)
		}
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	c.dependencies[iface] = append(c.dependencies[iface], dependency{
		value:     reflect.ValueOf(dep),
		factory:   factory,
		singleton: singleton,
	})

	if name == "" {
		name = reflect.TypeOf(dep).String()
	}
	c.registry[name] = dependencyInfo{
		iface: iface,
		dep:   reflect.TypeOf(dep),
	}
	return nil
}

// resolve returns a dependency of the given interface type and name.
func (c *injector) resolve(t reflect.Type, name string) (any, error) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	deps, ok := c.dependencies[t]
	if !ok || len(deps) == 0 {
		return nil, &ErrDependencyNotFound{}
	}

	switch name {
	case "":
		return deps[0].instantiate()
	case "-1":
		return deps[len(deps)-1].instantiate()
	}

	info, ok := c.registry[name]
	if !ok {
		return nil, &ErrDependencyNotFound{}
	}

	if info.iface != t {
		return nil, errors.New("interface type does not match")
	}

	index := 0
	for i := range deps {
		if deps[i].value.Type() == info.dep {
			index = i
			break
		}
	}

	return deps[index].instantiate()
}

// instantiate creates a new instance of a dependency.
func (d *dependency) instantiate() (any, error) {
	if !d.value.IsValid() && d.factory == nil {
		return nil, errors.New("dependency is not valid")
	}

	if d.singleton {
		d.once.Do(func() {
			if d.factory != nil {
				instance := d.factory()
				d.value = reflect.ValueOf(instance)
			}
		})
		return d.value.Interface(), nil
	}

	if d.factory != nil {
		return d.factory(), nil
	}
	return reflect.New(d.value.Type()).Elem().Interface(), nil
}

var _ error = (*ErrDependencyNotFound)(nil)

// ErrDependencyNotFound is an error that is returned when a dependency is not found.
type ErrDependencyNotFound struct{}

// Error returns the error message.
func (e ErrDependencyNotFound) Error() string {
	return "dependency not found"
}

// Is checks if the target error is an [ErrDependencyNotFound].
func (e *ErrDependencyNotFound) Is(target error) bool {
	_, ok := target.(*ErrDependencyNotFound)
	if !ok {
		_, ok = target.(ErrDependencyNotFound)
	}
	return ok
}
