package telemetry

import (
	"os"
	"testing"
)

func TestLoadConfig(t *testing.T) {
	tests := []struct {
		envVars  map[string]string
		expected *Config
		name     string
	}{
		{
			name:    "disabled by default",
			envVars: map[string]string{},
			expected: &Config{
				ServiceName:    "document-smbrelay-service",
				ServiceVersion: "1.0.0",
				Enabled:        false,
				TracingEnabled: false,
				MetricsEnabled: false,
			},
		},
		{
			name: "enabled explicitly",
			envVars: map[string]string{
				"OTEL_ENABLED": "true",
			},
			expected: &Config{
				ServiceName:    "document-smbrelay-service",
				ServiceVersion: "1.0.0",
				Enabled:        true,
				TracingEnabled: true,
				MetricsEnabled: true,
			},
		},
		{
			name: "custom service name and version",
			envVars: map[string]string{
				"OTEL_ENABLED":         "true",
				"OTEL_SERVICE_NAME":    "my-service",
				"OTEL_SERVICE_VERSION": "2.0.0",
			},
			expected: &Config{
				ServiceName:    "my-service",
				ServiceVersion: "2.0.0",
				Enabled:        true,
				TracingEnabled: true,
				MetricsEnabled: true,
			},
		},
		{
			name: "OTLP endpoint configured",
			envVars: map[string]string{
				"OTEL_ENABLED":                "true",
				"OTEL_EXPORTER_OTLP_ENDPOINT": "localhost:4318",
			},
			expected: &Config{
				ServiceName:    "document-smbrelay-service",
				ServiceVersion: "1.0.0",
				Enabled:        true,
				TracingEnabled: true,
				MetricsEnabled: true,
				OTLPEndpoint:   "localhost:4318",
			},
		},
		{
			name: "Application Insights connection string",
			envVars: map[string]string{
				"APPLICATIONINSIGHTS_CONNECTION_STRING": "InstrumentationKey=test-key;IngestionEndpoint=https://test.applicationinsights.azure.com",
			},
			expected: &Config{
				ServiceName:                      "document-smbrelay-service",
				ServiceVersion:                   "1.0.0",
				Enabled:                          true,
				TracingEnabled:                   true,
				MetricsEnabled:                   true,
				OTLPEndpoint:                     "https://test.applicationinsights.azure.com",
				AzureAppInsightsConnectionString: "InstrumentationKey=test-key;IngestionEndpoint=https://test.applicationinsights.azure.com",
			},
		},
		{
			name: "tracing disabled",
			envVars: map[string]string{
				"OTEL_ENABLED":         "true",
				"OTEL_TRACING_ENABLED": "false",
			},
			expected: &Config{
				ServiceName:    "document-smbrelay-service",
				ServiceVersion: "1.0.0",
				Enabled:        true,
				TracingEnabled: false,
				MetricsEnabled: true,
			},
		},
		{
			name: "metrics disabled",
			envVars: map[string]string{
				"OTEL_ENABLED":         "true",
				"OTEL_METRICS_ENABLED": "false",
			},
			expected: &Config{
				ServiceName:    "document-smbrelay-service",
				ServiceVersion: "1.0.0",
				Enabled:        true,
				TracingEnabled: true,
				MetricsEnabled: false,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Clear environment
			os.Clearenv()

			// Set environment variables
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			// Load config
			cfg := LoadConfig()

			// Check results
			if cfg.ServiceName != tt.expected.ServiceName {
				t.Errorf("ServiceName = %v, want %v", cfg.ServiceName, tt.expected.ServiceName)
			}
			if cfg.ServiceVersion != tt.expected.ServiceVersion {
				t.Errorf("ServiceVersion = %v, want %v", cfg.ServiceVersion, tt.expected.ServiceVersion)
			}
			if cfg.Enabled != tt.expected.Enabled {
				t.Errorf("Enabled = %v, want %v", cfg.Enabled, tt.expected.Enabled)
			}
			if cfg.TracingEnabled != tt.expected.TracingEnabled {
				t.Errorf("TracingEnabled = %v, want %v", cfg.TracingEnabled, tt.expected.TracingEnabled)
			}
			if cfg.MetricsEnabled != tt.expected.MetricsEnabled {
				t.Errorf("MetricsEnabled = %v, want %v", cfg.MetricsEnabled, tt.expected.MetricsEnabled)
			}
			if cfg.OTLPEndpoint != tt.expected.OTLPEndpoint {
				t.Errorf("OTLPEndpoint = %v, want %v", cfg.OTLPEndpoint, tt.expected.OTLPEndpoint)
			}
			if cfg.AzureAppInsightsConnectionString != tt.expected.AzureAppInsightsConnectionString {
				t.Errorf("AzureAppInsightsConnectionString = %v, want %v",
					cfg.AzureAppInsightsConnectionString, tt.expected.AzureAppInsightsConnectionString)
			}
		})
	}
}

func TestExtractIngestionEndpoint(t *testing.T) {
	tests := []struct {
		name     string
		connStr  string
		expected string
	}{
		{
			name:     "with ingestion endpoint",
			connStr:  "InstrumentationKey=test-key;IngestionEndpoint=https://test.applicationinsights.azure.com/",
			expected: "https://test.applicationinsights.azure.com",
		},
		{
			name:     "without ingestion endpoint",
			connStr:  "InstrumentationKey=test-key",
			expected: "https://dc.services.visualstudio.com",
		},
		{
			name:     "multiple parts",
			connStr:  "InstrumentationKey=test-key;LiveEndpoint=https://live.test.com;IngestionEndpoint=https://ingest.test.com",
			expected: "https://ingest.test.com",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractIngestionEndpoint(tt.connStr)
			if result != tt.expected {
				t.Errorf("extractIngestionEndpoint() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractInstrumentationKey(t *testing.T) {
	tests := []struct {
		name     string
		connStr  string
		expected string
	}{
		{
			name:     "with instrumentation key",
			connStr:  "InstrumentationKey=test-key-123;IngestionEndpoint=https://test.com",
			expected: "test-key-123",
		},
		{
			name:     "without instrumentation key",
			connStr:  "IngestionEndpoint=https://test.com",
			expected: "",
		},
		{
			name:     "only instrumentation key",
			connStr:  "InstrumentationKey=abc-def-ghi",
			expected: "abc-def-ghi",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := extractInstrumentationKey(tt.connStr)
			if result != tt.expected {
				t.Errorf("extractInstrumentationKey() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestLoadConfig_HeadersParsing(t *testing.T) {
	tests := []struct {
		expectedHeaders map[string]string
		name            string
		headersEnv      string
	}{
		{
			name:            "empty headers",
			headersEnv:      "",
			expectedHeaders: map[string]string{},
		},
		{
			name:       "single header",
			headersEnv: "key1=value1",
			expectedHeaders: map[string]string{
				"key1": "value1",
			},
		},
		{
			name:       "multiple headers",
			headersEnv: "key1=value1,key2=value2,key3=value3",
			expectedHeaders: map[string]string{
				"key1": "value1",
				"key2": "value2",
				"key3": "value3",
			},
		},
		{
			name:       "headers with spaces",
			headersEnv: "key1 = value1 , key2 = value2",
			expectedHeaders: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:       "malformed header without equals",
			headersEnv: "key1=value1,malformed,key2=value2",
			expectedHeaders: map[string]string{
				"key1": "value1",
				"key2": "value2",
			},
		},
		{
			name:       "header with equals in value",
			headersEnv: "key1=value=with=equals",
			expectedHeaders: map[string]string{
				"key1": "value=with=equals",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			if tt.headersEnv != "" {
				os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", tt.headersEnv)
			}

			cfg := LoadConfig()

			if len(cfg.OTLPHeaders) != len(tt.expectedHeaders) {
				t.Errorf("Expected %d headers, got %d", len(tt.expectedHeaders), len(cfg.OTLPHeaders))
			}

			for k, v := range tt.expectedHeaders {
				if cfg.OTLPHeaders[k] != v {
					t.Errorf("Expected header %s=%s, got %s", k, v, cfg.OTLPHeaders[k])
				}
			}
		})
	}
}

func TestLoadConfig_TracingMetricsToggle(t *testing.T) {
	tests := []struct {
		envVars         map[string]string
		name            string
		expectedEnabled bool
		expectedTracing bool
		expectedMetrics bool
	}{
		{
			name: "all enabled",
			envVars: map[string]string{
				"OTEL_ENABLED": "true",
			},
			expectedEnabled: true,
			expectedTracing: true,
			expectedMetrics: true,
		},
		{
			name: "enabled but tracing explicitly disabled",
			envVars: map[string]string{
				"OTEL_ENABLED":         "true",
				"OTEL_TRACING_ENABLED": "false",
			},
			expectedEnabled: true,
			expectedTracing: false,
			expectedMetrics: true,
		},
		{
			name: "enabled but metrics explicitly disabled",
			envVars: map[string]string{
				"OTEL_ENABLED":         "true",
				"OTEL_METRICS_ENABLED": "false",
			},
			expectedEnabled: true,
			expectedTracing: true,
			expectedMetrics: false,
		},
		{
			name: "disabled with tracing enabled has no effect",
			envVars: map[string]string{
				"OTEL_ENABLED":         "false",
				"OTEL_TRACING_ENABLED": "true",
			},
			expectedEnabled: false,
			expectedTracing: false,
			expectedMetrics: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Clearenv()
			for k, v := range tt.envVars {
				os.Setenv(k, v)
			}

			cfg := LoadConfig()

			if cfg.Enabled != tt.expectedEnabled {
				t.Errorf("Expected Enabled=%v, got %v", tt.expectedEnabled, cfg.Enabled)
			}
			if cfg.TracingEnabled != tt.expectedTracing {
				t.Errorf("Expected TracingEnabled=%v, got %v", tt.expectedTracing, cfg.TracingEnabled)
			}
			if cfg.MetricsEnabled != tt.expectedMetrics {
				t.Errorf("Expected MetricsEnabled=%v, got %v", tt.expectedMetrics, cfg.MetricsEnabled)
			}
		})
	}
}
