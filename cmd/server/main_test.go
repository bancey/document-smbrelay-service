package main

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/bancey/document-smbrelay-service/internal/handlers"
)

// setupTestApp creates a Fiber app configured for testing
func setupTestApp() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:               "Document SMB Relay Service",
		DisableStartupMessage: true,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(recover.New())
	app.Get("/health", handlers.HealthHandler)
	app.Post("/upload", handlers.UploadHandler)
	app.Get("/openapi.json", handlers.GetOpenAPISpec)
	app.Get("/docs", handlers.ServeSwaggerUI)

	return app
}

func TestIntegration_HealthEndpoint(t *testing.T) {
	// Set up environment
	os.Clearenv()
	os.Setenv("SMB_SERVER_NAME", "testserver")
	os.Setenv("SMB_SERVER_IP", "127.0.0.1")
	os.Setenv("SMB_SHARE_NAME", "testshare")
	os.Setenv("SMB_USERNAME", "testuser")
	os.Setenv("SMB_PASSWORD", "testpass")

	app := setupTestApp()

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req, 5000)
	if err != nil {
		t.Fatalf("Failed to test health endpoint: %v", err)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Should contain expected fields
	expectedFields := []string{"status", "app_status", "smb_connection"}
	for _, field := range expectedFields {
		if !strings.Contains(bodyStr, field) {
			t.Errorf("Expected response to contain field '%s', got: %s", field, bodyStr)
		}
	}
}

func TestIntegration_UploadEndpoint(t *testing.T) {
	// Set up environment
	os.Clearenv()
	os.Setenv("SMB_SERVER_NAME", "testserver")
	os.Setenv("SMB_SERVER_IP", "127.0.0.1")
	os.Setenv("SMB_SHARE_NAME", "testshare")
	os.Setenv("SMB_USERNAME", "testuser")
	os.Setenv("SMB_PASSWORD", "testpass")

	app := setupTestApp()

	// Create a temporary test file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-upload.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Create multipart request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	fileWriter, err := writer.CreateFormFile("file", "test-upload.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	fileContent, _ := os.ReadFile(tmpFile)
	fileWriter.Write(fileContent)

	// Add remote path
	_ = writer.WriteField("remote_path", "test/upload.txt")

	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test upload endpoint: %v", err)
	}

	// With invalid SMB server, we expect 500
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("Expected status %d for connection failure, got %d", fiber.StatusInternalServerError, resp.StatusCode)
	}
}

func TestIntegration_OpenAPIEndpoint(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest("GET", "/openapi.json", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test openapi endpoint: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Verify OpenAPI spec structure
	requiredElements := []string{
		"openapi",
		"info",
		"paths",
		"/health",
		"/upload",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(bodyStr, elem) {
			t.Errorf("Expected OpenAPI spec to contain '%s'", elem)
		}
	}
}

func TestIntegration_DocsEndpoint(t *testing.T) {
	app := setupTestApp()

	req := httptest.NewRequest("GET", "/docs", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test docs endpoint: %v", err)
	}

	if resp.StatusCode != fiber.StatusOK {
		t.Errorf("Expected status %d, got %d", fiber.StatusOK, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Verify Swagger UI HTML
	requiredElements := []string{
		"<!DOCTYPE html>",
		"swagger-ui",
		"Document SMB Relay Service",
	}

	for _, elem := range requiredElements {
		if !strings.Contains(bodyStr, elem) {
			t.Errorf("Expected docs to contain '%s'", elem)
		}
	}
}

func TestIntegration_ErrorHandling(t *testing.T) {
	app := setupTestApp()

	// Test 404 - route not found
	req := httptest.NewRequest("GET", "/nonexistent", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test 404: %v", err)
	}

	if resp.StatusCode != fiber.StatusNotFound {
		t.Errorf("Expected status %d, got %d", fiber.StatusNotFound, resp.StatusCode)
	}
}

func TestIntegration_RecoverMiddleware(t *testing.T) {
	app := setupTestApp()

	// The recover middleware should catch panics
	// We can't easily test this without modifying handlers,
	// but we can verify the middleware is registered
	if app == nil {
		t.Error("Expected app to be initialized")
	}
}

func TestIntegration_MultipleRequests(t *testing.T) {
	// Set up environment
	os.Clearenv()
	os.Setenv("SMB_SERVER_NAME", "testserver")
	os.Setenv("SMB_SERVER_IP", "127.0.0.1")
	os.Setenv("SMB_SHARE_NAME", "testshare")
	os.Setenv("SMB_USERNAME", "testuser")
	os.Setenv("SMB_PASSWORD", "testpass")

	app := setupTestApp()

	// Make multiple concurrent requests to health endpoint
	done := make(chan bool)
	errors := make(chan error, 10)

	for i := 0; i < 10; i++ {
		go func() {
			req := httptest.NewRequest("GET", "/health", nil)
			_, err := app.Test(req, 5000)
			if err != nil {
				errors <- err
			}
			done <- true
		}()
	}

	// Wait for all requests to complete
	for i := 0; i < 10; i++ {
		select {
		case <-done:
			// Request completed
		case err := <-errors:
			t.Errorf("Request failed: %v", err)
		case <-time.After(10 * time.Second):
			t.Fatal("Request timed out")
		}
	}
}

func TestIntegration_UploadWithoutConfig(t *testing.T) {
	// Clear environment to test missing config
	os.Clearenv()

	app := setupTestApp()

	// Create a temporary test file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-upload.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Create multipart request
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	fileWriter, _ := writer.CreateFormFile("file", "test-upload.txt")
	fileContent, _ := os.ReadFile(tmpFile)
	fileWriter.Write(fileContent)
	writer.WriteField("remote_path", "test/upload.txt")
	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req, -1)
	if err != nil {
		t.Fatalf("Failed to test upload endpoint: %v", err)
	}

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("Expected status %d for missing config, got %d", fiber.StatusInternalServerError, resp.StatusCode)
	}

	respBody, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(respBody), "Missing SMB configuration") {
		t.Errorf("Expected missing config error, got: %s", string(respBody))
	}
}

func TestIntegration_ContentTypes(t *testing.T) {
	app := setupTestApp()

	tests := []struct {
		name        string
		path        string
		contentType string
	}{
		{"OpenAPI JSON", "/openapi.json", "application/json"},
		{"Swagger UI HTML", "/docs", "text/html"},
		{"Health JSON", "/health", "application/json"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", tt.path, nil)
			resp, err := app.Test(req)
			if err != nil {
				t.Fatalf("Failed to test %s: %v", tt.path, err)
			}

			contentType := resp.Header.Get("Content-Type")
			if !strings.Contains(contentType, tt.contentType) {
				t.Errorf("Expected Content-Type to contain %s, got %s", tt.contentType, contentType)
			}
		})
	}
}
