// Package telemetry provides OpenTelemetry instrumentation for the application.
package telemetry

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"go.opentelemetry.io/otel/trace"
)

// Middleware returns a Fiber middleware that instruments HTTP requests with OpenTelemetry
func Middleware(serviceName string) fiber.Handler {
	tracer := otel.Tracer(serviceName)
	meter := otel.Meter(serviceName)

	// Create metrics
	httpRequestDuration, _ := meter.Float64Histogram(
		"http.server.request.duration",
		metric.WithDescription("Duration of HTTP requests"),
		metric.WithUnit("ms"),
	)

	httpRequestsTotal, _ := meter.Int64Counter(
		"http.server.requests.total",
		metric.WithDescription("Total number of HTTP requests"),
	)

	httpRequestSize, _ := meter.Int64Histogram(
		"http.server.request.size",
		metric.WithDescription("Size of HTTP requests"),
		metric.WithUnit("bytes"),
	)

	httpResponseSize, _ := meter.Int64Histogram(
		"http.server.response.size",
		metric.WithDescription("Size of HTTP responses"),
		metric.WithUnit("bytes"),
	)

	return func(c *fiber.Ctx) error {
		// Skip telemetry for health checks if desired
		// if c.Path() == "/health" {
		// 	return c.Next()
		// }

		// Extract trace context from headers
		ctx := otel.GetTextMapPropagator().Extract(
			c.Context(),
			propagation.HeaderCarrier(c.GetReqHeaders()),
		)

		// Start span
		spanName := fmt.Sprintf("%s %s", c.Method(), c.Path())
		ctx, span := tracer.Start(ctx, spanName,
			trace.WithSpanKind(trace.SpanKindServer),
			trace.WithAttributes(
				semconv.HTTPRequestMethodKey.String(c.Method()),
				semconv.HTTPRouteKey.String(c.Path()),
				semconv.URLFull(c.OriginalURL()),
				semconv.URLScheme(c.Protocol()),
				semconv.ServerAddress(c.Hostname()),
				semconv.UserAgentOriginal(string(c.Request().Header.UserAgent())),
				attribute.String("client.address", c.IP()),
			),
		)
		defer span.End()

		// Store context in fiber context
		c.SetUserContext(ctx)

		// Record request start time
		startTime := time.Now()

		// Process request
		err := c.Next()

		// Calculate duration
		duration := time.Since(startTime).Milliseconds()
		statusCode := c.Response().StatusCode()

		// Common attributes for metrics
		attrs := []attribute.KeyValue{
			attribute.String("http.method", c.Method()),
			attribute.String("http.route", c.Path()),
			attribute.Int("http.status_code", statusCode),
		}

		// Record metrics
		httpRequestDuration.Record(ctx, float64(duration), metric.WithAttributes(attrs...))
		httpRequestsTotal.Add(ctx, 1, metric.WithAttributes(attrs...))

		// Record request/response sizes
		requestSize := len(c.Request().Body())
		responseSize := len(c.Response().Body())
		httpRequestSize.Record(ctx, int64(requestSize), metric.WithAttributes(attrs...))
		httpResponseSize.Record(ctx, int64(responseSize), metric.WithAttributes(attrs...))

		// Update span with response information
		span.SetAttributes(
			semconv.HTTPResponseStatusCode(statusCode),
			attribute.Int64("http.request.body.size", int64(requestSize)),
			attribute.Int64("http.response.body.size", int64(responseSize)),
		)

		// Set span status based on HTTP status code
		if statusCode >= 400 {
			span.SetStatus(codes.Error, fmt.Sprintf("HTTP %d", statusCode))
			if err != nil {
				span.RecordError(err)
			}
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}
