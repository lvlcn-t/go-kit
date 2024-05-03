package executors

import (
	"context"
	"time"
)

// Timeouter returns an effector that stops calling the task if it takes longer than the specified timeout.
func Timeouter(timeout time.Duration, effector Effector) Effector {
	if effector == nil {
		return noopEffector
	}

	return func(ctx context.Context) error {
		if timeout <= 0 {
			return context.DeadlineExceeded
		}
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return effector(ctx)
	}
}
