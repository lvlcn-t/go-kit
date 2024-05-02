package executors

import (
	"context"
	"time"
)

// Timeout returns an effector that stops calling the task if it takes longer than the specified timeout.
func Timeout(timeout time.Duration, effector Effector) Effector {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return effector(ctx)
	}
}
