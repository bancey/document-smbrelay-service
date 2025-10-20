// Package telemetry provides OpenTelemetry instrumentation for the application.
package telemetry

import (
	"context"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

const (
	instrumentationName = "github.com/bancey/document-smbrelay-service"
)

var (
	tracer trace.Tracer
	meter  metric.Meter

	// SMB operation metrics
	smbOperationDuration metric.Float64Histogram
	smbOperationsTotal   metric.Int64Counter
	smbErrorsTotal       metric.Int64Counter
	smbFileSize          metric.Int64Histogram
)

func init() {
	tracer = otel.Tracer(instrumentationName)
	meter = otel.Meter(instrumentationName)

	// Initialize metrics
	var err error

	smbOperationDuration, err = meter.Float64Histogram(
		"smb.operation.duration",
		metric.WithDescription("Duration of SMB operations"),
		metric.WithUnit("ms"),
	)
	if err != nil {
		// Log error but don't fail initialization
	}

	smbOperationsTotal, err = meter.Int64Counter(
		"smb.operations.total",
		metric.WithDescription("Total number of SMB operations"),
	)
	if err != nil {
		// Log error but don't fail initialization
	}

	smbErrorsTotal, err = meter.Int64Counter(
		"smb.errors.total",
		metric.WithDescription("Total number of SMB errors"),
	)
	if err != nil {
		// Log error but don't fail initialization
	}

	smbFileSize, err = meter.Int64Histogram(
		"smb.file.size",
		metric.WithDescription("Size of files in SMB operations"),
		metric.WithUnit("bytes"),
	)
	if err != nil {
		// Log error but don't fail initialization
	}
}

// StartSMBSpan starts a new span for an SMB operation
func StartSMBSpan(ctx context.Context, operation string, attributes ...attribute.KeyValue) (context.Context, trace.Span) {
	attrs := append([]attribute.KeyValue{
		attribute.String("smb.operation", operation),
	}, attributes...)

	return tracer.Start(ctx, "SMB "+operation,
		trace.WithSpanKind(trace.SpanKindClient),
		trace.WithAttributes(attrs...),
	)
}

// RecordSMBOperation records metrics for an SMB operation
func RecordSMBOperation(ctx context.Context, operation string, durationMs float64, err error, attrs ...attribute.KeyValue) {
	baseAttrs := []attribute.KeyValue{
		attribute.String("operation", operation),
	}
	baseAttrs = append(baseAttrs, attrs...)

	// Record duration
	if smbOperationDuration != nil {
		smbOperationDuration.Record(ctx, durationMs, metric.WithAttributes(baseAttrs...))
	}

	// Record operation count
	if smbOperationsTotal != nil {
		successAttr := attribute.Bool("success", err == nil)
		opAttrs := append(baseAttrs, successAttr)
		smbOperationsTotal.Add(ctx, 1, metric.WithAttributes(opAttrs...))
	}

	// Record error if present
	if err != nil && smbErrorsTotal != nil {
		errorAttrs := append(baseAttrs, attribute.String("error", err.Error()))
		smbErrorsTotal.Add(ctx, 1, metric.WithAttributes(errorAttrs...))
	}
}

// RecordSMBFileSize records the size of a file in an SMB operation
func RecordSMBFileSize(ctx context.Context, operation string, sizeBytes int64) {
	if smbFileSize != nil {
		attrs := []attribute.KeyValue{
			attribute.String("operation", operation),
		}
		smbFileSize.Record(ctx, sizeBytes, metric.WithAttributes(attrs...))
	}
}

// EndSpanWithError ends a span and records an error if present
func EndSpanWithError(span trace.Span, err error) {
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
	} else {
		span.SetStatus(codes.Ok, "")
	}
	span.End()
}

// AddSpanAttributes adds attributes to a span
func AddSpanAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// AddSpanEvent adds an event to a span
func AddSpanEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attrs...))
}
