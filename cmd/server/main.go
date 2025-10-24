// Package main provides the entry point for the SMB relay service.
package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/bancey/document-smbrelay-service/internal/handlers"
	"github.com/bancey/document-smbrelay-service/internal/logger"
	"github.com/bancey/document-smbrelay-service/internal/telemetry"
)

func main() {
	// Initialize logger
	logger.Info("Starting Document SMB Relay Service")

	// Initialize OpenTelemetry
	ctx := context.Background()
	telemetryConfig := telemetry.LoadConfig()
	telemetryProvider, err := telemetry.Initialize(ctx, telemetryConfig)
	if err != nil {
		logger.Error("Failed to initialize telemetry: %v", err)
		// Continue without telemetry rather than failing
	}
	defer func() {
		if telemetryProvider != nil {
			shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if err := telemetryProvider.Shutdown(shutdownCtx); err != nil {
				logger.Error("Failed to shutdown telemetry: %v", err)
			}
		}
	}()

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName:               "Document SMB Relay Service",
		DisableStartupMessage: false,
		ReadBufferSize:        16 * 1024, // 16KB - increased from default 4KB to handle larger headers
		// (e.g., OpenTelemetry trace context, large cookies, auth tokens)
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			logger.Error("Request error: %v", err)
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	// Middleware
	app.Use(recover.New())

	// Add OpenTelemetry middleware if enabled
	if telemetryConfig.Enabled {
		app.Use(telemetry.Middleware(telemetryConfig.ServiceName))
		logger.Info("OpenTelemetry middleware enabled")
	}

	// Routes
	app.Get("/health", handlers.HealthHandler)
	app.Get("/list", handlers.ListHandler)
	app.Post("/upload", handlers.UploadHandler)
	app.Delete("/delete", handlers.DeleteHandler)
	app.Get("/openapi.json", handlers.GetOpenAPISpec)
	app.Get("/docs", handlers.ServeSwaggerUI)

	// Handle graceful shutdown
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-c
		logger.Info("Shutting down server...")
		if err := app.Shutdown(); err != nil {
			logger.Error("Error during shutdown: %v", err)
		}
	}()

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	logger.Info("Server starting on port %s", port)
	if err := app.Listen(fmt.Sprintf("0.0.0.0:%s", port)); err != nil {
		logger.Error("Server error: %v", err)
		os.Exit(1)
	}
}
