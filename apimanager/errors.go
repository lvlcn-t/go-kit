package apimanager

type ErrAlreadyRunning struct{}

func (e *ErrAlreadyRunning) Error() string {
	return "cannot mount routes while the server is running"
}

func (e *ErrAlreadyRunning) Is(target error) bool {
	_, ok := target.(*ErrAlreadyRunning)
	return ok
}
