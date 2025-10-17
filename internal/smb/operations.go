package smb

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/bancey/document-smbrelay-service/internal/config"
)

// FileInfo represents information about a file or directory
type FileInfo struct {
	Name      string `json:"name"`
	Timestamp string `json:"timestamp,omitempty"`
	Size      int64  `json:"size"`
	IsDir     bool   `json:"is_dir"`
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

	args, env, err := buildSmbClientArgs(cfg, cmd)
	if err != nil {
		return nil, err
	}

	var output string
	if executor, ok := smbClientExec.(*DefaultSmbClientExecutor); ok {
		output, err = executor.ExecuteWithEnvAndLogging(args, env, cfg.LogSmbCommands)
	} else {
		// For mock executors in tests
		output, err = smbClientExec.Execute(args)
	}
	if err != nil {
		// Parse error messages
		if strings.Contains(output, "NT_STATUS_OBJECT_NAME_NOT_FOUND") ||
			strings.Contains(output, "NT_STATUS_OBJECT_PATH_NOT_FOUND") {
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
	lines := strings.Split(output, "\n")
	files := make([]FileInfo, 0, len(lines))

	// Regex to parse smbclient ls output format:
	// "  filename                        A     1024  Mon Jan  1 12:34:56 2024"
	// Captures: (1) filename, (2) attributes, (3) size, (4) timestamp
	lineRegex := regexp.MustCompile(`^\s+(.+?)\s+([A-Za-z]+)\s+(\d+)\s+(.*)$`)

	for _, line := range lines {
		// Check for empty lines (after trimming)
		if strings.TrimSpace(line) == "" {
			continue
		}

		// Skip header lines (check before regex matching)
		if strings.Contains(line, "blocks of size") ||
			strings.Contains(line, "blocks available") {
			continue
		}

		// Use regex to parse the line (don't trim before regex - it needs the leading whitespace)
		matches := lineRegex.FindStringSubmatch(line)
		if len(matches) != 5 {
			// Line doesn't match expected format, skip it
			continue
		}

		name := strings.TrimSpace(matches[1])
		attributes := matches[2]
		sizeStr := matches[3]
		timestamp := strings.TrimSpace(matches[4])

		// Skip "." and ".." entries
		if name == "." || name == ".." {
			continue
		}

		file := FileInfo{
			Name:      name,
			IsDir:     strings.Contains(attributes, "D"),
			Timestamp: timestamp,
		}

		// Parse size
		if size, err := strconv.ParseInt(sizeStr, 10, 64); err == nil {
			file.Size = size
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
		args, env, err := buildSmbClientArgs(cfg, checkCmd)
		if err != nil {
			return err
		}

		var output string
		if executor, ok := smbClientExec.(*DefaultSmbClientExecutor); ok {
			output, err = executor.ExecuteWithEnvAndLogging(args, env, cfg.LogSmbCommands)
		} else {
			// For mock executors in tests
			output, err = smbClientExec.Execute(args)
		}
		// If the file is found in the output, it exists
		// Note: We ignore the error here as the command may fail if file doesn't exist
		if err == nil && (strings.Contains(output, remotePath) || strings.Contains(output, "blocks of size")) {
			return fmt.Errorf("remote file already exists: %s", remotePath)
		}
	}

	// Upload the file
	return uploadFileViaSmbClient(localPath, remotePath, cfg)
}

// DeleteFile deletes a file from the SMB share using smbclient
func DeleteFile(remotePath string, cfg *config.SMBConfig) error {
	// Normalize remote path
	remotePath = strings.TrimPrefix(remotePath, "/")
	remotePath = strings.TrimPrefix(remotePath, "\\")
	remotePath = strings.ReplaceAll(remotePath, "\\", "/")

	if remotePath == "" || remotePath == "." {
		return fmt.Errorf("invalid remote path: cannot delete root directory")
	}

	// Build the del command
	cmd := fmt.Sprintf(`del "%s"`, remotePath)

	args, env, err := buildSmbClientArgs(cfg, cmd)
	if err != nil {
		return err
	}

	var output string
	if executor, ok := smbClientExec.(*DefaultSmbClientExecutor); ok {
		output, err = executor.ExecuteWithEnvAndLogging(args, env, cfg.LogSmbCommands)
	} else {
		// For mock executors in tests
		output, err = smbClientExec.Execute(args)
	}
	if err != nil {
		// Parse error messages
		if strings.Contains(output, "NT_STATUS_OBJECT_NAME_NOT_FOUND") ||
			strings.Contains(output, "NT_STATUS_OBJECT_PATH_NOT_FOUND") {
			return fmt.Errorf("file not found: %s", remotePath)
		}
		if strings.Contains(output, "NT_STATUS_ACCESS_DENIED") {
			return fmt.Errorf("access denied: cannot delete %s", remotePath)
		}
		if strings.Contains(output, "NT_STATUS_FILE_IS_A_DIRECTORY") {
			return fmt.Errorf("cannot delete directory: %s (use rmdir for directories)", remotePath)
		}
		return fmt.Errorf("failed to delete file: %w", err)
	}

	return nil
}
