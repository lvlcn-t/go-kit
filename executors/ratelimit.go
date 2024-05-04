package executors

import (
	"context"
	"errors"
	"fmt"

	"golang.org/x/time/rate"
)

// ErrInvalidRateLimit is the error returned when the rate limit is invalid.
type ErrInvalidRateLimit struct{}

// Error returns the error message.
func (e ErrInvalidRateLimit) Error() string {
	return "invalid rate limit"
}

// ErrWaitRateLimit is the error returned when waiting for the rate limit fails.
type ErrWaitRateLimit struct {
	Err error
}

// Error returns the error message.
func (e ErrWaitRateLimit) Error() string {
	return fmt.Sprintf("wait rate limit: %v", e.Err)
}

// RateLimit is the rate limit for the rate limiter.
// It is an alias for rate.Limit.
//
// For more information, see https://pkg.go.dev/golang.org/x/time/rate#Limit.
type RateLimit = rate.Limit

// RateLimiter runs the effector with the specified rate limit.
func RateLimiter(r RateLimit, effector Effector) Effector {
	if effector == nil {
		return noopEffector
	}
	if r <= 0 {
		return func(_ context.Context) error {
			return ErrInvalidRateLimit{}
		}
	}

	limiter := rate.NewLimiter(r, 1)
	return func(ctx context.Context) error {
		if err := limiter.Wait(ctx); err != nil {
			if errors.Is(err, context.Canceled) {
				return err
			}
			return &ErrWaitRateLimit{Err: err}
		}
		return effector(ctx)
	}
}
