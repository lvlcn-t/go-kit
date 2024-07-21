package main

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/lvlcn-t/go-kit/executors"
)

func main() {
	// Create a task that may fail and needs to be retried
	task := executors.Effector(func(ctx context.Context) (err error) {
		// Do something that may fail. In this case, make an HTTP request.
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://example.com", http.NoBody)
		if err != nil {
			return err
		}

		resp, err := http.DefaultClient.Do(req)
		if err != nil {
			return err
		}
		defer func() {
			err = errors.Join(err, resp.Body.Close())
		}()

		if resp.StatusCode != http.StatusOK {
			return errors.New("unexpected status code")
		}

		return nil
	})

	// Apply multiple policies to the task
	task = task.WithRetry(executors.DefaultRetrier).
		WithTimeout(1*time.Second).
		WithRateLimit(executors.RateLimit(1)).
		WithCircuitBreaker(3, 1*time.Second)

	// Run the task with the applied policies and a context
	err := task.Do(context.Background())
	if err != nil {
		panic(err)
	}
}
