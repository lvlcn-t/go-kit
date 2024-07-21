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
	tasks := []executors.Effector{
		executors.Effector(request),
		executors.Effector(request),
	}

	// Apply multiple policies to the tasks
	for i := range tasks {
		tasks[i] = tasks[i].WithRetry(executors.DefaultRetrier).
			WithTimeout(1*time.Second).
			WithRateLimit(executors.RateLimit(1)).
			WithCircuitBreaker(3, 1*time.Second)
	}

	// Run each task in a goroutine and collect the error channels
	var errChans []<-chan error
	for _, task := range tasks {
		errChans = append(errChans, task.Go(context.Background()))
	}

	// Wait for all tasks to finish
	var errs []error
	for _, errChan := range errChans {
		errs = append(errs, <-errChan)
	}

	// Check if any task may have failed
	for _, err := range errs {
		if err != nil {
			panic(err)
		}
	}
}

func request(ctx context.Context) error {
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
}
