package smb

import (
	"fmt"
	"strings"

	"github.com/bancey/document-smbrelay-service/internal/config"
)

// UploadFile uploads a local file to the SMB share using smbclient
func UploadFile(localPath string, remotePath string, cfg *config.SMBConfig, overwrite bool) error {
	// Normalize remote path (remove leading slash)
	remotePath = strings.TrimPrefix(remotePath, "/")
	remotePath = strings.TrimPrefix(remotePath, "\\")
	
	// If overwrite is false, we need to check if file exists first
	// Skip the check if remotePath is empty (uploading to root with original filename)
	if !overwrite && remotePath != "" {
		// Try to stat the file - if it exists, smbclient will show it
		checkCmd := fmt.Sprintf("ls \"%s\"", remotePath)
		args, err := buildSmbClientArgs(cfg, checkCmd)
		if err != nil {
			return err
		}
		
		output, _ := smbClientExec.Execute(args)
		// If the file is found in the output, it exists
		if strings.Contains(output, remotePath) || strings.Contains(output, "blocks of size") {
			return fmt.Errorf("remote file already exists: %s", remotePath)
		}
	}
	
	// Upload the file
	return uploadFileViaSmbClient(localPath, remotePath, cfg)
}
