package smb

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/bancey/document-smbrelay-service/internal/config"
	"github.com/bancey/document-smbrelay-service/internal/logger"
)

// ClientExecutor defines the interface for executing smbclient commands
// This allows for easy mocking in tests
type ClientExecutor interface {
	Execute(args []string) (string, error)
}

// DefaultSmbClientExecutor uses the real smbclient binary
type DefaultSmbClientExecutor struct {
	BinaryPath string
}

// getSmbClientPath determines the path to the smbclient binary
// It checks the SMBCLIENT_PATH environment variable first, then searches common locations
func getSmbClientPath() string {
	// Check environment variable first
	if path := os.Getenv("SMBCLIENT_PATH"); path != "" {
		// Validate the path exists and is executable
		if validateBinaryPath(path) {
			return path
		}
	}

	// Try to find smbclient in PATH
	if path, err := exec.LookPath("smbclient"); err == nil {
		return path
	}

	// Common locations as fallbacks
	commonPaths := []string{
		"/usr/bin/smbclient",
		"/bin/smbclient",
		"/usr/local/bin/smbclient",
	}

	for _, path := range commonPaths {
		if validateBinaryPath(path) {
			return path
		}
	}

	// Default fallback
	return "/usr/bin/smbclient"
}

// validateBinaryPath checks if a path exists and is executable
func validateBinaryPath(path string) bool {
	// Check if file exists
	info, err := os.Stat(path)
	if err != nil {
		return false
	}

	// Check if it's a regular file (not a directory)
	if !info.Mode().IsRegular() {
		return false
	}

	// Check if it's executable (Unix permission check)
	if info.Mode().Perm()&0111 == 0 {
		return false
	}

	return true
}

// sanitizeArgsForLogging replaces sensitive data in args for safe logging
func sanitizeArgsForLogging(args []string, env map[string]string) ([]string, map[string]string) {
	sanitized := make([]string, len(args))
	copy(sanitized, args)

	// Sanitize password in -U flag if present (format: -U username or -U username%password)
	for i := 0; i < len(sanitized); i++ {
		if sanitized[i] == "-U" && i+1 < len(sanitized) {
			// Check if the next arg contains a password (has % character)
			if strings.Contains(sanitized[i+1], "%") {
				parts := strings.SplitN(sanitized[i+1], "%", 2)
				sanitized[i+1] = parts[0] + "%***"
			}
		}
	}

	// Sanitize environment variables
	sanitizedEnv := make(map[string]string)
	for k, v := range env {
		if k == "PASSWD" || k == "PASSWORD" || strings.Contains(strings.ToUpper(k), "PASS") {
			sanitizedEnv[k] = "***"
		} else {
			sanitizedEnv[k] = v
		}
	}

	return sanitized, sanitizedEnv
}

// Execute runs smbclient with the given arguments
func (e *DefaultSmbClientExecutor) Execute(args []string) (string, error) {
	return e.ExecuteWithEnv(args, nil)
}

// ExecuteWithEnv runs smbclient with the given arguments and environment variables
func (e *DefaultSmbClientExecutor) ExecuteWithEnv(args []string, env map[string]string) (string, error) {
	return e.ExecuteWithEnvAndLogging(args, env, false)
}

// ExecuteWithEnvAndLogging runs smbclient with the given arguments,
// environment variables, and optional logging
func (e *DefaultSmbClientExecutor) ExecuteWithEnvAndLogging(
	args []string, env map[string]string, enableLogging bool,
) (string, error) {
	binaryPath := e.BinaryPath
	if binaryPath == "" {
		binaryPath = getSmbClientPath()
	}

	// Log command if enabled
	if enableLogging {
		sanitizedArgs, sanitizedEnv := sanitizeArgsForLogging(args, env)
		logger.Info(fmt.Sprintf("Executing smbclient: %s %s", binaryPath, strings.Join(sanitizedArgs, " ")))
		if len(sanitizedEnv) > 0 {
			logger.Debug(fmt.Sprintf("Environment variables: %v", sanitizedEnv))
		}
	}

	// #nosec G204 - binaryPath is validated and comes from trusted sources:
	// 1. Environment variable (SMBCLIENT_PATH) - user is responsible for ensuring input is properly
	//    sanitised and do not contain unsafe user-controlled data.
	// 2. System PATH via exec.LookPath()
	// 3. Hardcoded known paths checked with validateBinaryPath()
	cmd := exec.Command(binaryPath, args...)

	// Set environment variables if provided
	if len(env) > 0 {
		// Start with the current environment
		cmd.Env = os.Environ()
		// Add custom environment variables
		for key, value := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}

	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	err := cmd.Run()

	// Combine stdout and stderr for complete output
	output := stdout.String() + stderr.String()

	// Log output if enabled
	if enableLogging {
		if err != nil {
			logger.Error(fmt.Sprintf("smbclient failed with error: %v", err))
			if output != "" {
				// Log output at ERROR level so it's always visible when debugging
				logger.Error(fmt.Sprintf("smbclient output: %s", output))
			}
		} else {
			logger.Debug(fmt.Sprintf("smbclient succeeded. Output: %s", output))
		}
	}

	if err != nil {
		return output, fmt.Errorf("smbclient command failed: %w (output: %s)", err, output)
	}

	return output, nil
}

// Global executor that can be replaced in tests
var smbClientExec ClientExecutor = &DefaultSmbClientExecutor{}

// buildSmbClientArgs constructs the arguments for smbclient command
// Returns args and environment variables map
func buildSmbClientArgs(cfg *config.SMBConfig, command string) ([]string, map[string]string, error) {
	args := []string{}
	env := make(map[string]string)

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
		// Kerberos uses -N flag to avoid password prompt
		args = append(args, "-N")
	case "ntlm", "negotiate", "":
		// For NTLM and Negotiate, we need username and password
		if cfg.Username == "" || cfg.Password == "" {
			return nil, nil, fmt.Errorf("username and password are required for %s authentication", cfg.AuthProtocol)
		}
		// Pass username via -U flag (without password for security)
		args = append(args, "-U", cfg.Username)
		// Pass password via PASSWD environment variable to handle special characters
		// This is more secure and avoids issues with special characters like %, $, etc.
		env["PASSWD"] = cfg.Password
		// Do NOT use -N flag here - smbclient will read password from PASSWD env var
		// The -N flag means "no password" which would prevent reading from PASSWD
	default:
		return nil, nil, fmt.Errorf("unsupported authentication protocol: %s", cfg.AuthProtocol)
	}

	// Add the command to execute
	if command != "" {
		args = append(args, "-c", command)
	}

	return args, env, nil
}

// testConnection tests the connection to the SMB share
func testConnection(cfg *config.SMBConfig) error {
	args, env, err := buildSmbClientArgs(cfg, "ls")
	if err != nil {
		return err
	}

	// Use ExecuteWithEnvAndLogging with logging flag from config
	var output string
	if executor, ok := smbClientExec.(*DefaultSmbClientExecutor); ok {
		output, err = executor.ExecuteWithEnvAndLogging(args, env, cfg.LogSmbCommands)
	} else {
		// For mock executors in tests
		output, err = smbClientExec.Execute(args)
	}
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
		if strings.Contains(output, "NT_STATUS_INVALID_PARAMETER") {
			return fmt.Errorf("invalid authentication parameters (check username/password format and special characters)")
		}
		if strings.Contains(output, "Connection refused") || strings.Contains(output, "failed to connect") {
			return fmt.Errorf("failed to connect to SMB server: connection refused")
		}
		return err
	}

	return nil
}

// testBasePath validates that the configured base path exists on the SMB share
func testBasePath(cfg *config.SMBConfig) error {
	// Normalize base path
	basePath := strings.Trim(cfg.BasePath, "/\\")
	basePath = strings.ReplaceAll(basePath, "\\", "/")

	if basePath == "" || basePath == "." {
		return nil // No base path to validate
	}

	// Try to change to the base path directory - this validates it exists and is accessible
	// Using 'cd' works correctly for nested directories like "apps/myapp"
	cmd := fmt.Sprintf("cd \"%s\"", basePath)
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
			return fmt.Errorf("base path does not exist: %s", cfg.BasePath)
		}
		if strings.Contains(output, "NT_STATUS_ACCESS_DENIED") {
			return fmt.Errorf("access denied to base path: %s", cfg.BasePath)
		}
		return fmt.Errorf("failed to access base path: %w", err)
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
		args, env, err := buildSmbClientArgs(cfg, mkdirCmd)
		if err != nil {
			return err
		}
		// Try to create the parent directory, ignoring errors as it might already exist
		if executor, ok := smbClientExec.(*DefaultSmbClientExecutor); ok {
			_, _ = executor.ExecuteWithEnvAndLogging(args, env, cfg.LogSmbCommands) // nolint:errcheck
		} else {
			_, _ = smbClientExec.Execute(args) // nolint:errcheck
		}
	}

	// Build the put command
	// Change to the directory containing the file first, then use relative path
	localDir := filepath.Dir(localPath)
	localFile := filepath.Base(localPath)

	// Build command: lcd <localdir>; put <localfile> <remotepath>
	command := fmt.Sprintf("lcd \"%s\"; put \"%s\" \"%s\"", localDir, localFile, remotePath)

	args, env, err := buildSmbClientArgs(cfg, command)
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
