package smb

import (
	"github.com/bancey/document-smbrelay-service/internal/config"
)

// HealthCheckResult represents the result of an SMB health check
type HealthCheckResult struct {
	Status             string `json:"status"`
	AppStatus          string `json:"app_status"`
	SMBConnection      string `json:"smb_connection"`
	SMBShareAccessible bool   `json:"smb_share_accessible"`
	Server             string `json:"server"`
	Share              string `json:"share"`
	Error              string `json:"error,omitempty"`
}

// CheckHealth performs a health check on the SMB server and share using smbclient
func CheckHealth(cfg *config.SMBConfig) *HealthCheckResult {
	result := &HealthCheckResult{
		AppStatus: "ok",
		Server:    cfg.GetServerDisplay(),
		Share:     cfg.ShareName,
	}

	// Test connection using smbclient
	err := testConnection(cfg)
	if err != nil {
		result.Status = "unhealthy"
		result.SMBConnection = "failed"
		result.SMBShareAccessible = false
		result.Error = err.Error()
		return result
	}

	result.Status = "healthy"
	result.SMBConnection = "ok"
	result.SMBShareAccessible = true
	return result
}
