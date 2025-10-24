// Package telemetry provides OpenTelemetry instrumentation for the application.
package telemetry

import (
	"os"
	"strings"
)

// Config holds the telemetry configuration
type Config struct {
	// OTLPHeaders are additional headers to send with OTLP requests (e.g., for authentication)
	OTLPHeaders map[string]string
	// ServiceName is the name of the service (defaults to "document-smbrelay-service")
	ServiceName string
	// ServiceVersion is the version of the service
	ServiceVersion string
	// OTLPEndpoint is the OTLP endpoint for traces and metrics
	OTLPEndpoint string
	// AzureAppInsightsConnectionString is the Application Insights connection string
	AzureAppInsightsConnectionString string
	// Enabled determines if telemetry is enabled
	Enabled bool
	// TracingEnabled determines if tracing is enabled
	TracingEnabled bool
	// MetricsEnabled determines if metrics are enabled
	MetricsEnabled bool
}

// LoadConfig loads telemetry configuration from environment variables
func LoadConfig() *Config {
	enabled := strings.ToLower(os.Getenv("OTEL_ENABLED")) == "true"
	tracingEnabled := strings.ToLower(os.Getenv("OTEL_TRACING_ENABLED")) != "false" // enabled by default
	metricsEnabled := strings.ToLower(os.Getenv("OTEL_METRICS_ENABLED")) != "false" // enabled by default

	serviceName := os.Getenv("OTEL_SERVICE_NAME")
	if serviceName == "" {
		serviceName = "document-smbrelay-service"
	}

	serviceVersion := os.Getenv("OTEL_SERVICE_VERSION")
	if serviceVersion == "" {
		serviceVersion = "1.0.0"
	}

	otlpEndpoint := os.Getenv("OTEL_EXPORTER_OTLP_ENDPOINT")

	// Parse OTLP headers from environment variable
	// Format: key1=value1,key2=value2
	headersStr := os.Getenv("OTEL_EXPORTER_OTLP_HEADERS")
	headers := make(map[string]string)
	if headersStr != "" {
		pairs := strings.Split(headersStr, ",")
		for _, pair := range pairs {
			kv := strings.SplitN(strings.TrimSpace(pair), "=", 2)
			if len(kv) == 2 {
				headers[strings.TrimSpace(kv[0])] = strings.TrimSpace(kv[1])
			}
		}
	}

	// Azure Application Insights connection string
	appInsightsConnStr := os.Getenv("APPLICATIONINSIGHTS_CONNECTION_STRING")

	// Azure Application Insights setup
	// Note: As of October 2025, Azure Application Insights does NOT support direct OTLP HTTP ingestion.
	// The recommended approach is to use an OpenTelemetry Collector as a bridge:
	//   1. Deploy OpenTelemetry Collector with Azure Monitor exporter
	//   2. Point your application to the collector's OTLP endpoint
	//   3. The collector forwards telemetry to Application Insights
	//
	// To use this service with Application Insights:
	//   - Set OTEL_ENABLED=true
	//   - Set OTEL_EXPORTER_OTLP_ENDPOINT to your OpenTelemetry Collector endpoint
	//   - Configure the collector with your APPLICATIONINSIGHTS_CONNECTION_STRING
	//
	// The connection string is stored for reference but does not enable direct export.
	// If appInsightsConnStr is set without otlpEndpoint, the Initialize function will return an error.

	return &Config{
		ServiceName:                      serviceName,
		ServiceVersion:                   serviceVersion,
		Enabled:                          enabled,
		TracingEnabled:                   tracingEnabled && enabled,
		MetricsEnabled:                   metricsEnabled && enabled,
		OTLPEndpoint:                     otlpEndpoint,
		OTLPHeaders:                      headers,
		AzureAppInsightsConnectionString: appInsightsConnStr,
	}
}
