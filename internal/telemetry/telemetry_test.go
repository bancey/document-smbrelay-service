package telemetry

import (
	"context"
	"os"
	"testing"
	"time"
)

func TestInitialize_Disabled(t *testing.T) {
	cfg := &Config{
		Enabled: false,
	}

	provider, err := Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Initialize() with disabled telemetry should not error: %v", err)
	}
	if provider == nil {
		t.Fatal("Initialize() should return non-nil provider even when disabled")
	}
	if provider.config != cfg {
		t.Error("Initialize() should store config reference")
	}
}

func TestInitialize_WithStdout(t *testing.T) {
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		MetricsEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		// No OTLP endpoint - should use stdout exporters
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	provider, err := Initialize(ctx, cfg)
	if err != nil {
		t.Fatalf("Initialize() with stdout exporters failed: %v", err)
	}
	if provider == nil {
		t.Fatal("Initialize() should return non-nil provider")
	}
	if provider.tracerProvider == nil {
		t.Error("Initialize() should create tracer provider when tracing is enabled")
	}
	if provider.meterProvider == nil {
		t.Error("Initialize() should create meter provider when metrics are enabled")
	}

	// Test shutdown
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer shutdownCancel()

	err = provider.Shutdown(shutdownCtx)
	if err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
}

func TestInitialize_TracingOnly(t *testing.T) {
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		MetricsEnabled: false,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}

	ctx := context.Background()
	provider, err := Initialize(ctx, cfg)
	if err != nil {
		t.Fatalf("Initialize() with tracing only failed: %v", err)
	}
	if provider.tracerProvider == nil {
		t.Error("Initialize() should create tracer provider")
	}
	if provider.meterProvider != nil {
		t.Error("Initialize() should not create meter provider when metrics disabled")
	}

	// Clean up
	_ = provider.Shutdown(context.Background())
}

func TestInitialize_MetricsOnly(t *testing.T) {
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: false,
		MetricsEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}

	ctx := context.Background()
	provider, err := Initialize(ctx, cfg)
	if err != nil {
		t.Fatalf("Initialize() with metrics only failed: %v", err)
	}
	if provider.tracerProvider != nil {
		t.Error("Initialize() should not create tracer provider when tracing disabled")
	}
	if provider.meterProvider == nil {
		t.Error("Initialize() should create meter provider")
	}

	// Clean up
	_ = provider.Shutdown(context.Background())
}

func TestShutdown_Disabled(t *testing.T) {
	provider := &Provider{
		config: &Config{
			Enabled: false,
		},
	}

	err := provider.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Shutdown() with disabled telemetry should not error: %v", err)
	}
}

func TestShutdown_WithProviders(t *testing.T) {
	// Initialize a provider with both tracing and metrics
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		MetricsEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}

	ctx := context.Background()
	provider, err := Initialize(ctx, cfg)
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Test shutdown
	err = provider.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
}

func TestInitialize_WithAppInsights(t *testing.T) {
	// Test with a mock Application Insights connection string
	cfg := &Config{
		Enabled:                          true,
		TracingEnabled:                   true,
		MetricsEnabled:                   true,
		ServiceName:                      "test-service",
		ServiceVersion:                   "1.0.0",
		AzureAppInsightsConnectionString: "InstrumentationKey=test-key;IngestionEndpoint=https://test.applicationinsights.azure.com",
		OTLPEndpoint:                     "test.applicationinsights.azure.com",
		OTLPHeaders:                      make(map[string]string),
	}

	ctx := context.Background()
	provider, err := Initialize(ctx, cfg)

	// Note: This will fail to connect but should still initialize the provider
	if provider == nil {
		t.Fatal("Initialize() should return non-nil provider even if connection fails")
	}

	// The initialization might fail due to network issues, but we test that it doesn't panic
	if err == nil {
		// Clean up only if successful
		_ = provider.Shutdown(context.Background())
	}
}

func TestLoadConfig_WithHeaders(t *testing.T) {
	// Clean environment
	os.Clearenv()

	os.Setenv("OTEL_ENABLED", "true")
	os.Setenv("OTEL_EXPORTER_OTLP_HEADERS", "key1=value1,key2=value2")

	cfg := LoadConfig()

	if !cfg.Enabled {
		t.Error("Expected Enabled to be true")
	}

	if len(cfg.OTLPHeaders) != 2 {
		t.Errorf("Expected 2 headers, got %d", len(cfg.OTLPHeaders))
	}

	if cfg.OTLPHeaders["key1"] != "value1" {
		t.Errorf("Expected header key1=value1, got %s", cfg.OTLPHeaders["key1"])
	}

	if cfg.OTLPHeaders["key2"] != "value2" {
		t.Errorf("Expected header key2=value2, got %s", cfg.OTLPHeaders["key2"])
	}
}

func TestInitialize_WithOTLPEndpoint(t *testing.T) {
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		MetricsEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		OTLPEndpoint:   "localhost:4318",
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	provider, err := Initialize(ctx, cfg)
	// Note: This might fail to connect but should still initialize
	if provider == nil {
		t.Fatal("Initialize() should return non-nil provider")
	}

	// Clean up
	if err == nil {
		_ = provider.Shutdown(context.Background())
	}
}

func TestInitialize_WithOTLPHeaders(t *testing.T) {
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		MetricsEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
		OTLPEndpoint:   "localhost:4318",
		OTLPHeaders: map[string]string{
			"x-api-key":     "test-key",
			"custom-header": "custom-value",
		},
	}

	ctx := context.Background()
	provider, err := Initialize(ctx, cfg)

	if provider == nil {
		t.Fatal("Initialize() should return non-nil provider")
	}

	// Clean up
	if err == nil {
		_ = provider.Shutdown(context.Background())
	}
}

func TestShutdown_OnlyTracer(t *testing.T) {
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		MetricsEnabled: false,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}

	ctx := context.Background()
	provider, err := Initialize(ctx, cfg)
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Test shutdown with only tracer provider
	err = provider.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
}

func TestShutdown_OnlyMeter(t *testing.T) {
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: false,
		MetricsEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}

	ctx := context.Background()
	provider, err := Initialize(ctx, cfg)
	if err != nil {
		t.Fatalf("Initialize() failed: %v", err)
	}

	// Test shutdown with only meter provider
	err = provider.Shutdown(ctx)
	if err != nil {
		t.Errorf("Shutdown() failed: %v", err)
	}
}

func TestShutdown_NilProviders(t *testing.T) {
	provider := &Provider{
		config: &Config{
			Enabled: true,
		},
		tracerProvider: nil,
		meterProvider:  nil,
	}

	err := provider.Shutdown(context.Background())
	if err != nil {
		t.Errorf("Shutdown() should not error with nil providers: %v", err)
	}
}

func TestExtractInstrumentationKey_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		connStr  string
		expected string
	}{
		{
			name:     "empty string",
			connStr:  "",
			expected: "",
		},
		{
			name:     "key at end",
			connStr:  "IngestionEndpoint=https://test.com;InstrumentationKey=end-key",
			expected: "end-key",
		},
		{
			name:     "key in middle",
			connStr:  "LiveEndpoint=https://live.com;InstrumentationKey=mid-key;IngestionEndpoint=https://test.com",
			expected: "mid-key",
		},
		{
			name:     "only key",
			connStr:  "InstrumentationKey=only-key",
			expected: "only-key",
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

func TestStripScheme(t *testing.T) {
	tests := []struct {
		name     string
		endpoint string
		expected string
	}{
		{
			name:     "https URL",
			endpoint: "https://uksouth-1.in.applicationinsights.azure.com",
			expected: "uksouth-1.in.applicationinsights.azure.com",
		},
		{
			name:     "http URL",
			endpoint: "http://localhost:4318",
			expected: "localhost:4318",
		},
		{
			name:     "no scheme",
			endpoint: "localhost:4318",
			expected: "localhost:4318",
		},
		{
			name:     "hostname only",
			endpoint: "example.com",
			expected: "example.com",
		},
		{
			name:     "hostname with port",
			endpoint: "example.com:8080",
			expected: "example.com:8080",
		},
		{
			name:     "https with port",
			endpoint: "https://example.com:8080",
			expected: "example.com:8080",
		},
		{
			name:     "empty string",
			endpoint: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := stripScheme(tt.endpoint)
			if result != tt.expected {
				t.Errorf("stripScheme() = %v, want %v", result, tt.expected)
			}
		})
	}
}
