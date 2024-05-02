package executors

import (
	"context"
	"errors"
	"testing"
	"time"
)

func TestCircuitBreaker(t *testing.T) {
	tests := []struct {
		name         string
		maxFailures  int
		resetTimeout time.Duration
		effector     Effector
		wantErr      bool
	}{
		{
			name:         "circuit open",
			maxFailures:  2,
			resetTimeout: 10 * time.Millisecond,
			effector: func(ctx context.Context) error {
				return errors.New("error")
			},
			wantErr: true,
		},
		{
			name:         "circuit closed",
			maxFailures:  2,
			resetTimeout: 10 * time.Millisecond,
			effector:     noopEffector,
			wantErr:      false,
		},
		{
			name:         "circuit open then closed",
			maxFailures:  2,
			resetTimeout: 10 * time.Millisecond,
			effector: func(ctx context.Context) error {
				return errors.New("error")
			},
			wantErr: true,
		},
		{
			name:     "nil effector",
			effector: nil,
			wantErr:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effector := CircuitBreaker(tt.maxFailures, tt.resetTimeout, tt.effector)

			for i := 0; i < tt.maxFailures; i++ {
				if err := effector(context.Background()); err != nil {
					if !tt.wantErr {
						t.Errorf("CircuitBreaker() error = %v, wantErr %v", err, tt.wantErr)
					}
				}
			}

			if err := effector(context.Background()); (err != nil) != tt.wantErr {
				t.Errorf("CircuitBreaker() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
