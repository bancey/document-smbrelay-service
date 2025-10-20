package telemetry

import (
	"context"
	"errors"
	"testing"

	"go.opentelemetry.io/otel/attribute"
)

func TestStartSMBSpan(t *testing.T) {
	// Initialize telemetry for testing
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}
	provider, err := Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer provider.Shutdown(context.Background())

	ctx := context.Background()
	operation := "test-operation"
	attrs := []attribute.KeyValue{
		attribute.String("test.attr", "test-value"),
	}

	newCtx, span := StartSMBSpan(ctx, operation, attrs...)
	if span == nil {
		t.Fatal("StartSMBSpan() should return non-nil span")
	}
	if newCtx == nil {
		t.Fatal("StartSMBSpan() should return non-nil context")
	}

	// End the span
	span.End()
}

func TestRecordSMBOperation_Success(t *testing.T) {
	// Initialize telemetry for testing
	cfg := &Config{
		Enabled:        true,
		MetricsEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}
	provider, err := Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer provider.Shutdown(context.Background())

	ctx := context.Background()
	operation := "upload"
	durationMs := 123.45
	attrs := []attribute.KeyValue{
		attribute.String("smb.server", "testserver"),
	}

	// Should not panic
	RecordSMBOperation(ctx, operation, durationMs, nil, attrs...)
}

func TestRecordSMBOperation_WithError(t *testing.T) {
	// Initialize telemetry for testing
	cfg := &Config{
		Enabled:        true,
		MetricsEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}
	provider, err := Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer provider.Shutdown(context.Background())

	ctx := context.Background()
	operation := "delete"
	durationMs := 50.0
	testErr := errors.New("test error")

	// Should not panic even with error
	RecordSMBOperation(ctx, operation, durationMs, testErr)
}

func TestRecordSMBFileSize(t *testing.T) {
	// Initialize telemetry for testing
	cfg := &Config{
		Enabled:        true,
		MetricsEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}
	provider, err := Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer provider.Shutdown(context.Background())

	ctx := context.Background()
	operation := "upload"
	sizeBytes := int64(1024 * 1024) // 1 MB

	// Should not panic
	RecordSMBFileSize(ctx, operation, sizeBytes)
}

func TestEndSpanWithError_NoError(t *testing.T) {
	// Initialize telemetry for testing
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}
	provider, err := Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer provider.Shutdown(context.Background())

	ctx, span := StartSMBSpan(context.Background(), "test")
	if ctx == nil || span == nil {
		t.Fatal("Failed to create span")
	}

	// Should not panic
	EndSpanWithError(span, nil)
}

func TestEndSpanWithError_WithError(t *testing.T) {
	// Initialize telemetry for testing
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}
	provider, err := Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer provider.Shutdown(context.Background())

	ctx, span := StartSMBSpan(context.Background(), "test")
	if ctx == nil || span == nil {
		t.Fatal("Failed to create span")
	}

	testErr := errors.New("test error")

	// Should not panic
	EndSpanWithError(span, testErr)
}

func TestAddSpanAttributes(t *testing.T) {
	// Initialize telemetry for testing
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}
	provider, err := Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer provider.Shutdown(context.Background())

	ctx, span := StartSMBSpan(context.Background(), "test")
	if ctx == nil || span == nil {
		t.Fatal("Failed to create span")
	}
	defer span.End()

	attrs := []attribute.KeyValue{
		attribute.String("test.key", "test.value"),
		attribute.Int("test.number", 42),
	}

	// Should not panic
	AddSpanAttributes(span, attrs...)
}

func TestAddSpanEvent(t *testing.T) {
	// Initialize telemetry for testing
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}
	provider, err := Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer provider.Shutdown(context.Background())

	ctx, span := StartSMBSpan(context.Background(), "test")
	if ctx == nil || span == nil {
		t.Fatal("Failed to create span")
	}
	defer span.End()

	eventName := "test-event"
	attrs := []attribute.KeyValue{
		attribute.String("event.detail", "test detail"),
	}

	// Should not panic
	AddSpanEvent(span, eventName, attrs...)
}

func TestInstrumentation_WithNilMetrics(t *testing.T) {
	// Test that functions handle nil metrics gracefully
	// This simulates when metric creation fails
	ctx := context.Background()

	// These should not panic even if metrics are nil
	RecordSMBOperation(ctx, "test", 100.0, nil)
	RecordSMBFileSize(ctx, "test", 1024)
}

func TestInstrumentation_WithoutTelemetry(t *testing.T) {
	// Test that functions work even without telemetry initialized
	ctx := context.Background()

	// These should not panic
	ctx2, span := StartSMBSpan(ctx, "test")
	if ctx2 == nil {
		t.Error("StartSMBSpan should return non-nil context even without telemetry")
	}
	if span == nil {
		t.Error("StartSMBSpan should return non-nil span even without telemetry")
	}

	EndSpanWithError(span, nil)
	AddSpanAttributes(span, attribute.String("test", "value"))
	AddSpanEvent(span, "test-event")
	RecordSMBOperation(ctx, "test", 100.0, nil)
	RecordSMBFileSize(ctx, "test", 1024)
}
