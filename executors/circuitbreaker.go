package executors

import (
	"context"
	"errors"
	"time"
)

// ErrCircuitOpen is the error returned when the circuit is open.
var ErrCircuitOpen = errors.New("circuit open")

// CircuitBreaker returns an effector that stops calling the task if it fails a certain number of times, until a certain amount of time has passed.
func CircuitBreaker(maxFailures int, resetTimeout time.Duration, effector Effector) Effector {
	// failures is the number of failures that have occurred.
	var failures int
	// lastFailure is the time of the last failure.
	var lastFailure time.Time

	return func(ctx context.Context) error {
		if failures >= maxFailures && time.Since(lastFailure) < resetTimeout {
			return ErrCircuitOpen
		}

		if err := effector(ctx); err != nil {
			failures++
			lastFailure = time.Now()
			if failures >= maxFailures {
				return ErrCircuitOpen
			}
			return err
		}

		failures = 0
		return nil
	}
}
