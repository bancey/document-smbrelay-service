package smb

import (
	"fmt"
	"strings"

	"github.com/bancey/document-smbrelay-service/internal/config"
)

// FileInfo represents information about a file or directory
type FileInfo struct {
	Name      string `json:"name"`
	Size      int64  `json:"size"`
	IsDir     bool   `json:"is_dir"`
	Timestamp string `json:"timestamp,omitempty"`
}

// ListFiles lists files and folders at the given path on the SMB share
func ListFiles(remotePath string, cfg *config.SMBConfig) ([]FileInfo, error) {
	// Normalize remote path
	normalizedPath := strings.TrimPrefix(remotePath, "/")
	normalizedPath = strings.TrimPrefix(normalizedPath, "\\")
	normalizedPath = strings.ReplaceAll(normalizedPath, "\\", "/")

	// Build the ls command
	var cmd string
	if normalizedPath == "" || normalizedPath == "." {
		cmd = "ls"
	} else {
		cmd = fmt.Sprintf(`cd "%s"; ls`, normalizedPath)
	}

	args, err := buildSmbClientArgs(cfg, cmd)
	if err != nil {
		return nil, err
	}

	output, err := smbClientExec.Execute(args)
	if err != nil {
		// Parse error messages
		if strings.Contains(output, "NT_STATUS_OBJECT_NAME_NOT_FOUND") || strings.Contains(output, "NT_STATUS_OBJECT_PATH_NOT_FOUND") {
			return nil, fmt.Errorf("path not found: %s", remotePath)
		}
		if strings.Contains(output, "NT_STATUS_ACCESS_DENIED") {
			return nil, fmt.Errorf("access denied to path: %s", remotePath)
		}
		return nil, fmt.Errorf("failed to list files: %w", err)
	}

	// Parse the output
	return parseLsOutput(output), nil
}

// parseLsOutput parses the output from smbclient ls command
func parseLsOutput(output string) []FileInfo {
	var files []FileInfo
	lines := strings.Split(output, "\n")

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Skip header lines
		if strings.Contains(line, "blocks of size") ||
			strings.Contains(line, "blocks available") {
			continue
		}

		// Parse line format: "  filename                        A     size  timestamp"
		// or "  dirname                         D        0  timestamp"
		fields := strings.Fields(line)
		if len(fields) < 4 {
			continue
		}

		// First field is the name, second is attributes (D for directory, A for archive)
		name := fields[0]
		attributes := fields[1]

		// Skip "." and ".." entries
		if name == "." || name == ".." {
			continue
		}

		file := FileInfo{
			Name:  name,
			IsDir: strings.Contains(attributes, "D"),
		}

		// Try to parse size (third field)
		if len(fields) >= 3 {
			var size int64
			fmt.Sscanf(fields[2], "%d", &size)
			file.Size = size
		}

		// Timestamp is typically in fields[3:]
		if len(fields) > 3 {
			file.Timestamp = strings.Join(fields[3:], " ")
		}

		files = append(files, file)
	}

	return files
}

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

		output, err := smbClientExec.Execute(args)
		// If the file is found in the output, it exists
		// Note: We ignore the error here as the command may fail if file doesn't exist
		if err == nil && (strings.Contains(output, remotePath) || strings.Contains(output, "blocks of size")) {
			return fmt.Errorf("remote file already exists: %s", remotePath)
		}
	}

	// Upload the file
	return uploadFileViaSmbClient(localPath, remotePath, cfg)
}
