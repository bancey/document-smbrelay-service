package telemetry

import (
	"context"
	"io"
	"net/http/httptest"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestMiddleware_Disabled(t *testing.T) {
	// Create Fiber app
	app := fiber.New()

	// Create a simple test handler
	called := false
	app.Use(func(c *fiber.Ctx) error {
		called = true
		return c.SendString("test")
	})

	// Create test request
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if !called {
		t.Error("Handler should have been called")
	}
	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMiddleware_Enabled(t *testing.T) {
	// Initialize telemetry
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		MetricsEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}
	provider, err := Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer provider.Shutdown(context.Background())

	// Create Fiber app with middleware
	app := fiber.New()
	app.Use(Middleware("test-service"))

	// Add test route
	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("test response")
	})

	// Test GET request
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if string(body) != "test response" {
		t.Errorf("Expected 'test response', got %s", string(body))
	}
}

func TestMiddleware_WithError(t *testing.T) {
	// Initialize telemetry
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		MetricsEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}
	provider, err := Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer provider.Shutdown(context.Background())

	// Create Fiber app with middleware
	app := fiber.New()
	app.Use(Middleware("test-service"))

	// Add test route that returns error
	app.Get("/error", func(c *fiber.Ctx) error {
		return fiber.NewError(500, "test error")
	})

	// Test error request
	req := httptest.NewRequest("GET", "/error", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 500 {
		t.Errorf("Expected status 500, got %d", resp.StatusCode)
	}
}

func TestMiddleware_404(t *testing.T) {
	// Initialize telemetry
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		MetricsEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}
	provider, err := Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer provider.Shutdown(context.Background())

	// Create Fiber app with middleware
	app := fiber.New()
	app.Use(Middleware("test-service"))

	// Test 404 request
	req := httptest.NewRequest("GET", "/notfound", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 404 {
		t.Errorf("Expected status 404, got %d", resp.StatusCode)
	}
}

func TestMiddleware_POST(t *testing.T) {
	// Initialize telemetry
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		MetricsEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}
	provider, err := Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer provider.Shutdown(context.Background())

	// Create Fiber app with middleware
	app := fiber.New()
	app.Use(Middleware("test-service"))

	// Add POST route
	app.Post("/test", func(c *fiber.Ctx) error {
		return c.SendString("posted")
	})

	// Test POST request
	req := httptest.NewRequest("POST", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMiddleware_WithNilMetrics(t *testing.T) {
	// Test that middleware handles nil metrics gracefully
	// This simulates when metric creation fails

	// Create Fiber app with middleware (without initializing telemetry)
	app := fiber.New()
	app.Use(Middleware("test-service"))

	app.Get("/test", func(c *fiber.Ctx) error {
		return c.SendString("ok")
	})

	// Test request - should not panic even with nil metrics
	req := httptest.NewRequest("GET", "/test", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
}

func TestMiddleware_DifferentStatusCodes(t *testing.T) {
	// Initialize telemetry
	cfg := &Config{
		Enabled:        true,
		TracingEnabled: true,
		MetricsEnabled: true,
		ServiceName:    "test-service",
		ServiceVersion: "1.0.0",
	}
	provider, err := Initialize(context.Background(), cfg)
	if err != nil {
		t.Fatalf("Failed to initialize telemetry: %v", err)
	}
	defer provider.Shutdown(context.Background())

	testCases := []struct {
		name       string
		statusCode int
		handler    fiber.Handler
	}{
		{
			name:       "200 OK",
			statusCode: 200,
			handler: func(c *fiber.Ctx) error {
				return c.SendStatus(200)
			},
		},
		{
			name:       "201 Created",
			statusCode: 201,
			handler: func(c *fiber.Ctx) error {
				return c.SendStatus(201)
			},
		},
		{
			name:       "400 Bad Request",
			statusCode: 400,
			handler: func(c *fiber.Ctx) error {
				return c.SendStatus(400)
			},
		},
		{
			name:       "403 Forbidden",
			statusCode: 403,
			handler: func(c *fiber.Ctx) error {
				return c.SendStatus(403)
			},
		},
		{
			name:       "500 Internal Server Error",
			statusCode: 500,
			handler: func(c *fiber.Ctx) error {
				return c.SendStatus(500)
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			app := fiber.New()
			app.Use(Middleware("test-service"))
			app.Get("/test", tc.handler)

			req := httptest.NewRequest("GET", "/test", nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to test request: %v", err)
			}
			defer resp.Body.Close()

			if resp.StatusCode != tc.statusCode {
				t.Errorf("Expected status %d, got %d", tc.statusCode, resp.StatusCode)
			}
		})
	}
}
