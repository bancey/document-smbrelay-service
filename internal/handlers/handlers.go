// Package handlers provides HTTP request handlers for the SMB relay service.
package handlers

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofiber/fiber/v2"

	"github.com/bancey/document-smbrelay-service/internal/config"
	"github.com/bancey/document-smbrelay-service/internal/logger"
	"github.com/bancey/document-smbrelay-service/internal/smb"
)

// HealthHandler handles GET /health requests
func HealthHandler(c *fiber.Ctx) error {
	cfg, missing := config.LoadFromEnv()

	if len(missing) > 0 {
		missingVars := strings.Join(missing, ", ")
		errorMsg := fmt.Sprintf("Missing SMB configuration environment variables: %s", missingVars)
		return c.Status(fiber.StatusServiceUnavailable).JSON(fiber.Map{
			"status":               "unhealthy",
			"app_status":           "ok",
			"smb_connection":       "not_configured",
			"smb_share_accessible": false,
			"error":                errorMsg,
		})
	}

	result := smb.CheckHealth(cfg)

	if result.Status == "healthy" {
		return c.JSON(result)
	}

	return c.Status(fiber.StatusServiceUnavailable).JSON(result)
}

// ListHandler handles GET /list requests
func ListHandler(c *fiber.Ctx) error {
	// Load configuration
	cfg, missing := config.LoadFromEnv()
	if len(missing) > 0 {
		missingVars := strings.Join(missing, ", ")
		errorMsg := fmt.Sprintf("Missing SMB configuration environment variables: %s", missingVars)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": errorMsg,
		})
	}

	// Get path from query parameter (default to root)
	path := c.Query("path", "")

	// List files
	files, err := smb.ListFiles(path, cfg)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"detail": err.Error(),
			})
		}
		if strings.Contains(err.Error(), "access denied") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"detail": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"path":  path,
		"files": files,
	})
}

// UploadHandler handles POST /upload requests
func UploadHandler(c *fiber.Ctx) error {
	// Load configuration
	cfg, missing := config.LoadFromEnv()
	if len(missing) > 0 {
		missingVars := strings.Join(missing, ", ")
		errorMsg := fmt.Sprintf("Missing SMB configuration environment variables: %s", missingVars)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": errorMsg,
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

	// Save uploaded file to temp location
	tmpDir := os.TempDir()
	tmpPath := filepath.Join(tmpDir, fmt.Sprintf("smb-upload-%s", filepath.Base(file.Filename)))

	err = c.SaveFile(file, tmpPath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": fmt.Sprintf("Failed to save uploaded file: %v", err),
		})
	}
	defer func() {
		if removeErr := os.Remove(tmpPath); removeErr != nil {
			logger.Error("Failed to remove temp file %s: %v", tmpPath, removeErr)
		}
	}()

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

// DeleteHandler handles DELETE /delete requests
func DeleteHandler(c *fiber.Ctx) error {
	// Load configuration
	cfg, missing := config.LoadFromEnv()
	if len(missing) > 0 {
		missingVars := strings.Join(missing, ", ")
		errorMsg := fmt.Sprintf("Missing SMB configuration environment variables: %s", missingVars)
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": errorMsg,
		})
	}

	// Get path from query parameter
	remotePath := c.Query("path")
	if remotePath == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"detail": "path is required",
		})
	}

	// Delete file from SMB share
	err := smb.DeleteFile(remotePath, cfg)
	if err != nil {
		if strings.Contains(err.Error(), "not found") {
			return c.Status(fiber.StatusNotFound).JSON(fiber.Map{
				"detail": err.Error(),
			})
		}
		if strings.Contains(err.Error(), "access denied") {
			return c.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"detail": err.Error(),
			})
		}
		if strings.Contains(err.Error(), "invalid remote path") || strings.Contains(err.Error(), "cannot delete directory") {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"detail": err.Error(),
			})
		}
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"detail": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"path":   remotePath,
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
			"/list": map[string]interface{}{
				"get": map[string]interface{}{
					"summary":     "List files and folders",
					"description": "Lists files and folders at a given path on the SMB share",
					"parameters": []map[string]interface{}{
						{
							"name":        "path",
							"in":          "query",
							"description": "Path within the SMB share (defaults to root)",
							"required":    false,
							"schema": map[string]interface{}{
								"type":    "string",
								"default": "",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "List of files and folders",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"path": map[string]interface{}{
												"type": "string",
											},
											"files": map[string]interface{}{
												"type": "array",
												"items": map[string]interface{}{
													"type": "object",
													"properties": map[string]interface{}{
														"name": map[string]interface{}{
															"type": "string",
														},
														"size": map[string]interface{}{
															"type": "integer",
														},
														"is_dir": map[string]interface{}{
															"type": "boolean",
														},
														"timestamp": map[string]interface{}{
															"type": "string",
														},
													},
												},
											},
										},
									},
								},
							},
						},
						"404": map[string]interface{}{
							"description": "Path not found",
						},
						"403": map[string]interface{}{
							"description": "Access denied",
						},
						"500": map[string]interface{}{
							"description": "Server error",
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
			"/delete": map[string]interface{}{
				"delete": map[string]interface{}{
					"summary":     "Delete file from SMB share",
					"description": "Deletes a file at the specified path on the SMB share",
					"parameters": []map[string]interface{}{
						{
							"name":        "path",
							"in":          "query",
							"description": "Path to the file within the SMB share",
							"required":    true,
							"schema": map[string]interface{}{
								"type": "string",
							},
						},
					},
					"responses": map[string]interface{}{
						"200": map[string]interface{}{
							"description": "File deleted successfully",
							"content": map[string]interface{}{
								"application/json": map[string]interface{}{
									"schema": map[string]interface{}{
										"type": "object",
										"properties": map[string]interface{}{
											"status": map[string]interface{}{
												"type": "string",
											},
											"path": map[string]interface{}{
												"type": "string",
											},
										},
									},
								},
							},
						},
						"400": map[string]interface{}{
							"description": "Invalid path or attempting to delete directory",
						},
						"403": map[string]interface{}{
							"description": "Access denied",
						},
						"404": map[string]interface{}{
							"description": "File not found",
						},
						"500": map[string]interface{}{
							"description": "Server error",
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
