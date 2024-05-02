package executors

import (
	"context"

	"golang.org/x/time/rate"
)

// RateLimit is the rate limit for the rate limiter.
// It is an alias for rate.Limit.
//
// For more information, see https://pkg.go.dev/golang.org/x/time/rate#Limit.
type RateLimit = rate.Limit

// RateLimiter runs the effector with the specified rate limit.
func RateLimiter(r RateLimit, effector Effector) Effector {
	limiter := rate.NewLimiter(r, 1)
	return func(ctx context.Context) error {
		if err := limiter.Wait(ctx); err != nil {
			return err
		}
		return effector(ctx)
	}
}
