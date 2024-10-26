package env_test

import (
	"errors"
	"os"
	"reflect"
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
		{
			name:         "Channel variable with custom converter",
			key:          "TEST_CHANNEL",
			envValue:     "42",
			defaultValue: 0,
			fun: func(s string, a any) any {
				return env.GetWithFallback(s, func(s string) (chan any, error) {
					return make(chan any, 1), nil
				})
			},
			invalidType: true,
		},
		{
			name:         "Pointer string variable set",
			key:          "TEST_POINTER",
			envValue:     "my-var",
			defaultValue: toPtr("default"),
			fun: func(s string, a any) any {
				return env.GetWithFallback(s, a.(*string))
			},
			want:        toPtr("my-var"),
			invalidType: false,
		},
		{
			name:         "Pointer int variable set",
			key:          "TEST_POINTER_NOT_SET",
			envValue:     "42",
			defaultValue: toPtr(50),
			fun: func(s string, a any) any {
				return env.GetWithFallback(s, a.(*int))
			},
			want: toPtr(42),
		},
		{
			name:         "Pointer uint variable set",
			key:          "TEST_POINTER_UINT",
			envValue:     "42",
			defaultValue: toPtr(uint(50)),
			fun: func(s string, a any) any {
				return env.GetWithFallback(s, a.(*uint))
			},
			want: toPtr(uint(42)),
		},
		{
			name:         "Pointer bool variable set",
			key:          "TEST_POINTER_BOOL",
			envValue:     "true",
			defaultValue: toPtr(false),
			fun: func(s string, a any) any {
				return env.GetWithFallback(s, a.(*bool))
			},
			want: toPtr(true),
		},
		{
			name:         "Pointer float64 variable set",
			key:          "TEST_POINTER_FLOAT64",
			envValue:     "42.42",
			defaultValue: toPtr(50.50),
			fun: func(s string, a any) any {
				return env.GetWithFallback(s, a.(*float64))
			},
			want: toPtr(42.42),
		},
		{
			name:         "Pointer complex128 variable set",
			key:          "TEST_POINTER_COMPLEX128",
			envValue:     "42.42+42.42i",
			defaultValue: toPtr(50.50 + 50.50i),
			fun: func(s string, a any) any {
				return env.GetWithFallback(s, a.(*complex128))
			},
			want: toPtr(42.42 + 42.42i),
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
			if reflect.TypeOf(got).Kind() == reflect.Pointer {
				got = reflect.ValueOf(got).Elem().Interface()
			}
			if reflect.TypeOf(tt.want).Kind() == reflect.Pointer {
				tt.want = reflect.ValueOf(tt.want).Elem().Interface()
			}

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Got %v (%T), want %v (%T)", got, got, tt.want, tt.want)
			}
		})
	}
}

func toPtr[T any](v T) *T {
	return &v
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
