// Package telemetry provides OpenTelemetry instrumentation for the application.
package telemetry

import (
	"os"
	"strings"
)

// Config holds the telemetry configuration
type Config struct {
	// ServiceName is the name of the service (defaults to "document-smbrelay-service")
	ServiceName string
	// ServiceVersion is the version of the service
	ServiceVersion string
	// Enabled determines if telemetry is enabled
	Enabled bool
	// TracingEnabled determines if tracing is enabled
	TracingEnabled bool
	// MetricsEnabled determines if metrics are enabled
	MetricsEnabled bool
	// OTLPEndpoint is the OTLP endpoint for traces and metrics
	OTLPEndpoint string
	// OTLPHeaders are additional headers to send with OTLP requests (e.g., for authentication)
	OTLPHeaders map[string]string
	// AzureAppInsightsConnectionString is the Application Insights connection string
	AzureAppInsightsConnectionString string
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

	// If Application Insights is configured, enable telemetry and set OTLP endpoint
	if appInsightsConnStr != "" {
		enabled = true
		// Azure Monitor ingests OpenTelemetry data via OTLP
		if otlpEndpoint == "" {
			// Extract ingestion endpoint from connection string
			otlpEndpoint = extractIngestionEndpoint(appInsightsConnStr)
		}
	}

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

// extractIngestionEndpoint extracts the ingestion endpoint from Application Insights connection string
func extractIngestionEndpoint(connStr string) string {
	// Connection string format: InstrumentationKey=xxx;IngestionEndpoint=https://xxx
	parts := strings.Split(connStr, ";")
	for _, part := range parts {
		if strings.HasPrefix(part, "IngestionEndpoint=") {
			endpoint := strings.TrimPrefix(part, "IngestionEndpoint=")
			// Azure Monitor OTLP endpoint is at /v1/traces and /v1/metrics
			return strings.TrimSuffix(endpoint, "/")
		}
	}
	// Default to global ingestion endpoint if not specified
	return "https://dc.services.visualstudio.com"
}
