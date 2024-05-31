package metrics

import (
	"context"
	"errors"
	"reflect"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
)

func TestCollector_GetRegistry(t *testing.T) {
	tests := []struct {
		name     string
		registry *prometheus.Registry
		want     *prometheus.Registry
	}{
		{
			name:     "simple registry",
			registry: prometheus.NewRegistry(),
			want:     prometheus.NewRegistry(),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &manager{registry: tt.registry}

			if got := m.GetRegistry(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("PrometheusMetrics.GetRegistry() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	testMetrics := New(Config{})
	testGauge := prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "TEST_GAUGE",
		},
	)

	t.Run("Register a collector", func(t *testing.T) {
		testMetrics.registry.MustRegister(
			testGauge,
		)
	})
}

func TestTracer_Initialize_Shutdown(t *testing.T) {
	tests := []struct {
		name        string
		config      Config
		initErr     bool
		shutdownErr bool
		wantErr     error
	}{
		{
			name: "success - stdout exporter",
			config: Config{
				Exporter: STDOUT,
				Url:      "",
				Token:    "",
			},
			initErr:     false,
			shutdownErr: false,
		},
		{
			name: "success - otlp exporter",
			config: Config{
				Exporter: HTTP,
				Url:      "http://localhost:4317",
				Token:    "",
			},
			initErr:     false,
			shutdownErr: false,
		},
		{
			name: "success - otlp exporter with token",
			config: Config{
				Exporter: GRPC,
				Url:      "http://localhost:4317",
				Token:    "my-super-secret-token",
			},
			initErr:     false,
			shutdownErr: false,
		},
		{
			name: "success - no exporter",
			config: Config{
				Exporter: NOOP,
				Url:      "",
				Token:    "",
			},
			initErr:     false,
			shutdownErr: false,
		},
		{
			name: "failure - unsupported exporter",
			config: Config{
				Exporter: "unsupported",
				Url:      "",
				Token:    "",
			},
			initErr:     true,
			shutdownErr: false,
			wantErr:     nil,
		},
		{
			name: "failure - already initialized",
			config: Config{
				Exporter: STDOUT,
				Url:      "",
				Token:    "",
			},
			initErr:     true,
			shutdownErr: false,
			wantErr:     &ErrAlreadyInitialized{name: "name", version: "v0.1.0"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := New(tt.config)
			if _, ok := tt.wantErr.(*ErrAlreadyInitialized); ok {
				m.tracer = sdktrace.NewTracerProvider()
			}

			err := m.Initialize(context.Background(), "name", "v0.1.0")
			if (err != nil) != tt.initErr {
				t.Errorf("Tracer.Initialize() error = %v", err)
			}
			defer func() {
				if err = m.Shutdown(context.Background()); (err != nil) != tt.shutdownErr {
					t.Errorf("Tracer.Shutdown() error = %v", err)
				}
			}()

			if tt.initErr {
				if tt.wantErr != nil {
					if err == nil || !errors.Is(err, tt.wantErr) {
						t.Errorf("Tracer.Initialize() error = %v, want = %v", nil, tt.wantErr)
					}

					if err.Error() != tt.wantErr.Error() {
						t.Errorf("Tracer.Initialize() error = %v, want = %v", err, tt.wantErr)
					}
				}
				return
			}

			tp, ok := otel.GetTracerProvider().(*sdktrace.TracerProvider)
			if !ok {
				t.Errorf("Tracer.Initialize() type = %T, want = %T", tp, &sdktrace.TracerProvider{})
			}

			if tp == nil || m.tracer == nil {
				t.Errorf("Tracer.Initialize() tracer = %v, want = %v", nil, &sdktrace.TracerProvider{})
			}
		})
	}
}
