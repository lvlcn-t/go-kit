package executors

import (
	"context"
	"testing"
	"time"
)

func TestTimeout(t *testing.T) {
	tests := []struct {
		name     string
		timeout  time.Duration
		effector Effector
		wantErr  bool
	}{
		{
			name:     "nil effector",
			timeout:  1 * time.Millisecond,
			effector: nil,
			wantErr:  false,
		},
		{
			name:     "success",
			timeout:  1 * time.Millisecond,
			effector: noopEffector,
			wantErr:  false,
		},
		{
			name:    "timeout",
			timeout: 1 * time.Millisecond,
			effector: func(ctx context.Context) error {
				select {
				case <-time.After(10 * time.Millisecond):
					return nil
				case <-ctx.Done():
					return ctx.Err()
				}
			},
			wantErr: true,
		},
		{
			name:    "timeout with default timeout",
			timeout: 0,
			effector: func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
			wantErr: true,
		},
		{
			name:    "timeout with negative timeout",
			timeout: -1 * time.Millisecond,
			effector: func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
			wantErr: true,
		},
		{
			name:    "timeout with zero timeout",
			timeout: 0,
			effector: func(ctx context.Context) error {
				time.Sleep(10 * time.Millisecond)
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effector := Timeouter(tt.timeout, tt.effector)

			err := effector(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Timeout() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
