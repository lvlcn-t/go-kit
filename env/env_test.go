package env_test

import (
	"errors"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/lvlcn-t/go-kit/env"
)

func TestGetWithFallback(t *testing.T) {
	tests := []struct {
		name         string
		key          string
		envValue     string
		defaultValue any
		fun          func(string, any) any
		want         any
		invalidType  bool
	}{
		{
			name:         "String variable set",
			key:          "TEST_STRING",
			envValue:     "hello",
			defaultValue: "default",
			fun: func(s string, a any) any {
				return env.GetWithFallback(s, a.(string))
			},
			want: "hello",
		},
		{
			name:         "String variable not set",
			key:          "TEST_STRING_NOT_SET",
			defaultValue: "default",
			fun: func(s string, a any) any {
				return env.GetWithFallback(s, a.(string))
			},
			want: "default",
		},
		{
			name:         "Int variable set",
			key:          "TEST_INT",
			envValue:     "42",
			defaultValue: 0,
			fun: func(s string, a any) any {
				return env.GetWithFallback(s, a.(int))
			},
			want: 42,
		},
		{
			name:         "Int variable invalid, returns default",
			key:          "TEST_INT_INVALID",
			envValue:     "invalid",
			defaultValue: 10,
			fun: func(s string, a any) any {
				return env.GetWithFallback(s, a.(int))
			},
			want: 10,
		},
		{
			name:         "Bool variable true",
			key:          "TEST_BOOL_TRUE",
			envValue:     "true",
			defaultValue: false,
			fun: func(s string, a any) any {
				return env.GetWithFallback(s, a.(bool))
			},
			want: true,
		},
		{
			name:         "Bool variable invalid, returns default",
			key:          "TEST_BOOL_INVALID",
			envValue:     "notabool",
			defaultValue: false,
			fun: func(s string, a any) any {
				return env.GetWithFallback(s, a.(bool))
			},
			want: false,
		},
		{
			name:         "Duration variable with custom converter",
			key:          "TEST_DURATION",
			envValue:     "5s",
			defaultValue: 0 * time.Second,
			fun: func(s string, a any) any {
				return env.GetWithFallback(s, a.(time.Duration), time.ParseDuration)
			},
			want: 5 * time.Second,
		},
		{
			name:         "Interface variable with custom converter",
			key:          "TEST_INTERFACE",
			envValue:     "42",
			defaultValue: 0,
			fun: func(s string, a any) any {
				return env.GetWithFallback(s, a, func(s string) (any, error) {
					return strconv.Atoi(s)
				})
			},
			want:        42,
			invalidType: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.invalidType {
				defer func() {
					if r := recover(); r == nil {
						t.Errorf("env.GetWithFallback() did not panic for invalid type")
					}
				}()
			}

			if tt.envValue != "" {
				t.Setenv(tt.key, tt.envValue)
			} else {
				os.Unsetenv(tt.key)
			}

			got := tt.fun(tt.key, tt.defaultValue)
			if got != tt.want {
				t.Errorf("Got %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func TestGet_WithFallback(t *testing.T) {
	t.Setenv("TEST_VAR", "123")

	value, err := env.Get[int]("TEST_VAR").WithFallback(0).Value()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if value != 123 {
		t.Errorf("Expected 123, got %d", value)
	}
}

func TestGet_NoFallback(t *testing.T) {
	_, err := env.Get[int]("TEST_VAR").NoFallback().Value()
	if err == nil {
		t.Errorf("Expected error for missing required variable")
	}
}

func TestGet_OrDie(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("Expected panic for missing required variable")
		}
	}()

	_, _ = env.Get[int]("TEST_VAR").OrDie().Value()
}

func TestGet_CustomConverter(t *testing.T) {
	t.Setenv("TEST_VAR", "custom")

	converter := func(s string) (string, error) {
		if s == "custom" {
			return "converted", nil
		}
		return "", errors.New("conversion failed")
	}

	value, err := env.Get[string]("TEST_VAR").WithFallback("default").Convert(converter).Value()
	if err != nil {
		t.Errorf("Unexpected error: %v", err)
	}
	if value != "converted" {
		t.Errorf("Expected 'converted', got %s", value)
	}
}
