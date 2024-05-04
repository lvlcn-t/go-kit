package executors

import (
	"context"
	"time"
)

// ErrCircuitOpen is the error returned when the circuit is open.
type ErrCircuitOpen struct{}

// Error returns the error message.
func (e ErrCircuitOpen) Error() string {
	return "circuit open"
}

// CircuitBreaker returns an effector that stops calling the task if it fails a certain number of times, until a certain amount of time has passed.
func CircuitBreaker(maxFailures int, resetTimeout time.Duration, effector Effector) Effector {
	if effector == nil {
		return noopEffector
	}

	// failures is the number of failures that have occurred.
	var failures int
	// lastFailure is the time of the last failure.
	var lastFailure time.Time

	return func(ctx context.Context) error {
		if failures >= maxFailures && time.Since(lastFailure) < resetTimeout {
			return ErrCircuitOpen{}
		}

		if err := effector(ctx); err != nil {
			failures++
			lastFailure = time.Now()
			if failures >= maxFailures {
				return ErrCircuitOpen{}
			}
			return err
		}

		failures = 0
		return nil
	}
}
