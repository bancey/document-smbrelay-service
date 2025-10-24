// Package telemetry provides OpenTelemetry instrumentation for the application.
package telemetry

import (
	"context"
	"fmt"
	"strings"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutmetric"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"

	"github.com/bancey/document-smbrelay-service/internal/logger"
)

// Provider holds the telemetry providers
type Provider struct {
	tracerProvider *sdktrace.TracerProvider
	meterProvider  *sdkmetric.MeterProvider
	config         *Config
}

// Initialize sets up OpenTelemetry with the provided configuration
func Initialize(ctx context.Context, cfg *Config) (*Provider, error) {
	if !cfg.Enabled {
		logger.Info("OpenTelemetry is disabled")
		return &Provider{config: cfg}, nil
	}

	// Check for Azure Application Insights without OTLP endpoint configuration
	if cfg.AzureAppInsightsConnectionString != "" && cfg.OTLPEndpoint == "" {
		logger.Warn("⚠️  Azure Application Insights connection string detected but OTEL_EXPORTER_OTLP_ENDPOINT is not set")
		logger.Warn("⚠️  Azure Application Insights does NOT support direct OTLP HTTP ingestion")
		logger.Warn("⚠️  Please deploy an OpenTelemetry Collector and set OTEL_EXPORTER_OTLP_ENDPOINT")
		logger.Warn("⚠️  See docs/OPENTELEMETRY.md for setup instructions")
		return nil, fmt.Errorf(
			"azure Application Insights requires OpenTelemetry Collector - set OTEL_EXPORTER_OTLP_ENDPOINT",
		)
	}

	logger.Info("Initializing OpenTelemetry instrumentation")
	logger.Info("Service: %s, Version: %s", cfg.ServiceName, cfg.ServiceVersion)

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceNameKey.String(cfg.ServiceName),
			semconv.ServiceVersionKey.String(cfg.ServiceVersion),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	provider := &Provider{config: cfg}

	// Initialize tracing
	if cfg.TracingEnabled {
		tracerProvider, err := initTracing(ctx, cfg, res)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize tracing: %w", err)
		}
		provider.tracerProvider = tracerProvider
		otel.SetTracerProvider(tracerProvider)
		logger.Info("Tracing initialized successfully")
	}

	// Initialize metrics
	if cfg.MetricsEnabled {
		meterProvider, err := initMetrics(ctx, cfg, res)
		if err != nil {
			return nil, fmt.Errorf("failed to initialize metrics: %w", err)
		}
		provider.meterProvider = meterProvider
		otel.SetMeterProvider(meterProvider)
		logger.Info("Metrics initialized successfully")
	}

	// Set up propagators for distributed tracing
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	logger.Info("OpenTelemetry initialization complete")
	return provider, nil
}

// initTracing initializes the tracing provider
func initTracing(ctx context.Context, cfg *Config, res *resource.Resource) (*sdktrace.TracerProvider, error) {
	var exporter sdktrace.SpanExporter
	var err error

	if cfg.OTLPEndpoint != "" {
		// Use OTLP exporter (for Azure Application Insights or other OTLP backends)
		logger.Info("Configuring OTLP trace exporter: %s", cfg.OTLPEndpoint)
		exporter, err = createOTLPTraceExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}
	} else {
		// Use stdout exporter for development
		logger.Info("Configuring stdout trace exporter (development mode)")
		exporter, err = stdouttrace.New(stdouttrace.WithPrettyPrint())
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout trace exporter: %w", err)
		}
	}

	// Create tracer provider with batch span processor
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter,
			sdktrace.WithBatchTimeout(5*time.Second),
		),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
	)

	return tp, nil
}

// initMetrics initializes the metrics provider
func initMetrics(ctx context.Context, cfg *Config, res *resource.Resource) (*sdkmetric.MeterProvider, error) {
	var exporter sdkmetric.Exporter
	var err error

	if cfg.OTLPEndpoint != "" {
		// Use OTLP exporter
		logger.Info("Configuring OTLP metric exporter: %s", cfg.OTLPEndpoint)
		exporter, err = createOTLPMetricExporter(ctx, cfg)
		if err != nil {
			return nil, err
		}
	} else {
		// Use stdout exporter for development
		logger.Info("Configuring stdout metric exporter (development mode)")
		exporter, err = stdoutmetric.New()
		if err != nil {
			return nil, fmt.Errorf("failed to create stdout metric exporter: %w", err)
		}
	}

	// Create meter provider with periodic reader
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(sdkmetric.NewPeriodicReader(exporter,
			sdkmetric.WithInterval(60*time.Second),
		)),
		sdkmetric.WithResource(res),
	)

	return mp, nil
}

// Shutdown gracefully shuts down the telemetry providers
func (p *Provider) Shutdown(ctx context.Context) error {
	if !p.config.Enabled {
		return nil
	}

	logger.Info("Shutting down OpenTelemetry")

	if p.tracerProvider != nil {
		if err := p.tracerProvider.Shutdown(ctx); err != nil {
			logger.Error("Error shutting down tracer provider: %v", err)
			return err
		}
	}

	if p.meterProvider != nil {
		if err := p.meterProvider.Shutdown(ctx); err != nil {
			logger.Error("Error shutting down meter provider: %v", err)
			return err
		}
	}

	logger.Info("OpenTelemetry shutdown complete")
	return nil
}

// otlpConfig holds common OTLP configuration
type otlpConfig struct {
	headers      map[string]string
	endpoint     string
	exporterType string // "trace" or "metric"
	isInsecure   bool
}

// buildOTLPConfig creates common OTLP configuration from the telemetry config
func buildOTLPConfig(cfg *Config, exporterType string) *otlpConfig {
	endpoint := stripScheme(cfg.OTLPEndpoint)
	isInsecure := isLocalEndpoint(endpoint)

	// Log TLS configuration
	if !isInsecure {
		logger.Info("Using TLS for OTLP %s export", exporterType)
	} else {
		logger.Info("Using insecure HTTP for local OTLP %s export", exporterType)
	}

	return &otlpConfig{
		endpoint:     endpoint,
		isInsecure:   isInsecure,
		headers:      cfg.OTLPHeaders,
		exporterType: exporterType,
	}
}

// createOTLPTraceExporter creates an OTLP trace exporter with common configuration
//
//nolint:dupl // Unavoidable duplication due to different exporter types (trace vs metric)
func createOTLPTraceExporter(ctx context.Context, cfg *Config) (sdktrace.SpanExporter, error) {
	otlpCfg := buildOTLPConfig(cfg, "trace")

	opts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(otlpCfg.endpoint),
	}

	if otlpCfg.isInsecure {
		opts = append(opts, otlptracehttp.WithInsecure())
	}

	if len(otlpCfg.headers) > 0 {
		opts = append(opts, otlptracehttp.WithHeaders(otlpCfg.headers))
	}

	exporter, err := otlptracehttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
	}
	return exporter, nil
}

// createOTLPMetricExporter creates an OTLP metric exporter with common configuration
//
//nolint:dupl // Unavoidable duplication due to different exporter types (trace vs metric)
func createOTLPMetricExporter(ctx context.Context, cfg *Config) (sdkmetric.Exporter, error) {
	otlpCfg := buildOTLPConfig(cfg, "metric")

	opts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(otlpCfg.endpoint),
	}

	if otlpCfg.isInsecure {
		opts = append(opts, otlpmetrichttp.WithInsecure())
	}

	if len(otlpCfg.headers) > 0 {
		opts = append(opts, otlpmetrichttp.WithHeaders(otlpCfg.headers))
	}

	exporter, err := otlpmetrichttp.New(ctx, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
	}
	return exporter, nil
}

// stripScheme removes the http:// or https:// prefix from a URL
// The OTLP HTTP exporters expect only the host:port part
func stripScheme(endpoint string) string {
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	return endpoint
}

// isLocalEndpoint checks if an endpoint is a local development endpoint
// Returns true for localhost, 127.0.0.1, and 0.0.0.0 addresses
func isLocalEndpoint(endpoint string) bool {
	return strings.HasPrefix(endpoint, "localhost") ||
		strings.HasPrefix(endpoint, "127.0.0.1") ||
		strings.HasPrefix(endpoint, "0.0.0.0")
}
