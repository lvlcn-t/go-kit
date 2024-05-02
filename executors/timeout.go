package executors

import (
	"context"
	"time"
)

func Timeout(timeout time.Duration, effector Effector) Effector {
	return func(ctx context.Context) error {
		ctx, cancel := context.WithTimeout(ctx, timeout)
		defer cancel()
		return effector(ctx)
	}
}
