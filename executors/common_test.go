package executors

import (
	"context"
	"errors"
	"reflect"
	"sync/atomic"
	"testing"
	"time"

	"golang.org/x/time/rate"
)

func TestEffector_Run(t *testing.T) {
	applyAll := func(effector Effector, timeout time.Duration, rateLimit rate.Limit, circuitBreakerCount int, circuitBreakerTimeout time.Duration) Effector {
		return effector.
			WithRetry(Retrier{MaxRetries: 3, Backoff: func(retries uint) time.Duration { return 10 * time.Millisecond }}).
			WithTimeout(timeout).
			WithRateLimit(rateLimit).
			WithCircuitBreaker(circuitBreakerCount, circuitBreakerTimeout).
			WithProtection()
	}

	run := func(effector Effector, ctx context.Context) error { return effector.Do(ctx) }
	runGo := func(effector Effector, ctx context.Context) error {
		errChan := effector.Go(ctx)
		select {
		case err := <-errChan:
			return err
		case <-ctx.Done():
			return ctx.Err()
		}
	}

	tests := []struct {
		name     string
		effector Effector
		policies []func(Effector) Effector
		run      func(Effector, context.Context) error
		wantErr  bool
		errType  reflect.Type
	}{
		{
			name: "success",
			effector: func(ctx context.Context) error {
				return nil
			},
			run:     run,
			wantErr: false,
		},
		{
			name: "error",
			effector: func(ctx context.Context) error {
				return errors.New("task failed")
			},
			run:     run,
			wantErr: true,
			errType: reflect.TypeOf(errors.New("")),
		},
		{
			name: "timeout",
			effector: func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(50 * time.Millisecond):
					return errors.New("task failed")
				}
			},
			run: run,
			policies: []func(Effector) Effector{
				func(e Effector) Effector {
					return e.WithTimeout(10 * time.Millisecond)
				},
			},
			wantErr: true,
			errType: reflect.TypeOf(context.DeadlineExceeded),
		},
		{
			name: "rate limit",
			effector: func(ctx context.Context) error {
				return errors.New("task failed")
			},
			run: run,
			policies: []func(Effector) Effector{
				func(e Effector) Effector {
					return e.WithRateLimit(rate.Every(200 * time.Millisecond))
				},
			},
			wantErr: true,
			errType: reflect.TypeOf(errors.New("")),
		},
		{
			name: "circuit breaker",
			effector: func(ctx context.Context) error {
				return errors.New("task failed")
			},
			run: run,
			policies: []func(Effector) Effector{
				func(e Effector) Effector {
					return e.WithCircuitBreaker(1, 500*time.Millisecond)
				},
			},
			wantErr: true,
			errType: reflect.TypeOf(ErrCircuitOpen{}),
		},
		{
			name: "protection",
			effector: func(ctx context.Context) error {
				panic("task panicked")
			},
			run: run,
			policies: []func(Effector) Effector{
				func(e Effector) Effector {
					return e.WithProtection()
				},
			},
			wantErr: true,
			errType: reflect.TypeOf(errors.Join(errors.New(""))),
		},
		{
			name: "all policies with context timeout",
			effector: func(ctx context.Context) error {
				select {
				case <-ctx.Done():
					return ctx.Err()
				case <-time.After(50 * time.Millisecond):
					return errors.New("task failed")
				}
			},
			run: run,
			policies: []func(Effector) Effector{
				func(e Effector) Effector {
					return applyAll(e, 10*time.Millisecond, rate.Every(200*time.Millisecond), 2, 500*time.Millisecond)
				},
			},
			wantErr: true,
			errType: reflect.TypeOf(context.DeadlineExceeded),
		},
		{
			name:     "all policies with success",
			effector: noopEffector,
			run:      run,
			policies: []func(Effector) Effector{
				func(e Effector) Effector {
					return applyAll(e, 100*time.Millisecond, rate.Every(200*time.Millisecond), 1, 500*time.Millisecond)
				},
			},
			wantErr: false,
		},
		{
			name:     "all policies with success parallel",
			effector: noopEffector,
			run:      runGo,
			policies: []func(Effector) Effector{
				func(e Effector) Effector {
					return applyAll(e, 100*time.Millisecond, rate.Every(200*time.Millisecond), 1, 500*time.Millisecond)
				},
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effector := tt.effector
			for _, policy := range tt.policies {
				effector = policy(effector)
			}

			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
			defer cancel()

			err := tt.run(effector, ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Effector.Do() error = %v, wantErr %v", err, tt.wantErr)
			}

			if tt.errType != nil && reflect.TypeOf(err) != tt.errType {
				t.Errorf("Effector.Do() error = %v (%T), wantErr %v", err, err, tt.errType)
			}
		})
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
