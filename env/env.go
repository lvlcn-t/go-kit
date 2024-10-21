package env

import (
	"fmt"
	"os"
	"reflect"
)

// Variable represents a builder for an environment variable of type T.
//
// The variable can be required or optional.
// Required variables must be set, while optional variables have a fallback value.
type Variable[T any] interface {
	// OrDie sets the die flag to panic if the environment variable is not set or an error occurs.
	OrDie() RequiredVariable[T]
	// WithFallback sets the default value to be used if the environment variable is not set or an error occurs.
	WithFallback(defaultValue T) OptionalVariable[T]
	// NoFallback flags the environment variable builder to not use a fallback value.
	// If the environment variable is not set or an error occurs, it returns the zero value of type T and the error.
	NoFallback() RequiredVariable[T]
}

// RequiredVariable represents a required environment variable.
//
// The environment variable must be set and always be a valid value.
type RequiredVariable[T any] interface{ TypedVariable[T] }

// OptionalVariable represents an optional environment variable.
//
// It always returns a valid value, either the value of the environment variable
// or the fallback value if the environment variable is not set or an error occurs.
type OptionalVariable[T any] interface{ TypedVariable[T] }

// TypedVariable represents a typed environment variable.
type TypedVariable[T any] interface {
	// Convert sets a custom converter function to convert the string value of the environment variable to type T.
	// If no converter is set, it tries to resolve a default converter based on the type T.
	//
	// Default converters are available for the following types:
	//  - string
	//  - int, int8, int16, int32, int64
	//  - uint, uint8, uint16, uint32, uint64, uintptr
	//  - float32, float64
	//  - complex64, complex128
	//  - bool
	//
	// If no default converter is available, a custom converter must be provided otherwise the environment variable value cannot be retrieved.
	Convert(converter Converter[T]) TypedVariable[T]
	// Value retrieves the environment variable value.
	//
	// If the environment variable is not set or an error occurs, it either returns preconditioned values and an error or panics if the die flag is set.
	Value() (T, error)
}

// variable represents the builder for an environment variable of type T.
type variable[T any] struct {
	key          string
	defaultValue T
	converter    Converter[T]
	die          bool
	isOptional   bool
}

// Get returns a [Variable] for the provided key.
// You can then use the builder to set fallback values, converters, and more.
//
// The type T must be a valid type for an environment variable. Panics if T is an invalid type for an environment variable.
//
// Example:
//
//	// Get the environment variable "MY_VAR" as a string and panic if an error occurs.
//	value := env.Get[string]("MY_VAR").OrDie().Value()
//
//	// Get the environment variable "MY_VAR" as an integer.
//	// If the variable is not set or an error occurs, return 42.
//	value := env.Get[int]("MY_VAR").WithFallback(42).Value()
//
//	// Get the environment variable "MY_VAR" as a [time.Duration] using a custom converter.
//	// If the variable is not set or an error occurs, return 5 * time.Second.
//	value := env.Get[time.Duration]("MY_VAR").WithFallback(5*time.Second).Convert(time.ParseDuration).Value()
//
//	// Get the environment variable "MY_VAR" as a boolean.
//	// If the variable is not set or an error occurs, return the error.
//	value, err := env.Get[bool]("MY_VAR").NoFallback().Value()
func Get[T any](key string) Variable[T] {
	if reflect.TypeFor[T]().Kind() == reflect.Interface {
		panic(fmt.Errorf("cannot use %v as type for an environment variable", reflect.TypeFor[T]()))
	}
	return &variable[T]{key: key}
}

// GetWithFallback retrieves an environment variable and tries to convert it to the desired type.
// If the environment variable is not set or the conversion fails, it returns the provided default value.
//
// Example:
//
//	// Get the environment variable "MY_VAR" as a string.
//	// If the variable is not set or an error occurs, return "default".
//	value := env.GetWithFallback("MY_VAR", "default")
//
//	// Get the environment variable "MY_VAR" as an integer.
//	// If the variable is not set or an error occurs, return 42.
//	value := env.GetWithFallback("MY_VAR", 42)
//
//	// Get the environment variable "MY_VAR" as a [time.Duration] using a custom converter.
//	// If the variable is not set or an error occurs, return 5 * time.Second.
//	value := env.GetWithFallback("MY_VAR", 5*time.Second, time.ParseDuration)
func GetWithFallback[T any](key string, defaultValue T, converter ...Converter[T]) T {
	var c Converter[T]
	if len(converter) > 0 {
		c = converter[0]
	}

	v, err := Get[T](key).WithFallback(defaultValue).Convert(c).Value()
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
	}
	return v
}

// OrDie sets the die flag to panic if the environment variable is not set or an error occurs.
//
// Panics when used in combination with WithFallback.
func (v *variable[T]) OrDie() RequiredVariable[T] {
	if v.isOptional {
		// This shouldn't happen, but just in case the user decides
		// to type assert the builder to [Variable] and call OrDie.
		// e.g. env.Get[string]("test").WithFallback("default").(Variable[string]).OrDie()
		panic("Cannot use OrDie with WithFallback")
	}
	v.die = true
	return v
}

// NoFallback flags the environment variable builder to not use a fallback value.
//
// Panics when used in combination with WithFallback.
func (v *variable[T]) NoFallback() RequiredVariable[T] {
	if v.isOptional {
		// This shouldn't happen, but just in case the user decides
		// to type assert the builder to [Variable] and call NoFallback.
		// e.g. env.Get[string]("test").WithFallback("default").(Variable[string]).NoFallback()
		panic("Cannot use NoFallback with WithFallback")
	}
	return v
}

// WithFallback sets the default value to be used if the environment variable is not set or conversion fails.
func (v *variable[T]) WithFallback(defaultValue T) OptionalVariable[T] {
	v.defaultValue = defaultValue
	v.isOptional = true
	return v
}

// Convert sets a custom converter function to convert the string value to type T.
func (v *variable[T]) Convert(converter Converter[T]) TypedVariable[T] {
	v.converter = converter
	return v
}

// Value retrieves the environment variable value.
func (v *variable[T]) Value() (T, error) {
	val, ok := os.LookupEnv(v.key)
	if !ok {
		return v.handleError(nil)
	}

	if v.converter == nil {
		var err error
		v.converter, err = defaultConverter[T]()
		if err != nil {
			return v.handleError(fmt.Errorf("no default converter for type %T: %w", v.defaultValue, err))
		}
	}

	value, err := v.converter(val)
	if err != nil {
		return v.handleError(fmt.Errorf("failed to convert value %q: %w", val, err))
	}

	return value, nil
}

// handleError returns the default value and the provided error.
// If the die flag is set, it panics with the error.
func (v *variable[T]) handleError(err error) (T, error) {
	if v.isOptional {
		return v.defaultValue, err
	}

	if err == nil {
		err = fmt.Errorf("environment variable %q is required", v.key)
	} else {
		err = fmt.Errorf("failed to get environment variable %q: %w", v.key, err)
	}

	if v.die {
		panic(err)
	}
	return v.defaultValue, err
}
