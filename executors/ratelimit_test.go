package executors

import (
	"context"
	"testing"

	"golang.org/x/time/rate"
)

func TestRateLimiter(t *testing.T) {
	tests := []struct {
		name     string
		rate     rate.Limit
		effector Effector
		wantErr  bool
	}{
		{
			name:     "nil effector",
			rate:     1,
			effector: nil,
			wantErr:  false,
		},
		{
			name:     "success",
			rate:     1,
			effector: noopEffector,
			wantErr:  false,
		},
		{
			name: "rate limit with negative rate",
			rate: -1,
			effector: func(ctx context.Context) error {
				return nil
			},
			wantErr: true,
		},
		{
			name: "rate limit with zero rate",
			rate: 0,
			effector: func(ctx context.Context) error {
				return nil
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			effector := RateLimiter(tt.rate, tt.effector)

			err := effector(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("RateLimiter() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
