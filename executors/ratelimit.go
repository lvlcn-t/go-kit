package executors

import (
	"context"

	"golang.org/x/time/rate"
)

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
