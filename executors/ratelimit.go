package executors

import (
	"context"
	"errors"

	"golang.org/x/time/rate"
)

var ErrInvalidRate = errors.New("invalid rate limit")

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
			return ErrInvalidRate
		}
	}

	limiter := rate.NewLimiter(r, 1)
	return func(ctx context.Context) error {
		if err := limiter.Wait(ctx); err != nil {
			return err
		}
		return effector(ctx)
	}
}
