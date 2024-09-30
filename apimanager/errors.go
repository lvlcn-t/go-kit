package apimanager

// ErrAlreadyRunning is an error that is returned when the server is already running.
type ErrAlreadyRunning struct{}

// Error returns the error message.
func (e *ErrAlreadyRunning) Error() string {
	return "cannot mount routes while the server is running"
}

// Is checks if the target error is an [ErrAlreadyRunning] error.
func (e *ErrAlreadyRunning) Is(target error) bool {
	_, ok := target.(*ErrAlreadyRunning)
	return ok
}

// ErrNotRunning is an error that is returned when the server is not running.
type ErrNotRunning struct{}

// Error returns the error message.
func (e *ErrNotRunning) Error() string {
	return "cannot stop the server because it is not running"
}

// Is checks if the target error is an [ErrNotRunning] error.
func (e *ErrNotRunning) Is(target error) bool {
	_, ok := target.(*ErrNotRunning)
	return ok
}
