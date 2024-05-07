package config

import "fmt"

// ErrFieldRequired is an error type that indicates a required field is missing in the configuration
type ErrFieldRequired struct {
	Field   string
	Type    string
	Message string
}

// Error returns the error message.
func (e ErrFieldRequired) Error() string {
	return fmt.Sprintf("field %q (%s) is required: %s", e.Field, e.Type, e.Message)
}

// Is returns true if the target error is an ErrFieldRequired.
func (e ErrFieldRequired) Is(target error) bool {
	_, ok := target.(ErrFieldRequired)
	return ok
}

// ErrFieldInvalid is an error type that indicates an invalid field in the configuration
type ErrFieldInvalid struct {
	Field   string
	Type    string
	Message string
}

// Error returns the error message.
func (e ErrFieldInvalid) Error() string {
	return fmt.Sprintf("field %q (%s) is invalid: %s", e.Field, e.Type, e.Message)
}

// Is returns true if the target error is an ErrFieldInvalid.
func (e ErrFieldInvalid) Is(target error) bool {
	_, ok := target.(ErrFieldInvalid)
	return ok
}
