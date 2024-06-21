package fiberutils

import "fmt"

type ErrParameterNotFound struct {
	name string
}

func (e *ErrParameterNotFound) Error() string {
	return fmt.Sprintf("parameter %q not found", e.name)
}

func (e *ErrParameterNotFound) Is(target error) bool {
	_, ok := target.(*ErrParameterNotFound)
	return ok
}
