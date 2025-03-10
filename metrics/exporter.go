package metrics

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"

	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"google.golang.org/grpc/credentials"
)

// Config holds the configuration for OpenTelemetry.
type Config struct {
	// Exporter is the otlp exporter used to export the traces.
	Exporter Exporter `yaml:"exporter" mapstructure:"exporter"`
	// Url is the Url of the collector to which the traces are exported.
	Url string `yaml:"url" mapstructure:"url"`
	// Token is the token used to authenticate with the collector.
	Token string `yaml:"token" mapstructure:"token"`
	// TLS is the tls configuration for the exporter.
	TLS TLSConfig `yaml:"tls" mapstructure:"tls"`
}

// TLSConfig holds the configuration for the tls.
type TLSConfig struct {
	// Enabled is a flag to enable or disable the tls configuration.
	Enabled bool `yaml:"enabled" mapstructure:"enabled"`
	// CertPath is the path to a custom tls certificate file.
	// If not set, the system's root certificates are used.
	CertPath string `yaml:"certPath" mapstructure:"certPath"`
}

// IsEmpty returns true if the configuration is empty
func (c Config) IsEmpty() bool {
	return c == (Config{})
}

// Validate validates the configuration.
func (c *Config) Validate() error {
	err := c.Exporter.Validate()
	if c.Exporter.isExporting() && c.Url == "" {
		err = errors.Join(err, fmt.Errorf("url is required for otlp exporter %q", c.Exporter))
	}
	return err
}

// Exporter is the protocol used to export the traces
type Exporter string

const (
	// HTTP is the protocol used to export the traces via HTTP/1.1
	HTTP Exporter = "http"
	// GRPC is the protocol used to export the traces via HTTP/2 (gRPC)
	GRPC Exporter = "grpc"
	// STDOUT is the protocol used to export the traces to the standard output
	STDOUT Exporter = "stdout"
	// NOOP is the protocol used to not export the traces
	NOOP Exporter = ""
)

// String returns the string representation of the protocol
func (e Exporter) String() string {
	return string(e)
}

// Validate validates the protocol
func (e Exporter) Validate() error {
	if _, ok := registry[e]; !ok {
		return fmt.Errorf("unsupported exporter type: %s", e.String())
	}
	return nil
}

// isExporting returns true if the protocol is exporting the traces
func (e Exporter) isExporting() bool {
	return e == HTTP || e == GRPC
}

// exporterFactory is a function that creates a new exporter
type exporterFactory func(ctx context.Context, config *Config) (sdktrace.SpanExporter, error)

// registry contains the mapping of the exporter to the factory function
var registry = map[Exporter]exporterFactory{
	HTTP:   newHTTPExporter,
	GRPC:   newGRPCExporter,
	STDOUT: newStdoutExporter,
	NOOP:   newNoopExporter,
}

// Create creates a new exporter based on the configuration
func (e Exporter) Create(ctx context.Context, config *Config) (sdktrace.SpanExporter, error) {
	if factory, ok := registry[e]; ok {
		return factory(ctx, config)
	}
	return nil, fmt.Errorf("unsupported exporter type: %s", config.Exporter.String())
}

// RegisterExporter registers the exporter with the factory function.
// Not safe for concurrent use.
func (e Exporter) RegisterExporter(factory exporterFactory) error {
	if _, ok := registry[e]; ok {
		return fmt.Errorf("exporter %q is already registered", e.String())
	}
	registry[e] = factory
	return nil
}

// newHTTPExporter creates a new HTTP exporter
func newHTTPExporter(ctx context.Context, config *Config) (sdktrace.SpanExporter, error) {
	conf, err := newExporterConfig(config)
	if err != nil {
		return nil, err
	}

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(config.Url),
		otlptracehttp.WithHeaders(conf.headers),
	}

	if !config.TLS.Enabled {
		opts = append(opts, otlptracehttp.WithInsecure())
		return otlptracehttp.New(ctx, opts...)
	}
	if conf.tls != nil {
		opts = append(opts, otlptracehttp.WithTLSClientConfig(conf.tls))
	}

	return otlptracehttp.New(ctx, opts...)
}

// newGRPCExporter creates a new gRPC exporter
func newGRPCExporter(ctx context.Context, config *Config) (sdktrace.SpanExporter, error) {
	conf, err := newExporterConfig(config)
	if err != nil {
		return nil, err
	}

	opts := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(config.Url),
		otlptracegrpc.WithHeaders(conf.headers),
	}

	if !config.TLS.Enabled {
		opts = append(opts, otlptracegrpc.WithInsecure())
		return otlptracegrpc.New(ctx, opts...)
	}
	if conf.tls != nil {
		opts = append(opts, otlptracegrpc.WithTLSCredentials(credentials.NewTLS(conf.tls)))
	}

	return otlptracegrpc.New(ctx, opts...)
}

// newStdoutExporter creates a new stdout exporter
func newStdoutExporter(_ context.Context, _ *Config) (sdktrace.SpanExporter, error) {
	return stdouttrace.New(stdouttrace.WithPrettyPrint())
}

// newNoopExporter creates a new noop exporter
func newNoopExporter(_ context.Context, _ *Config) (sdktrace.SpanExporter, error) {
	return nil, nil
}

// exporterConfig holds the common configuration for the exporters.
type exporterConfig struct {
	// headers is the map of headers to be sent with the request.
	headers map[string]string
	// tls is the tls configuration for the exporter.
	tls *tls.Config
}

// newExporterConfig returns the common configuration for the exporters
func newExporterConfig(config *Config) (exporterConfig, error) {
	headers := map[string]string{}
	if config.Token != "" {
		headers["Authorization"] = fmt.Sprintf("Bearer %s", config.Token)
	}

	tlsCfg, err := getTLSConfig(config.TLS.CertPath)
	if err != nil {
		return exporterConfig{}, fmt.Errorf("failed to create TLS configuration: %w", err)
	}

	return exporterConfig{headers: headers, tls: tlsCfg}, nil
}

// fileOpener is the function used to open a file
type fileOpener func(string) (fs.File, error)

// openFile is the function used to open a file
var openFile fileOpener = func(name string) (fs.File, error) {
	return os.Open(name) // #nosec G304 // How else to open the file?
}

func getTLSConfig(certFile string) (conf *tls.Config, err error) {
	if certFile == "" {
		return nil, nil
	}

	file, err := openFile(certFile)
	if err != nil {
		return nil, fmt.Errorf("failed to open certificate file: %w", err)
	}
	defer func() {
		if cErr := file.Close(); cErr != nil {
			err = errors.Join(err, cErr)
		}
	}()

	b, err := io.ReadAll(file)
	if err != nil {
		return nil, fmt.Errorf("failed to read certificate file: %w", err)
	}

	pool := x509.NewCertPool()
	if !pool.AppendCertsFromPEM(b) {
		return nil, fmt.Errorf("failed to append certificate(s) from file: %s", certFile)
	}

	return &tls.Config{
		RootCAs:            pool,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
	}, nil
}
