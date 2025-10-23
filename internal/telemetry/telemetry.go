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

		// Strip scheme from endpoint - the OTLP HTTP exporter expects only host:port
		endpoint := stripScheme(cfg.OTLPEndpoint)

		opts := []otlptracehttp.Option{
			otlptracehttp.WithEndpoint(endpoint),
		}

		// Add headers if provided
		if len(cfg.OTLPHeaders) > 0 {
			opts = append(opts, otlptracehttp.WithHeaders(cfg.OTLPHeaders))
		}

		// If using Azure Application Insights connection string
		if cfg.AzureAppInsightsConnectionString != "" {
			logger.Info("Configuring for Azure Application Insights")
			// Add instrumentation key as header
			instrumentationKey := extractInstrumentationKey(cfg.AzureAppInsightsConnectionString)
			if instrumentationKey != "" {
				if cfg.OTLPHeaders == nil {
					cfg.OTLPHeaders = make(map[string]string)
				}
				cfg.OTLPHeaders["x-api-key"] = instrumentationKey
				opts = append(opts, otlptracehttp.WithHeaders(cfg.OTLPHeaders))
			}
		}

		exporter, err = otlptracehttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP trace exporter: %w", err)
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

		// Strip scheme from endpoint - the OTLP HTTP exporter expects only host:port
		endpoint := stripScheme(cfg.OTLPEndpoint)

		opts := []otlpmetrichttp.Option{
			otlpmetrichttp.WithEndpoint(endpoint),
		}

		// Add headers if provided
		if len(cfg.OTLPHeaders) > 0 {
			opts = append(opts, otlpmetrichttp.WithHeaders(cfg.OTLPHeaders))
		}

		// If using Azure Application Insights
		if cfg.AzureAppInsightsConnectionString != "" {
			instrumentationKey := extractInstrumentationKey(cfg.AzureAppInsightsConnectionString)
			if instrumentationKey != "" {
				if cfg.OTLPHeaders == nil {
					cfg.OTLPHeaders = make(map[string]string)
				}
				cfg.OTLPHeaders["x-api-key"] = instrumentationKey
				opts = append(opts, otlpmetrichttp.WithHeaders(cfg.OTLPHeaders))
			}
		}

		exporter, err = otlpmetrichttp.New(ctx, opts...)
		if err != nil {
			return nil, fmt.Errorf("failed to create OTLP metric exporter: %w", err)
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

// extractInstrumentationKey extracts the instrumentation key from Application Insights connection string
func extractInstrumentationKey(connStr string) string {
	// Connection string format: InstrumentationKey=xxx;IngestionEndpoint=https://xxx
	parts := strings.Split(connStr, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "InstrumentationKey=") {
			return strings.TrimPrefix(part, "InstrumentationKey=")
		}
	}
	return ""
}

// stripScheme removes the http:// or https:// prefix from a URL
// The OTLP HTTP exporters expect only the host:port part
func stripScheme(endpoint string) string {
	endpoint = strings.TrimPrefix(endpoint, "https://")
	endpoint = strings.TrimPrefix(endpoint, "http://")
	return endpoint
}
