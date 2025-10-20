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
