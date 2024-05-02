package executors

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestRetrier_Retry(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		retrier  *Retrier
		effector Effector
		wantErr  bool
	}{
		{
			name:     "nil effector",
			retrier:  &Retrier{},
			effector: nil,
			wantErr:  false,
		},
		{
			name:    "nil backoff",
			retrier: &Retrier{},
			effector: func(ctx context.Context) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "success",
			retrier: &Retrier{
				MaxRetries: 3,
				Backoff: func(retries int) time.Duration {
					return 1
				},
			},
			effector: func(ctx context.Context) error {
				return nil
			},
			wantErr: false,
		},
		{
			name:    "success with default retrier",
			retrier: nil,
			effector: func(ctx context.Context) error {
				return nil
			},
			wantErr: false,
		},
		{
			name: "error",
			retrier: &Retrier{
				MaxRetries: 3,
				Backoff: func(retries int) time.Duration {
					return 0
				},
			},
			effector: func(ctx context.Context) error {
				return errors.New("error while executing")
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effector := Retry(tt.effector)
			if tt.retrier != nil {
				effector = tt.retrier.Retry(tt.effector)
			}

			err := effector(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Retrier.Retry() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRetrier_Retry_Context_Cancel(t *testing.T) {
	t.Parallel()

	retrier := &Retrier{
		MaxRetries: 3,
		Backoff: func(retries int) time.Duration {
			return 10 * time.Millisecond
		},
	}

	effector := func(ctx context.Context) error {
		return errors.New("error while executing")
	}

	ctx, cancel := context.WithCancel(context.Background())

	go func() {
		time.Sleep(5 * time.Millisecond)
		cancel()
	}()

	err := retrier.Retry(effector)(ctx)
	if err == nil {
		t.Errorf("Retrier.Retry() error = %v, wantErr %v", err, true)
	}
}

func TestDefaultRetrier_SetBackoff(t *testing.T) {
	tests := []struct {
		name    string
		backoff Backoff
	}{
		{
			name:    "nil backoff",
			backoff: nil,
		},
		{
			name: "success",
			backoff: func(retries int) time.Duration {
				return 1
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetBackoff(tt.backoff)
			if reflect.ValueOf(DefaultRetrier.Backoff).Pointer() != reflect.ValueOf(tt.backoff).Pointer() {
				t.Errorf("DefaultRetrier.SetBackoff() backoff = %v, want %v", DefaultRetrier.Backoff, tt.backoff)
			}
		})
	}
}

func TestDefaultRetrier_SetMaxRetries(t *testing.T) {
	tests := []struct {
		name       string
		maxRetries int
	}{
		{
			name:       "zero max retries",
			maxRetries: 0,
		},
		{
			name:       "success",
			maxRetries: 3,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			SetMaxRetries(tt.maxRetries)
			if DefaultRetrier.MaxRetries != tt.maxRetries {
				t.Errorf("DefaultRetrier.SetMaxRetries() maxRetries = %v, want %v", DefaultRetrier.MaxRetries, tt.maxRetries)
			}
		})
	}
}

func TestRetrier_DefaultBackoff(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		retries int
		want    time.Duration
	}{
		{
			name:    "zero retries",
			retries: 0,
			want:    1 * time.Second,
		},
		{
			name:    "success",
			retries: 3,
			want:    8 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := DefaultBackoff(tt.retries)
			if got != tt.want {
				t.Errorf("DefaultBackoff() = %v, want %v", got, tt.want)
			}
		})
	}
}
