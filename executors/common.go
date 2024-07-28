// Package executors provides a set of common executors that can be used to perform given actions.
// An executor can be seen as a policy that wraps an action and applies some rules to it.
package executors

import (
	"context"
	"errors"
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

// WithFallback returns an effector that runs the fallback effector if the first one returns an error.
// Returns all errors that occurred as wrapped [errors.joinError] error.
func (e Effector) WithFallback(fallback Effector) Effector {
	return func(ctx context.Context) error {
		if err := e(ctx); err != nil {
			return errors.Join(err, fallback(ctx))
		}
		return nil
	}
}

// Concurrent returns an effector that runs the effectors concurrently and
// returns all errors that occurred as wrapped [errors.Join] error.
//
// Safe to use concurrently.
func Concurrent(effectors ...Effector) Effector {
	return func(ctx context.Context) error {
		g, ctx := errgroup.WithContext(ctx)
		errs := make(chan error, len(effectors))
		for _, effector := range effectors {
			effector := effector
			g.Go(func() error {
				err := effector(ctx)
				errs <- err
				return err
			})
		}
		go func() {
			// The returned error is ignored here on purpose,
			// as we are interested in all errors and not just the first one.
			_ = g.Wait()
			close(errs)
		}()

		var err error
		for e := range errs {
			err = errors.Join(err, e)
		}
		return err
	}
}

// Sequential returns an effector that runs the effectors sequentially and
// returns all errors that occurred as wrapped [errors.Join] error.
//
// Safe to use concurrently.
func Sequential(effectors ...Effector) Effector {
	return func(ctx context.Context) error {
		var errs []error
		for _, effector := range effectors {
			errs = append(errs, effector(ctx))
		}
		return errors.Join(errs...)
	}
}
