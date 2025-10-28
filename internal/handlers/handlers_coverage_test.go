package handlers

import (
	"bytes"
	"io"
	"mime/multipart"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

// TestListHandler_MissingConfigExtended tests ListHandler with missing SMB configuration
func TestListHandler_MissingConfigExtended(t *testing.T) {
	// Clear all SMB environment variables
	os.Clearenv()

	app := fiber.New()
	app.Get("/list", ListHandler)

	req := httptest.NewRequest("GET", "/list", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test list endpoint: %v", err)
	}

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", fiber.StatusInternalServerError, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Missing SMB configuration") {
		t.Errorf("Expected missing config message in response, got: %s", string(body))
	}
}

// TestListHandler_EmptyPath tests ListHandler with empty path (should default to root)
func TestListHandler_EmptyPath(t *testing.T) {
	// Set up test environment variables
	os.Clearenv()
	os.Setenv("SMB_SERVER_NAME", "testserver")
	os.Setenv("SMB_SERVER_IP", "127.0.0.1")
	os.Setenv("SMB_SHARE_NAME", "testshare")
	os.Setenv("SMB_USERNAME", "testuser")
	os.Setenv("SMB_PASSWORD", "testpass")

	app := fiber.New()
	app.Get("/list", ListHandler)

	req := httptest.NewRequest("GET", "/list", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test list endpoint: %v", err)
	}

	// Empty path should attempt to list root, but will fail with 500 due to no SMB server
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", fiber.StatusInternalServerError, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "failed to list files") && !strings.Contains(string(body), "detail") {
		t.Errorf("Expected error message in response, got: %s", string(body))
	}
}

// TestDeleteHandler_MissingConfigExtended tests DeleteHandler with missing SMB configuration
func TestDeleteHandler_MissingConfigExtended(t *testing.T) {
	// Clear all SMB environment variables
	os.Clearenv()

	app := fiber.New()
	app.Delete("/delete", DeleteHandler)

	req := httptest.NewRequest("DELETE", "/delete", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test delete endpoint: %v", err)
	}

	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", fiber.StatusInternalServerError, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "Missing SMB configuration") {
		t.Errorf("Expected missing config message in response, got: %s", string(body))
	}
}

// TestDeleteHandler_MissingPathExtended tests DeleteHandler with missing path parameter
func TestDeleteHandler_MissingPathExtended(t *testing.T) {
	// Set up test environment variables
	os.Clearenv()
	os.Setenv("SMB_SERVER_NAME", "testserver")
	os.Setenv("SMB_SERVER_IP", "127.0.0.1")
	os.Setenv("SMB_SHARE_NAME", "testshare")
	os.Setenv("SMB_USERNAME", "testuser")
	os.Setenv("SMB_PASSWORD", "testpass")

	app := fiber.New()
	app.Delete("/delete", DeleteHandler)

	req := httptest.NewRequest("DELETE", "/delete", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test delete endpoint: %v", err)
	}

	if resp.StatusCode != fiber.StatusBadRequest {
		t.Errorf("Expected status %d, got %d", fiber.StatusBadRequest, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	if !strings.Contains(string(body), "path is required") {
		t.Errorf("Expected 'path is required' message in response, got: %s", string(body))
	}
}

// TestUploadHandler_InvalidOverwriteValue tests UploadHandler with invalid overwrite parameter
func TestUploadHandler_InvalidOverwriteValue(t *testing.T) {
	// Set up test environment variables
	os.Clearenv()
	os.Setenv("SMB_SERVER_NAME", "testserver")
	os.Setenv("SMB_SERVER_IP", "127.0.0.1")
	os.Setenv("SMB_SHARE_NAME", "testshare")
	os.Setenv("SMB_USERNAME", "testuser")
	os.Setenv("SMB_PASSWORD", "testpass")

	app := fiber.New()
	app.Post("/upload", UploadHandler)

	// Create a temporary test file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-upload-invalid-overwrite.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Create multipart request with invalid overwrite value
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add file
	fileWriter, err := writer.CreateFormFile("file", "test-upload.txt")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	fileContent, _ := os.ReadFile(tmpFile)
	fileWriter.Write(fileContent)

	// Add remote path and invalid overwrite value
	_ = writer.WriteField("remote_path", "test/upload.txt")
	_ = writer.WriteField("overwrite", "invalid-bool")

	writer.Close()

	req := httptest.NewRequest("POST", "/upload", body)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	resp, err := app.Test(req, -1) // -1 = no timeout for potentially long operation
	if err != nil {
		t.Fatalf("Failed to test upload endpoint: %v", err)
	}

	// Invalid boolean should be treated as false, so the upload should attempt
	// Since SMB server is not available, we expect a 500 error
	if resp.StatusCode != fiber.StatusInternalServerError {
		t.Errorf("Expected status %d, got %d", fiber.StatusInternalServerError, resp.StatusCode)
	}
}

// TestHealthHandler_CheckHealthError tests HealthHandler when CheckHealth returns error
func TestHealthHandler_CheckHealthError(t *testing.T) {
	// Set up test environment variables with valid config but unreachable server
	os.Clearenv()
	os.Setenv("SMB_SERVER_NAME", "testserver")
	os.Setenv("SMB_SERVER_IP", "127.0.0.1")
	os.Setenv("SMB_SHARE_NAME", "testshare")
	os.Setenv("SMB_USERNAME", "testuser")
	os.Setenv("SMB_PASSWORD", "testpass")

	app := fiber.New()
	app.Get("/health", HealthHandler)

	req := httptest.NewRequest("GET", "/health", nil)
	resp, err := app.Test(req)
	if err != nil {
		t.Fatalf("Failed to test health endpoint: %v", err)
	}

	// With unreachable SMB server, we expect 503
	if resp.StatusCode != fiber.StatusServiceUnavailable {
		t.Errorf("Expected status %d, got %d", fiber.StatusServiceUnavailable, resp.StatusCode)
	}

	body, _ := io.ReadAll(resp.Body)
	bodyStr := string(body)

	// Should contain unhealthy status
	if !strings.Contains(bodyStr, "unhealthy") || !strings.Contains(bodyStr, "status") {
		t.Errorf("Expected unhealthy status in response, got: %s", bodyStr)
	}
}
