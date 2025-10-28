package smb

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/bancey/document-smbrelay-service/internal/config"
)

// TestExecute tests the Execute wrapper function
func TestExecute(t *testing.T) {
	executor := &DefaultSmbClientExecutor{}

	// Test with invalid command to trigger error
	_, err := executor.Execute([]string{"--invalid-flag-test"})
	if err == nil {
		t.Error("Expected error for invalid command")
	}
}

// TestExecuteWithEnv tests the ExecuteWithEnv wrapper function
func TestExecuteWithEnv(t *testing.T) {
	executor := &DefaultSmbClientExecutor{}

	env := map[string]string{
		"TEST_VAR": "test_value",
	}

	// Test with invalid command to trigger error
	_, err := executor.ExecuteWithEnv([]string{"--invalid-flag-test"}, env)
	if err == nil {
		t.Error("Expected error for invalid command")
	}
}

// TestExecuteWithEnvAndLogging_WithLogging tests logging enabled path
func TestExecuteWithEnvAndLogging_WithLogging(t *testing.T) {
	executor := &DefaultSmbClientExecutor{}

	env := map[string]string{
		"PASSWD": "secret123",
	}

	// Test with invalid command and logging enabled
	_, err := executor.ExecuteWithEnvAndLogging([]string{"--invalid-flag-test"}, env, true)
	if err == nil {
		t.Error("Expected error for invalid command")
	}

	// Test with logging disabled
	_, err = executor.ExecuteWithEnvAndLogging([]string{"--invalid-flag-test"}, env, false)
	if err == nil {
		t.Error("Expected error for invalid command")
	}
}

// TestExecuteWithEnvAndLogging_NilEnv tests with nil environment
func TestExecuteWithEnvAndLogging_NilEnv(t *testing.T) {
	executor := &DefaultSmbClientExecutor{}

	// Test with nil env and logging enabled
	_, err := executor.ExecuteWithEnvAndLogging([]string{"--invalid-flag-test"}, nil, true)
	if err == nil {
		t.Error("Expected error for invalid command")
	}
}

// TestTestBasePath_EmptyPath tests testBasePath with empty path
func TestTestBasePath_EmptyPath(t *testing.T) {
	// Save and restore executor
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock that succeeds
	mock := NewMockExecutor()
	mock.ExecuteFunc = func(_ []string) (string, error) {
		return "success", nil
	}
	smbClientExec = mock

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		BasePath:     "", // Empty base path
		AuthProtocol: "ntlm",
	}

	err := testBasePath(cfg)
	if err != nil {
		t.Errorf("Expected no error for empty base path, got: %v", err)
	}
}

// TestTestBasePath_DotPath tests testBasePath with "." path
func TestTestBasePath_DotPath(t *testing.T) {
	// Save and restore executor
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock that succeeds
	mock := NewMockExecutor()
	mock.ExecuteFunc = func(_ []string) (string, error) {
		return "success", nil
	}
	smbClientExec = mock

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		BasePath:     ".", // Dot path
		AuthProtocol: "ntlm",
	}

	err := testBasePath(cfg)
	if err != nil {
		t.Errorf("Expected no error for dot base path, got: %v", err)
	}
}

// TestTestBasePath_ObjectNameNotFound tests testBasePath with path not found error
func TestTestBasePath_ObjectNameNotFound(t *testing.T) {
	// Save and restore executor
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock that returns path not found error
	mock := NewMockExecutor()
	mock.ExecuteFunc = func(_ []string) (string, error) {
		return "NT_STATUS_OBJECT_NAME_NOT_FOUND", fmt.Errorf("path not found")
	}
	smbClientExec = mock

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		BasePath:     "nonexistent/path",
		AuthProtocol: "ntlm",
	}

	err := testBasePath(cfg)
	if err == nil {
		t.Error("Expected error for non-existent base path")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected 'does not exist' error, got: %v", err)
	}
}

// TestTestBasePath_ObjectPathNotFound tests testBasePath with object path not found
func TestTestBasePath_ObjectPathNotFound(t *testing.T) {
	// Save and restore executor
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock that returns object path not found error
	mock := NewMockExecutor()
	mock.ExecuteFunc = func(_ []string) (string, error) {
		return "NT_STATUS_OBJECT_PATH_NOT_FOUND", fmt.Errorf("path not found")
	}
	smbClientExec = mock

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		BasePath:     "bad/path",
		AuthProtocol: "ntlm",
	}

	err := testBasePath(cfg)
	if err == nil {
		t.Error("Expected error for object path not found")
	}
	if !strings.Contains(err.Error(), "does not exist") {
		t.Errorf("Expected 'does not exist' error, got: %v", err)
	}
}

// TestTestBasePath_AccessDenied tests testBasePath with access denied error
func TestTestBasePath_AccessDenied(t *testing.T) {
	// Save and restore executor
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock that returns access denied error
	mock := NewMockExecutor()
	mock.ExecuteFunc = func(_ []string) (string, error) {
		return "NT_STATUS_ACCESS_DENIED", fmt.Errorf("access denied")
	}
	smbClientExec = mock

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		BasePath:     "restricted/path",
		AuthProtocol: "ntlm",
	}

	err := testBasePath(cfg)
	if err == nil {
		t.Error("Expected error for access denied")
	}
	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("Expected 'access denied' error, got: %v", err)
	}
}

// TestTestBasePath_GenericError tests testBasePath with generic error
func TestTestBasePath_GenericError(t *testing.T) {
	// Save and restore executor
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock that returns generic error
	mock := NewMockExecutor()
	mock.ExecuteFunc = func(_ []string) (string, error) {
		return "Generic error message", fmt.Errorf("some error")
	}
	smbClientExec = mock

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		BasePath:     "some/path",
		AuthProtocol: "ntlm",
	}

	err := testBasePath(cfg)
	if err == nil {
		t.Error("Expected error for generic failure")
	}
	if !strings.Contains(err.Error(), "failed to access base path") {
		t.Errorf("Expected 'failed to access base path' error, got: %v", err)
	}
}

// TestTestConnection_BadNetworkName tests connection with bad network name error
func TestTestConnection_BadNetworkName(t *testing.T) {
	// Save and restore executor
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock that returns bad network name error
	mock := NewMockExecutor()
	mock.ExecuteFunc = func(_ []string) (string, error) {
		return "NT_STATUS_BAD_NETWORK_NAME", fmt.Errorf("bad network name")
	}
	smbClientExec = mock

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "nonexistent",
		Username:     "testuser",
		Password:     "testpass",
		AuthProtocol: "ntlm",
	}

	err := testConnection(cfg)
	if err == nil {
		t.Error("Expected error for bad network name")
	}
	if !strings.Contains(err.Error(), "share not found") {
		t.Errorf("Expected 'share not found' error, got: %v", err)
	}
}

// TestTestConnection_InvalidParameter tests connection with invalid parameter error
func TestTestConnection_InvalidParameter(t *testing.T) {
	// Save and restore executor
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock that returns invalid parameter error
	mock := NewMockExecutor()
	mock.ExecuteFunc = func(_ []string) (string, error) {
		return "NT_STATUS_INVALID_PARAMETER", fmt.Errorf("invalid parameter")
	}
	smbClientExec = mock

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		AuthProtocol: "ntlm",
	}

	err := testConnection(cfg)
	if err == nil {
		t.Error("Expected error for invalid parameter")
	}
	if !strings.Contains(err.Error(), "invalid authentication parameters") {
		t.Errorf("Expected 'invalid authentication parameters' error, got: %v", err)
	}
}

// TestTestConnection_ConnectionRefused tests connection with connection refused error
func TestTestConnection_ConnectionRefused(t *testing.T) {
	// Save and restore executor
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock that returns connection refused error
	mock := NewMockExecutor()
	mock.ExecuteFunc = func(_ []string) (string, error) {
		return "Connection refused", fmt.Errorf("connection refused")
	}
	smbClientExec = mock

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		AuthProtocol: "ntlm",
	}

	err := testConnection(cfg)
	if err == nil {
		t.Error("Expected error for connection refused")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("Expected 'connection refused' error, got: %v", err)
	}
}

// TestTestConnection_FailedToConnect tests connection with failed to connect error
func TestTestConnection_FailedToConnect(t *testing.T) {
	// Save and restore executor
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Setup mock that returns failed to connect error
	mock := NewMockExecutor()
	mock.ExecuteFunc = func(_ []string) (string, error) {
		return "failed to connect", fmt.Errorf("failed to connect")
	}
	smbClientExec = mock

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		AuthProtocol: "ntlm",
	}

	err := testConnection(cfg)
	if err == nil {
		t.Error("Expected error for failed to connect")
	}
	if !strings.Contains(err.Error(), "connection refused") {
		t.Errorf("Expected 'connection refused' error, got: %v", err)
	}
}

// TestUploadFileViaSmbClient_ObjectNameCollision tests upload with file already exists error
func TestUploadFileViaSmbClient_ObjectNameCollision(t *testing.T) {
	// Save and restore executor
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Create a temp file for upload
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-collision.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Setup mock that returns object name collision error
	mock := NewMockExecutor()
	mock.ExecuteFunc = func(_ []string) (string, error) {
		return "NT_STATUS_OBJECT_NAME_COLLISION", fmt.Errorf("file exists")
	}
	smbClientExec = mock

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		AuthProtocol: "ntlm",
	}

	err = uploadFileViaSmbClient(tmpFile, "existing/file.txt", cfg)
	if err == nil {
		t.Error("Expected error for file already exists")
	}
	if !strings.Contains(err.Error(), "already exists") {
		t.Errorf("Expected 'already exists' error, got: %v", err)
	}
}

// TestUploadFileViaSmbClient_AccessDenied tests upload with access denied error
func TestUploadFileViaSmbClient_AccessDenied(t *testing.T) {
	// Save and restore executor
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Create a temp file for upload
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-denied.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Setup mock that returns access denied error
	mock := NewMockExecutor()
	mock.ExecuteFunc = func(_ []string) (string, error) {
		return "NT_STATUS_ACCESS_DENIED", fmt.Errorf("access denied")
	}
	smbClientExec = mock

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		AuthProtocol: "ntlm",
	}

	err = uploadFileViaSmbClient(tmpFile, "restricted/file.txt", cfg)
	if err == nil {
		t.Error("Expected error for access denied")
	}
	if !strings.Contains(err.Error(), "access denied") {
		t.Errorf("Expected 'access denied' error, got: %v", err)
	}
}

// TestUploadFileViaSmbClient_ObjectPathNotFound tests upload with path not found error
func TestUploadFileViaSmbClient_ObjectPathNotFound(t *testing.T) {
	// Save and restore executor
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Create a temp file for upload
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-notfound.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Setup mock that returns object path not found error
	mock := NewMockExecutor()
	mock.ExecuteFunc = func(_ []string) (string, error) {
		return "NT_STATUS_OBJECT_PATH_NOT_FOUND", fmt.Errorf("path not found")
	}
	smbClientExec = mock

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		AuthProtocol: "ntlm",
	}

	err = uploadFileViaSmbClient(tmpFile, "nonexistent/dir/file.txt", cfg)
	if err == nil {
		t.Error("Expected error for path not found")
	}
	if !strings.Contains(err.Error(), "path not found") {
		t.Errorf("Expected 'path not found' error, got: %v", err)
	}
}

// TestUploadFileViaSmbClient_UnexpectedOutput tests upload with unexpected output
func TestUploadFileViaSmbClient_UnexpectedOutput(t *testing.T) {
	// Save and restore executor
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	// Create a temp file for upload
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "test-unexpected.txt")
	err := os.WriteFile(tmpFile, []byte("test content"), 0644)
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Setup mock that returns unexpected output (no "putting file" or "put")
	mock := NewMockExecutor()
	mock.ExecuteFunc = func(args []string) (string, error) {
		// Check what command is in the args
		for i, arg := range args {
			if arg == "-c" && i+1 < len(args) {
				cmd := args[i+1]

				// If it's a mkdir command, return success
				if strings.Contains(cmd, "mkdir") {
					return "", nil
				}

				// If it's a put command, return message without expected keywords
				// Note: Avoid words like "output" which contain "put" as substring
				if strings.Contains(cmd, "lcd") {
					return "Some unexpected result from smbclient without expected keywords", nil
				}
			}
		}

		// Default: return empty
		return "", nil
	}
	smbClientExec = mock

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		AuthProtocol: "ntlm",
		MaxRetries:   0, // Disable retries to make test predictable
	}

	// Use simple filename (no path) to avoid mkdir call
	err = uploadFileViaSmbClient(tmpFile, "file.txt", cfg)

	if err == nil {
		t.Fatal("Expected error for unexpected message, got nil")
	}
	if !strings.Contains(err.Error(), "unexpected") {
		t.Errorf("Expected 'unexpected' error, got: %v", err)
	}
}

// TestUploadFileViaSmbClient_FileNotFound tests upload with local file not found
func TestUploadFileViaSmbClient_FileNotFound(t *testing.T) {
	// Save and restore executor
	origExec := smbClientExec
	defer func() { smbClientExec = origExec }()

	cfg := &config.SMBConfig{
		ServerName:   "testserver",
		ServerIP:     "127.0.0.1",
		ShareName:    "testshare",
		Username:     "testuser",
		Password:     "testpass",
		AuthProtocol: "ntlm",
	}

	err := uploadFileViaSmbClient("/nonexistent/file.txt", "test/file.txt", cfg)
	if err == nil {
		t.Error("Expected error for non-existent local file")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' error, got: %v", err)
	}
}

// TestGetSmbClientPath_UnexecutableFile tests getSmbClientPath with non-executable file
func TestGetSmbClientPath_UnexecutableFile(t *testing.T) {
	// Save original env and restore after test
	origPath := os.Getenv("SMBCLIENT_PATH")
	defer func() {
		if origPath != "" {
			os.Setenv("SMBCLIENT_PATH", origPath)
		} else {
			os.Unsetenv("SMBCLIENT_PATH")
		}
	}()

	// Create a temporary non-executable file
	tmpDir := os.TempDir()
	tmpFile := filepath.Join(tmpDir, "non-executable-test")
	err := os.WriteFile(tmpFile, []byte("test"), 0644) // No execute permission
	if err != nil {
		t.Fatalf("Failed to create test file: %v", err)
	}
	defer os.Remove(tmpFile)

	// Set SMBCLIENT_PATH to the non-executable file
	os.Setenv("SMBCLIENT_PATH", tmpFile)

	result := getSmbClientPath()

	// Should not return the non-executable file
	if result == tmpFile {
		t.Error("Expected fallback for non-executable file")
	}
}

// TestValidateBinaryPath_NonRegularFile tests validateBinaryPath edge cases
func TestValidateBinaryPath_NonRegularFile(t *testing.T) {
	// Test with empty string
	result := validateBinaryPath("")
	if result {
		t.Error("Expected validation to fail for empty path")
	}
}
