package apimanager

type ErrAlreadyRunning struct{}

func (e *ErrAlreadyRunning) Error() string {
	return "cannot mount routes while the server is running"
}

func (e *ErrAlreadyRunning) Is(target error) bool {
	_, ok := target.(*ErrAlreadyRunning)
	return ok
}

type ErrNotRunning struct{}

func (e *ErrNotRunning) Error() string {
	return "cannot stop the server because it is not running"
}

func (e *ErrNotRunning) Is(target error) bool {
	_, ok := target.(*ErrNotRunning)
	return ok
}
