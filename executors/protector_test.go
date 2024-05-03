package executors

import (
	"context"
	"errors"
	"testing"
)

func TestProtector(t *testing.T) {
	tests := []struct {
		name     string
		effector Effector
		wantErr  bool
	}{
		{
			name:     "nil effector",
			effector: nil,
			wantErr:  false,
		},
		{
			name:     "success without any errors",
			effector: noopEffector,
			wantErr:  false,
		},
		{
			name:     "panic",
			effector: func(_ context.Context) error { panic("test") },
			wantErr:  true,
		},
		{
			name:     "error",
			effector: func(_ context.Context) error { return errors.New("test") },
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := Protector(tt.effector)(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Protector() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
