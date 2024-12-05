package dependency

import (
	"fmt"
	"iter"
	"reflect"
	"testing"
	"time"
)

func TestNewContainer(t *testing.T) {
	c := NewContainer()
	if c.providers == nil {
		t.Error("expected providers to be initialized")
	}
	if c.namedProviders == nil {
		t.Error("expected namedProviders to be initialized")
	}
}

func TestContainer_Provide(t *testing.T) {
	tests := []struct {
		name          string
		providers     []Provider
		providerCount int
		namedCount    int
	}{
		{
			name: "single provider",
			providers: []Provider{
				NewSingleton("test"),
			},
			providerCount: 1,
		},
		{
			name: "multiple providers",
			providers: []Provider{
				NewFactory(func() string { return "test" }),
				NewSingleton("test"),
				NewSingleton[fmt.Stringer](time.Now()),
				NewSingletonFunc(time.Now),
			},
			providerCount: 3,
		},
		{
			name: "named providers",
			providers: []Provider{
				NewSingleton("test").Named("singleton-1"),
				NewSingleton("test").Named("singleton-2"),
			},
			providerCount: 1,
			namedCount:    2,
		},
		{
			name: "named and unnamed providers",
			providers: []Provider{
				NewSingleton("test"),
				NewSingleton("test").Named("singleton-1"),
			},
			providerCount: 1,
			namedCount:    1,
		},
		{
			name:          "no providers",
			providerCount: 0,
			namedCount:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewContainer()
			c.Provide(tt.providers...)

			if len(c.providers) != tt.providerCount {
				t.Errorf("%d providers found, want %d", len(c.providers), tt.providerCount)
			}
			if len(c.namedProviders) != tt.namedCount {
				t.Errorf("%d named providers found, want %d", len(c.namedProviders), tt.namedCount)
			}
		})
	}
}

func TestContainer_Resolve(t *testing.T) {
	tests := []struct {
		name      string
		providers []Provider
		want      any
	}{
		{
			name: "single provider",
			providers: []Provider{
				NewSingleton("test"),
			},
			want: "test",
		},
		{
			name: "multiple providers",
			providers: []Provider{
				NewFactory(func() string { return "test-factory" }),
				NewSingleton("test-singleton"),
				NewSingleton[fmt.Stringer](time.Now()),
				NewSingletonFunc(time.Now),
			},
			want: "test-factory",
		},
		{
			name: "no providers",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewContainer()
			c.Provide(tt.providers...)

			got := c.Resolve(reflect.TypeOf(tt.want))
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainer_ResolveNamed(t *testing.T) {
	tests := []struct {
		name      string
		providers []Provider
		want      any
	}{
		{
			name: "single provider",
			providers: []Provider{
				NewSingleton("test").Named("singleton-1"),
			},
			want: "test",
		},
		{
			name: "multiple providers",
			providers: []Provider{
				NewSingleton("test").Named("singleton-1"),
				NewSingleton("test2").Named("singleton-2"),
			},
			want: "test",
		},
		{
			name: "no providers",
			want: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewContainer()
			c.Provide(tt.providers...)

			got := c.ResolveNamed("singleton-1")
			if got != tt.want {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainer_ResolveAll(t *testing.T) {
	tests := []struct {
		name      string
		providers []Provider
		want      []any
	}{
		{
			name: "single provider",
			providers: []Provider{
				NewSingleton("test"),
			},
			want: []any{"test"},
		},
		{
			name: "multiple providers",
			providers: []Provider{
				NewSingleton("test"),
				NewSingleton("test2"),
			},
			want: []any{"test", "test2"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewContainer()
			c.Provide(tt.providers...)

			var got []any
			c.ResolveAll(reflect.TypeOf(tt.want[0]))(func(_ reflect.Type, v any) bool {
				got = append(got, v)
				return true
			})

			if len(got) != len(tt.want) {
				t.Errorf("got %v, want %v", got, tt.want)
			}
		})
	}
}

func TestContainer_ResolveAll_SpecificType(t *testing.T) {
	tests := []struct {
		name      string
		providers []Provider
		want      any
		wantType  reflect.Type
	}{
		{
			name: "single provider",
			providers: []Provider{
				NewSingleton("test"),
			},
			want:     "test",
			wantType: reflect.TypeFor[string](),
		},
		{
			name: "multiple providers",
			providers: []Provider{
				NewSingleton("test"),
				NewSingleton[fmt.Stringer](reflect.ValueOf("test")),
			},
			want:     reflect.ValueOf("test"),
			wantType: reflect.TypeFor[fmt.Stringer](),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := NewContainer()
			c.Provide(tt.providers...)

			next, stop := iter.Pull2(c.ResolveAll(tt.wantType))
			defer stop()

			for {
				typ, v, ok := next()
				if !ok {
					break
				}
				log, args := "Next iterator item = %v (%T)", []any{v, v}
				if typ != reflect.TypeOf(tt.want) {
					log, args = "Next iterator item = %v (%T implements %s)", []any{v, v, typ}
				}
				t.Logf(log, args...)

				switch v := v.(type) {
				case fmt.Stringer:
					if v.String() == tt.want.(fmt.Stringer).String() {
						return
					}
				default:
					if v == tt.want {
						return
					}
				}
			}

			logs, args := "No iterator item found with type %s", []any{tt.wantType}
			if tt.wantType == reflect.TypeOf(tt.want) {
				logs, args = "No iterator item found with value %v (%T)", []any{tt.want, tt.want}
			}
			t.Errorf(logs, args...)
		})
	}
}
