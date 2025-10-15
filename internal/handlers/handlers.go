package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/bancey/document-smbrelay-service/internal/config"
	"github.com/bancey/document-smbrelay-service/internal/smb"
	"github.com/gofiber/fiber/v2"
)

// HealthHandler handles GET /health requests
func HealthHandler(c *fiber.Ctx) error {
	cfg, missing := config.LoadFromEnv()

	if len(missing) > 0 {
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":               "unhealthy",
			"app_status":           "ok",
			"smb_connection":       "not_configured",
			"smb_share_accessible": false,
			"error":                fmt.Sprintf("Missing SMB configuration environment variables: %s", strings.Join(missing, ", ")),
		})
	}

	result := smb.CheckHealth(cfg)

	if result.Status == "healthy" {
		return c.JSON(result)
	}

	return c.Status(fiber.StatusServiceUnavailable).JSON(result)
}

// UploadHandler handles POST /upload requests
func UploadHandler(c *fiber.Ctx) error {
	// Load configuration
	cfg, missing := config.LoadFromEnv()
	if len(missing) > 0 {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("Missing SMB configuration environment variables: %s", strings.Join(missing, ", ")),
		})
	}

	// Get form parameters
	remotePath := c.FormValue("remote_path")
	if remotePath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"detail": "remote_path is required",
		})
	}

	overwriteStr := c.FormValue("overwrite")
	overwrite := overwriteStr == "true" || overwriteStr == "1"

	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"detail": "file is required",
		})
	}

	// Save uploaded file to temporary location
	tmpDir := os.TempDir()
	tmpPath := filepath.Join(tmpDir, fmt.Sprintf("smb-upload-%s", filepath.Base(file.Filename)))

	err = c.SaveFile(file, tmpPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("failed to save uploaded file: %v", err),
		})
	}
	defer os.Remove(tmpPath)

	// Upload to SMB share
	err = smb.UploadFile(tmpPath, remotePath, cfg, overwrite)
	if err != nil {
		// Check if it's a file exists error
		if strings.Contains(err.Error(), "already exists") {
			return c.Status(fiber.StatusConflict).JSON(fiber.Map{
				"detail": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status":      "ok",
		"remote_path": remotePath,
	})
}

// GetOpenAPISpec returns the OpenAPI specification
func GetOpenAPISpec(c *fiber.Ctx) error {
	spec := map[string]interface{}{
		"openapi": "3.0.0",
		"info": map[string]interface{}{
			"title":       "Document SMB Relay Service",
			"version":     "1.0.0",
			"description": "A minimal service that accepts file uploads via HTTP and writes them directly to SMB shares",
		},
		"paths": map[string]interface{}{
			"/health": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "Health check endpoint",
					"description": "Verifies application responsiveness and SMB connectivity",
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Application and SMB server are healthy",
						},
						"503": map[string]interface{}{
							"description": "Application is unhealthy or SMB server is inaccessible",
						},
					},
				},
			},
			"/upload": map[string]interface{}{
				"post": map[string]interface{}{
					"summary":     "Upload file to SMB share",
					"description": "Accepts multipart/form-data and writes file to SMB share",
					"requestBody": map[string]interface{}{
						"required": true,
						"content": map[string]interface{}{
							"multipart/form-data": map[string]interface{}{
								"schema": map[string]interface{}{
									"type": "object",
									"properties": map[string]interface{}{
										"file": map[string]interface{}{
											"type":        "string",
											"format":      "binary",
											"description": "The file to upload",
										},
										"remote_path": map[string]interface{}{
											"type":        "string",
											"description": "Path within the SMB share",
										},
										"overwrite": map[string]interface{}{
											"type":        "boolean",
											"description": "Whether to overwrite existing files",
											"default":     false,
										},
									},
									"required": []string{"file", "remote_path"},
								},
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "Upload successful",
						},
						"409": map[string]interface{}{
							"description": "File exists and overwrite is false",
						},
						"500": map[string]interface{}{
							"description": "Upload failed",
						},
					},
				},
			},
		},
	}

	return c.JSON(spec)
}

// ServeSwaggerUI serves a simple Swagger UI HTML page
func ServeSwaggerUI(c *fiber.Ctx) error {
	html := `<!DOCTYPE html>
<html lang="en">
<head>
    <meta charset="UTF-8">
    <meta name="viewport" content="width=device-width, initial-scale=1.0">
    <title>Document SMB Relay Service - Swagger UI</title>
    <link rel="stylesheet" type="text/css" href="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui.css">
    <style>
        body {
            margin: 0;
            padding: 0;
        }
    </style>
</head>
<body>
    <div id="swagger-ui"></div>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-bundle.js"></script>
    <script src="https://unpkg.com/swagger-ui-dist@5.9.0/swagger-ui-standalone-preset.js"></script>
    <script>
        window.onload = function() {
            SwaggerUIBundle({
                url: "/openapi.json",
                dom_id: '#swagger-ui',
                deepLinking: true,
                presets: [
                    SwaggerUIBundle.presets.apis,
                    SwaggerUIStandalonePreset
                ],
                plugins: [
                    SwaggerUIBundle.plugins.DownloadUrl
                ],
                layout: "StandaloneLayout"
            });
        };
    </script>
</body>
</html>`

	c.Set("Content-Type", "text/html")
	return c.SendString(html)
}
