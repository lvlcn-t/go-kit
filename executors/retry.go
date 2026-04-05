package executors

import (
	"context"
	"time"
)

// Backoff is a function that returns the duration to wait before the next retry
type Backoff func(retries uint) time.Duration

// Retrier is a struct that retries an action a number of times with a delay between each retry.
type Retrier struct {
	MaxRetries int
	Backoff    Backoff
}

// DefaultRetrier is the default retrier that retries 3 times with the default backoff.
var DefaultRetrier = Retrier{
	MaxRetries: 3,
	Backoff:    DefaultBackoff,
}

// Retry retries the effector a number of times with a delay between each retry.
// If no backoff function is provided, the default backoff function is used.
func (r *Retrier) Retry(effector Effector) Effector {
	if effector == nil {
		return noopEffector
	}

	if r.Backoff == nil {
		r.Backoff = DefaultBackoff
	}

	return func(ctx context.Context) (err error) {
		for i := 0; i < r.MaxRetries; i++ {
			err = effector(ctx)
			if err == nil {
				return nil
			}
			select {
			case <-ctx.Done():
				return ctx.Err()
			case <-time.After(r.Backoff(uint(i))):
			}
		}
		return err
	}
}

// Retry retries the effector a number of times with a delay between each retry.
// If the backoff function of the default retrier is nil, the default backoff function is used.
// Safe to use concurrently.
func Retry(effector Effector) Effector {
	return DefaultRetrier.Retry(effector)
}

// SetMaxRetries sets the maximum number of retries.
// Not safe to use concurrently.
func SetMaxRetries(maximum int) {
	DefaultRetrier.MaxRetries = maximum
}

// SetBackoff sets the backoff function.
// Not safe to use concurrently.
func SetBackoff(backoff Backoff) {
	DefaultRetrier.Backoff = backoff
}

// DefaultBackoff calculates the delay for the next retry.
// The delay is calculated as 2^retries seconds.
func DefaultBackoff(retries uint) time.Duration {
	return time.Duration(1<<retries) * time.Second
}
