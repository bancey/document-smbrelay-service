package smb

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bancey/document-smbrelay-service/internal/config"
)

// ClientExecutor defines the interface for executing smbclient commands
// This allows for easy mocking in tests
type ClientExecutor interface {
	Execute(args []string) (string, error)
}

// DefaultSmbClientExecutor uses the real smbclient binary
type DefaultSmbClientExecutor struct{}

// Execute runs smbclient with the given arguments
func (e *DefaultSmbClientExecutor) Execute(args []string) (string, error) {
	cmd := exec.Command("/bin/smbclient", args...)

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Combine stdout and stderr for complete output
	output := stdout.String() + stderr.String()

	if err != nil {
		return output, fmt.Errorf("smbclient command failed: %w (output: %s)", err, output)
	}

	return output, nil
}

// Global executor that can be replaced in tests
var smbClientExec ClientExecutor = &DefaultSmbClientExecutor{}

// buildSmbClientArgs constructs the arguments for smbclient command
func buildSmbClientArgs(cfg *config.SMBConfig, command string) ([]string, error) {
	args := []string{}

	// Build service name: //server/share
	server := cfg.ServerIP
	if server == "" {
		server = cfg.ServerName
	}
	service := fmt.Sprintf("//%s/%s", server, cfg.ShareName)
	args = append(args, service)

	// Add IP address if specified
	if cfg.ServerIP != "" && cfg.ServerName != "" {
		args = append(args, "-I", cfg.ServerIP)
	}

	// Add port if not default
	if cfg.Port != 445 {
		args = append(args, "-p", fmt.Sprintf("%d", cfg.Port))
	}

	// Add domain/workgroup if specified
	if cfg.Domain != "" {
		args = append(args, "-W", cfg.Domain)
	}

	// Handle authentication based on protocol
	switch strings.ToLower(cfg.AuthProtocol) {
	case "kerberos":
		args = append(args, "--use-kerberos=required")
		// For Kerberos, username/password are optional (uses system ticket cache)
		if cfg.Username != "" {
			args = append(args, "-U", cfg.Username)
		}
	case "ntlm", "negotiate", "":
		// For NTLM and Negotiate, we need username and password
		if cfg.Username == "" || cfg.Password == "" {
			return nil, fmt.Errorf("username and password are required for %s authentication", cfg.AuthProtocol)
		}
		// Use authentication format: username%password
		authStr := fmt.Sprintf("%s%%%s", cfg.Username, cfg.Password)
		args = append(args, "-U", authStr)
	default:
		return nil, fmt.Errorf("unsupported authentication protocol: %s", cfg.AuthProtocol)
	}

	// Disable prompts
	args = append(args, "-N")

	// Add the command to execute
	if command != "" {
		args = append(args, "-c", command)
	}

	return args, nil
}

// testConnection tests the connection to the SMB share
func testConnection(cfg *config.SMBConfig) error {
	args, err := buildSmbClientArgs(cfg, "ls")
	if err != nil {
		return err
	}

	output, err := smbClientExec.Execute(args)
	if err != nil {
		// Parse error message to provide more context
		if strings.Contains(output, "NT_STATUS_BAD_NETWORK_NAME") {
			return fmt.Errorf("share not found: %s", cfg.ShareName)
		}
		if strings.Contains(output, "NT_STATUS_LOGON_FAILURE") {
			return fmt.Errorf("authentication failed: invalid credentials")
		}
		if strings.Contains(output, "NT_STATUS_ACCESS_DENIED") {
			return fmt.Errorf("access denied to share: %s", cfg.ShareName)
		}
		if strings.Contains(output, "Connection refused") || strings.Contains(output, "failed to connect") {
			return fmt.Errorf("failed to connect to SMB server: connection refused")
		}
		return err
	}

	return nil
}

// uploadFileViaSmbClient uploads a file using smbclient
func uploadFileViaSmbClient(localPath string, remotePath string, cfg *config.SMBConfig) error {
	// Normalize remote path - remove leading slash
	remotePath = strings.TrimPrefix(remotePath, "/")
	remotePath = strings.TrimPrefix(remotePath, "\\")

	// Convert backslashes to forward slashes for consistency
	remotePath = strings.ReplaceAll(remotePath, "\\", "/")

	// Check if local file exists
	if _, err := os.Stat(localPath); os.IsNotExist(err) {
		return fmt.Errorf("local file not found: %s", localPath)
	}

	// Ensure parent directories exist by creating them first
	remoteDir := filepath.Dir(remotePath)
	if remoteDir != "." && remoteDir != "" {
		// Create directory command
		mkdirCmd := fmt.Sprintf("mkdir \"%s\"", remoteDir)
		args, err := buildSmbClientArgs(cfg, mkdirCmd)
		if err != nil {
			return err
		}
		// Try to create the parent directory, ignoring errors as it might already exist
		_, _ = smbClientExec.Execute(args) // nolint:errcheck
	}

	// Build the put command
	// Change to the directory containing the file first, then use relative path
	localDir := filepath.Dir(localPath)
	localFile := filepath.Base(localPath)

	// Build command: lcd <localdir>; put <localfile> <remotepath>
	command := fmt.Sprintf("lcd \"%s\"; put \"%s\" \"%s\"", localDir, localFile, remotePath)

	args, err := buildSmbClientArgs(cfg, command)
	if err != nil {
		return err
	}

	output, err := smbClientExec.Execute(args)
	if err != nil {
		// Parse error messages
		if strings.Contains(output, "NT_STATUS_OBJECT_NAME_COLLISION") {
			return fmt.Errorf("remote file already exists: %s", remotePath)
		}
		if strings.Contains(output, "NT_STATUS_ACCESS_DENIED") {
			return fmt.Errorf("access denied: cannot write to %s", remotePath)
		}
		if strings.Contains(output, "NT_STATUS_OBJECT_PATH_NOT_FOUND") {
			return fmt.Errorf("remote path not found: %s", filepath.Dir(remotePath))
		}
		return fmt.Errorf("failed to upload file: %w", err)
	}

	// Check if the output indicates success
	if !strings.Contains(output, "putting file") && !strings.Contains(output, "put") {
		return fmt.Errorf("upload may have failed: unexpected output")
	}

	return nil
}
