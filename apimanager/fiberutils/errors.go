package fiberutils

import "fmt"

// ErrParameterNotFound is an error that is returned when a parameter is not found.
type ErrParameterNotFound struct {
	name string
}

// Error returns the error message.
func (e *ErrParameterNotFound) Error() string {
	return fmt.Sprintf("parameter %q not found", e.name)
}

// Is checks if the target is an [ErrParameterNotFound].
func (e *ErrParameterNotFound) Is(target error) bool {
	_, ok := target.(*ErrParameterNotFound)
	return ok
}
