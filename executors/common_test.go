package executors

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestEffector_MultiplePolicies(t *testing.T) {
	task := func(ctx context.Context) error {
		// This task simulates a condition that can timeout or be rate-limited
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(50 * time.Millisecond):
			return errors.New("task failed")
		}
	}

	retrier := Retrier{MaxRetries: 3, Backoff: func(retries int) time.Duration { return 10 * time.Millisecond }}
	timeout := 100 * time.Millisecond
	rateLimit := rate.Every(200 * time.Millisecond)
	maxFailures := 1
	resetTimeout := 500 * time.Millisecond

	effector := Effector(task).
		WithRetry(retrier).
		WithTimeout(timeout).
		WithRateLimit(rateLimit).
		WithCircuitBreaker(maxFailures, resetTimeout)

	ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
	defer cancel()

	err := effector.Do(ctx)
	if err == nil {
		t.Error("Expected an error, but got none")
	}

	// We cannot predict the exact error, but it should be either a deadline exceeded error or a circuit open error
	if !errors.Is(err, context.DeadlineExceeded) && !errors.Is(err, ErrCircuitOpen) {
		t.Errorf("Expected either %q or a %q error, but got %v", context.DeadlineExceeded.Error(), ErrCircuitOpen.Error(), err)
	}
}

func TestParallel(t *testing.T) {
	tests := []struct {
		name      string
		effectors []Effector
		wantErr   bool
	}{
		{
			name:    "nil effectors",
			wantErr: false,
		},
		{
			name: "success",
			effectors: []Effector{
				func(ctx context.Context) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "one error",
			effectors: []Effector{
				func(ctx context.Context) error {
					return nil
				},
				func(ctx context.Context) error {
					return errors.New("task failed")
				},
			},
			wantErr: true,
		},
		{
			name: "all errors",
			effectors: []Effector{
				func(ctx context.Context) error {
					return errors.New("task failed")
				},
				func(ctx context.Context) error {
					return errors.New("task failed")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effector := Parallel(tt.effectors...)

			ctx, cancel := context.WithTimeout(context.Background(), 1*time.Second)
			defer cancel()

			err := effector.Do(ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parallel() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestSequential(t *testing.T) {
	var executionOrder int32
	tests := []struct {
		name      string
		effectors []Effector
		wantErr   bool
	}{
		{
			name:    "nil effectors",
			wantErr: false,
		},
		{
			name: "success",
			effectors: []Effector{
				func(ctx context.Context) error {
					return nil
				},
			},
			wantErr: false,
		},
		{
			name: "one error",
			effectors: []Effector{
				func(ctx context.Context) error {
					atomic.StoreInt32(&executionOrder, 1)
					return nil
				},
				func(ctx context.Context) error {
					if atomic.LoadInt32(&executionOrder) != 1 {
						t.Error("Expected the first effector to run first")
					}
					return errors.New("task failed")
				},
			},
			wantErr: true,
		},
		{
			name: "all errors",
			effectors: []Effector{
				func(ctx context.Context) error {
					atomic.StoreInt32(&executionOrder, 1)
					return errors.New("task failed")
				},
				func(ctx context.Context) error {
					if atomic.LoadInt32(&executionOrder) != 1 {
						t.Error("Expected the first effector to run first")
					}
					return errors.New("task failed")
				},
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		executionOrder = 0
		t.Run(tt.name, func(t *testing.T) {
			effector := Sequential(tt.effectors...)

			err := effector.Do()
			if (err != nil) != tt.wantErr {
				t.Errorf("Sequential() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
