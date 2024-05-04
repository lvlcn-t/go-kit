// executors package provides a set of common executors that can be used to perform given actions.
// An executor can be seen as a policy that wraps an action and applies some rules to it.
package executors

import (
	"context"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/time/rate"
)

// Effector is a function that performs an action and returns an error.
type Effector func(ctx context.Context) error

// noopEffector is an effector that does nothing.
func noopEffector(_ context.Context) error { return nil }

// Do runs the effector and returns the error.
// If a context is provided, it is used, otherwise a new context is created.
func (e Effector) Do(ctx ...context.Context) error {
	if len(ctx) > 0 {
		return e(ctx[0])
	}
	return e(context.Background())
}

// Go runs the effector concurrently and returns a channel that will receive the error.
// The channel is closed when the effector finishes.
func (e Effector) Go(ctx ...context.Context) <-chan error {
	ch := make(chan error, 1)
	go func() {
		ch <- e.Do(ctx...)
		close(ch)
	}()
	return ch
}

// WithRetry returns an effector that retries the effector a number of times with a delay between each retry.
func (e Effector) WithRetry(retrier Retrier) Effector {
	return retrier.Retry(e)
}

// WithTimeout returns an effector that runs the effector with a timeout.
func (e Effector) WithTimeout(timeout time.Duration) Effector {
	return Timeouter(timeout, e)
}

// WithRateLimit returns an effector that runs the effector with the specified rate limit.
func (e Effector) WithRateLimit(r rate.Limit) Effector {
	return RateLimiter(r, e)
}

// WithCircuitBreaker returns an effector that stops calling the task if it fails a certain number of times, until a certain amount of time has passed.
func (e Effector) WithCircuitBreaker(maxFailures int, resetTimeout time.Duration) Effector {
	return CircuitBreaker(maxFailures, resetTimeout, e)
}

// WithProtection returns an effector that recovers from panics and returns them as errors.
func (e Effector) WithProtection() Effector {
	return Protector(e)
}

// Parallel runs the effectors concurrently and returns the first error that occurs.
// Safe to use concurrently.
func Parallel(effectors ...Effector) Effector {
	return func(ctx context.Context) error {
		g, ctx := errgroup.WithContext(ctx)
		for _, effector := range effectors {
			g.Go(func() error { return effector(ctx) })
		}
		return g.Wait()
	}
}

// Sequential runs the effectors sequentially and returns the first error that occurs.
// Safe to use concurrently.
func Sequential(effectors ...Effector) Effector {
	return func(ctx context.Context) error {
		for _, effector := range effectors {
			if err := effector(ctx); err != nil {
				return err
			}
		}
		return nil
	}
}
