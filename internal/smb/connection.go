// Package smb provides SMB connectivity and file operations.
package smb

import (
	"fmt"

	"github.com/bancey/document-smbrelay-service/internal/config"
)

const (
	statusHealthy   = "healthy"
	statusUnhealthy = "unhealthy"
	statusOK        = "ok"
	statusFailed    = "failed"
)

// HealthCheckResult represents the result of an SMB health check
// Fields are ordered for optimal memory alignment
type HealthCheckResult struct {
	Status             string `json:"status"`
	AppStatus          string `json:"app_status"`
	SMBConnection      string `json:"smb_connection"`
	Server             string `json:"server"`
	Share              string `json:"share"`
	Error              string `json:"error,omitempty"`
	SMBShareAccessible bool   `json:"smb_share_accessible"`
}

// CheckHealth performs a health check on the SMB server and share using smbclient
func CheckHealth(cfg *config.SMBConfig) *HealthCheckResult {
	result := &HealthCheckResult{
		AppStatus: statusOK,
		Server:    cfg.GetServerDisplay(),
		Share:     cfg.ShareName,
	}

	// Test connection using smbclient
	err := testConnection(cfg)
	if err != nil {
		result.Status = statusUnhealthy
		result.SMBConnection = statusFailed
		result.SMBShareAccessible = false
		result.Error = err.Error()
		return result
	}

	// If a base path is configured, validate it exists
	if cfg.BasePath != "" {
		err = testBasePath(cfg)
		if err != nil {
			result.Status = statusUnhealthy
			result.SMBConnection = statusOK
			result.SMBShareAccessible = false
			result.Error = fmt.Sprintf("base path validation failed: %v", err)
			return result
		}
	}

	result.Status = statusHealthy
	result.SMBConnection = statusOK
	result.SMBShareAccessible = true
	return result
}
