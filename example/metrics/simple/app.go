package main

import (
	"context"
	"fmt"

	"github.com/lvlcn-t/go-kit/metrics"
	"github.com/prometheus/client_golang/prometheus"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/codes"
)

func main() {
	ctx := context.Background()

	// Initialize the metrics client with the given exporter, URL, token, and certificate path
	client := metrics.New(metrics.Config{
		Exporter: metrics.STDOUT,
		CertPath: "",
	})

	// Initialize the open telemetry tracer with the given service name and version
	err := client.Initialize(ctx, "my-service-name", "v0.1.0")
	if err != nil {
		fmt.Println("failed to initialize metrics:", err)
		return
	}
	defer func() {
		if err := client.Shutdown(ctx); err != nil {
			fmt.Println("failed to shutdown metrics:", err)
		}
	}()

	// Register some prometheus collectors to the registry
	registry := client.GetRegistry()
	registry.MustRegister(prometheus.NewGaugeVec(prometheus.GaugeOpts{
		Name: "my_gauge",
		Help: "My gauge help",
	}, []string{"label"}))

	logTraceEvent(ctx)
}

// logTraceEvent logs a trace event using the OpenTelemetry tracer
func logTraceEvent(ctx context.Context) {
	tp := otel.GetTracerProvider()
	tracer := tp.Tracer("my-service-name")

	_, span := tracer.Start(ctx, "my-span")
	defer span.End()

	span.AddEvent("my-event")
	span.SetStatus(codes.Error, "my-error")
	span.RecordError(fmt.Errorf("my-error"))
}
