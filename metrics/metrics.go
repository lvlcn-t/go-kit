// Package metrics provides a way to initialize OpenTelemetry metrics and Prometheus collectors.
package metrics

import (
	"context"
	"fmt"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

var (
	_ Tracer    = (*manager)(nil)
	_ Collector = (*manager)(nil)
)

// Tracer is the interface that wraps the basic methods of the OpenTelemetry tracer
type Tracer interface {
	// Initialize initializes the OpenTelemetry metrics with the given service name and version
	Initialize(ctx context.Context, serviceName, serviceVersion string) error
	// Shutdown shuts down the OpenTelemetry tracer
	Shutdown(ctx context.Context) error
}

// Collector is the interface that wraps the basic methods of the Prometheus collector
type Collector interface {
	// GetRegistry returns the prometheus registry instance containing the registered prometheus collectors
	GetRegistry() *prometheus.Registry
}

// manager implements the Tracer and Collector interfaces
// It is used to initialize the OpenTelemetry manager and Prometheus collectors
type manager struct {
	// config holds the configuration for the OpenTelemetry metrics
	config Config
	// registry holds the prometheus registry instance
	registry *prometheus.Registry
	// tracer holds the OpenTelemetry tracer tracer instance
	tracer *sdktrace.TracerProvider
}

// New creates a new metrics instance with the given configuration.
// This instance can be used to initialize the OpenTelemetry metrics and Prometheus collectors.
func New(config Config) *manager {
	registry := prometheus.NewRegistry()

	registry.MustRegister(
		collectors.NewGoCollector(),
		collectors.NewProcessCollector(collectors.ProcessCollectorOpts{}),
	)

	return &manager{
		config:   config,
		registry: registry,
	}
}

// GetRegistry returns the prometheus registry instance containing the registered prometheus collectors
func (m *manager) GetRegistry() *prometheus.Registry {
	return m.registry
}

// ErrAlreadyInitialized is an error that is returned when the metrics are already initialized.
// This error is returned when the Initialize method is called more than once.
type ErrAlreadyInitialized struct {
	name    string
	version string
}

// Error returns the error message.
func (e *ErrAlreadyInitialized) Error() string {
	return fmt.Sprintf("metrics already initialized for service %q version %q", e.name, e.version)
}

// Is reports whether the error is an ErrAlreadyInitialized.
func (e *ErrAlreadyInitialized) Is(target error) bool {
	_, ok := target.(*ErrAlreadyInitialized)
	return ok
}

// Initialize initializes the OpenTelemetry metrics with the given service name and version.
func (m *manager) Initialize(ctx context.Context, name, version string) error {
	if m.tracer != nil {
		return &ErrAlreadyInitialized{name: name, version: version}
	}

	res, err := resource.New(ctx,
		resource.WithHost(),
		resource.WithContainer(),
		resource.WithAttributes(
			semconv.ServiceNameKey.String(name),
			semconv.ServiceVersionKey.String(version),
		),
	)
	if err != nil {
		return fmt.Errorf("failed to create resource: %v", err)
	}

	exporter, err := m.config.Exporter.Create(ctx, &m.config)
	if err != nil {
		return fmt.Errorf("failed to create exporter: %v", err)
	}

	bsp := sdktrace.NewBatchSpanProcessor(exporter)
	tp := sdktrace.NewTracerProvider(
		// TODO: Keep track of the sampler if it runs into traffic issues due to the high volume of data.
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithSpanProcessor(bsp),
		sdktrace.WithResource(res),
	)
	otel.SetTracerProvider(tp)
	m.tracer = tp
	return nil
}

// Shutdown closes the connection to the OpenTelemetry metrics and Prometheus collectors.
func (m *manager) Shutdown(ctx context.Context) error {
	if m.tracer != nil {
		err := m.tracer.Shutdown(ctx)
		if err != nil {
			return fmt.Errorf("failed to shutdown tracer provider: %v", err)
		}
		m.tracer = nil
	}

	return nil
}
